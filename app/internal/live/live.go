package live

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/livekit/protocol/auth"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v4"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func NewLiveChat(url string) *LiveChat {
	return &LiveChat{
		url:      url,
		room:     nil,
		registry: NewStreamRegistry(),
	}
}

type LiveChat struct {
	url           string
	room          *lksdk.Room
	registry      *StreamRegistry
	decoderServer *httpServer // provides decoded video/audio streams when connected to a room
}

func codecSelector() *mediadevices.CodecSelector {
	x264Params, err := x264.NewParams()
	if err != nil {
		panic(err)
	}
	x264Params.BitRate = 500_000 // 500kbps

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}

	return mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
		mediadevices.WithAudioEncoders(&opusParams),
	)
}

func (l *LiveChat) getToken(identity, room string) (string, error) {
	at := auth.NewAccessToken("devkey", "secret")
	videoGrant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}

	sipGrant := &auth.SIPGrant{
		Admin: false,
		Call:  true,
	}

	at.SetSIPGrant(sipGrant).SetVideoGrant(videoGrant).SetIdentity(identity).SetValidFor(time.Hour)
	return at.ToJWT()
}

func (l *LiveChat) onParticipantConnected(rp *lksdk.RemoteParticipant) {
	app := application.Get()
	app.Event.Emit(
		EventParticipantConnected,
		ParticipantConnected{Identity: rp.Identity(), Channel: l.room.Name()},
	)

	log.Printf("available tracks:\n")
	for _, pub := range rp.TrackPublications() {
		log.Printf("%+v\n", pub)
	}
}

func (l *LiveChat) onParticipantDisconnected(rp *lksdk.RemoteParticipant) {
	app := application.Get()
	app.Event.Emit(
		EventParticipantDisconnected,
		ParticipantDisconnected{
			Identity: rp.Identity(),
			Channel:  rp.Name(),
		},
	)
}

func (l *LiveChat) Connect(channelName string) error {
	log.Printf("connecting to %s on %s", channelName, l.url)

	cb := &lksdk.RoomCallback{
		OnParticipantConnected:    l.onParticipantConnected,
		OnParticipantDisconnected: l.onParticipantDisconnected,
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: l.onTrackSubscribed,
		},
	}

	token, err := l.getToken("chad", channelName)
	if err != nil {
		log.Fatalf("error getting join token: %s", err.Error())
	}

	room, err := lksdk.ConnectToRoomWithToken(
		l.url,
		token,
		cb,
	)
	if err != nil {
		log.Printf("failed to connect to channel: %s", err.Error())
		return err
	}
	l.room = room
	log.Printf("connected to %s", channelName)

	l.startDecodeServer()

	app := application.Get()
	for _, rp := range l.room.GetRemoteParticipants() {
		app.Event.Emit(
			EventParticipantConnected,
			ParticipantConnected{
				Identity: rp.Identity(),
			},
		)
	}

	return nil
}

func (l *LiveChat) Disconnect(ctx context.Context) {
	if l.decoderServer != nil {
		serverCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		l.decoderServer.Shutdown(serverCtx)
		l.decoderServer = nil
	}
	if l.room != nil {
		l.room.Disconnect()
		l.room = nil
	}
}

func (l *LiveChat) Connected() bool {
	if l.room == nil {
		return false
	}

	return l.room.ConnectionState() == lksdk.ConnectionStateConnected
}

func (l *LiveChat) PublishWebcam() error {
	if !l.Connected() {
		return fmt.Errorf("cannot publish webcam: room not connected")
	}

	stream := getMediaDevices()
	videoTracks := stream.GetVideoTracks()

	if len(videoTracks) == 0 {
		return fmt.Errorf("cannot publish webcam: no webcam detected")
	}

	webcam := videoTracks[0]
	_, err := l.room.LocalParticipant.PublishTrack(
		webcam,
		&lksdk.TrackPublicationOptions{
			Name: "webcam",
		},
	)

	log.Printf("publlished webcam track: %+v", webcam)

	return err
}

func (l *LiveChat) PublishMicrophone() error {
	if !l.Connected() {
		return fmt.Errorf("cannot publish mic: not connected to a room")
	}

	ffmpegIn, err := ffmpegMicCapture()
	if err != nil {
		return err
	}

	track, err := lksdk.NewLocalReaderTrack(
		ffmpegIn,
		webrtc.MimeTypeOpus,
		lksdk.ReaderTrackWithFrameDuration(20*time.Millisecond),
		lksdk.ReaderTrackWithOnWriteComplete(func() { log.Println("microphone streaming ended") }),
	)
	if err != nil {
		return err
	}

	_, err = l.room.LocalParticipant.PublishTrack(
		track,
		&lksdk.TrackPublicationOptions{
			Name: "mic",
		},
	)

	return err
}

func (l *LiveChat) onTrackSubscribed(
	track *webrtc.TrackRemote,
	publication *lksdk.RemoteTrackPublication,
	rp *lksdk.RemoteParticipant,
) {
	identity := rp.Identity()
	trackID := track.ID()
	log.Printf("track subscribed: %s/%s\n", l.room.Name(), rp.Identity())

	kind := track.Kind()
	app := application.Get()
	switch kind {
	case webrtc.RTPCodecTypeVideo:
		_, err := l.decodeVideoStream(track, publication, rp)
		if err != nil {
			log.Fatalf("decodeVideoStream() error - %s", err.Error())
		}

	case webrtc.RTPCodecTypeAudio:
		l.decodeAudioStream(track, publication, rp)
	}

	ev := ParticipantTrackPublished{
		Identity: identity,
		TrackID:  trackID,
		Source:   publication.Source().String(),
		Kind:     track.Kind().String(),
		SubscribeURL: fmt.Sprintf(
			"http://%s/stream?pid=%s&tid=%s",
			l.decoderServer.Addr(),
			identity,
			trackID,
		),
	}

	log.Printf("Published: %+v", ev)
	app.Event.Emit(
		EventParticipantTrackPublished, ev,
	)
}

func getMediaDevices() mediadevices.MediaStream {
	stream, err := mediadevices.GetUserMedia(
		mediadevices.MediaStreamConstraints{
			Video: func(c *mediadevices.MediaTrackConstraints) {
				c.FrameFormat = prop.FrameFormat(frame.FormatI420)
				c.Width = prop.Int(640)
				c.Height = prop.Int(480)
			},
			Audio: func(c *mediadevices.MediaTrackConstraints) {},
			Codec: codecSelector(),
		},
	)
	if err != nil {
		panic(err)
	}

	return stream
}

func ffmpegMicCapture() (io.ReadCloser, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-f",
		"pulse",
		"-i",
		"default",
		"-c:a",
		"libopus",
		"-b:a",
		"64k",
		"-vbr",
		"on",
		"-f",
		"opus",
		"pipe:1",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
	}()

	return stdout, err
}
