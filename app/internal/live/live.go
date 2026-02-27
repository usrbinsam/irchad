package live

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
	"github.com/tinyzimmer/go-gst/gst"
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

	microphone io.Closer
	camera     io.Closer
	screen     *gst.Pipeline
}

func (l *LiveChat) getToken(identity, room string) (string, error) {
	at := auth.NewAccessToken("devkey", "secret")
	videoGrant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	videoGrant.SetCanPublish(true)
	videoGrant.SetCanSubscribe(true)

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

	log.Printf("%s joined the call\n", rp.Identity())
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

func (l *LiveChat) Connect(nick string, channelName string) error {
	log.Printf("connecting to %s on %s", channelName, l.url)

	cb := &lksdk.RoomCallback{
		OnParticipantConnected:    l.onParticipantConnected,
		OnParticipantDisconnected: l.onParticipantDisconnected,
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: l.onTrackSubscribed,
			OnTrackUnpublished: func(publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
				app := application.Get()
				app.Event.Emit(
					EventParticipantTrackUnpublished,
					ParticipantTrackUnpublished{
						Identity: rp.Identity(),
						TrackID:  publication.Track().ID(),
					},
				)
			},
		},
	}

	token, err := l.getToken(nick, channelName)
	if err != nil {
		log.Fatalf("error getting join token: %s", err.Error())
	}

	room, err := lksdk.ConnectToRoomWithToken(
		l.url,
		token,
		cb,
		lksdk.WithAutoSubscribe(true),
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

		for _, pub := range rp.TrackPublications() {
			if remotePub, ok := pub.(*lksdk.RemoteTrackPublication); ok {
				if track := remotePub.TrackRemote(); track != nil {
					l.onTrackSubscribed(track, remotePub, rp)
				}
			}
		}
	}

	return nil
}

func (l *LiveChat) Disconnect(ctx context.Context) {
	if l.decoderServer != nil {
		serverCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		l.decoderServer.Shutdown(serverCtx)
		l.decoderServer = nil
	}

	l.UnpublishMicrophone()
	l.UnpublishScreenShare()
	l.UnpublishWebcam()

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

func (l *LiveChat) PublishMicrophone() error {
	if !l.Connected() {
		return fmt.Errorf("cannot publish mic: not connected to a room")
	}

	if l.microphone != nil {
		return fmt.Errorf("cannot publish mic: mic already on")
	}

	track, err := lksdk.NewLocalSampleTrack(
		webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeOpus,
		},
	)
	if err != nil {
		log.Printf("error creating microphone track: %s\n", err.Error())
		return err
	}

	microphone, err := NewMicrophone(track)
	if err != nil {
		log.Printf("error getting microphone stream: %s\n", err.Error())
		return err
	}

	_, err = l.room.LocalParticipant.PublishTrack(
		track,
		&lksdk.TrackPublicationOptions{
			Name:   "mic",
			Source: livekit.TrackSource_MICROPHONE,
		},
	)
	if err != nil {
		log.Printf("error publishing microphone track: %s\n", err.Error())
		_ = microphone.Close()
		return err
	}

	l.microphone = microphone

	return nil
}

func (l *LiveChat) UnpublishMicrophone() {
	if l.microphone == nil {
		return
	}

	pub := l.room.LocalParticipant.GetTrackPublication(livekit.TrackSource_MICROPHONE)
	err := l.room.LocalParticipant.UnpublishTrack(pub.SID())
	if err != nil {
		log.Printf("failed to unpublish microphone track: %s\n", err.Error())
		return
	}

	err = l.microphone.Close()
	if err != nil {
		log.Printf("failed to stop microphone: %s\n", err.Error())
		return
	}
	l.microphone = nil
}

func (l *LiveChat) UnpublishWebcam() {
	if l.camera == nil {
		return
	}

	if err := l.camera.Close(); err != nil {
		log.Printf("failed to close camera stream: %s", err.Error())
		return
	}

	pub := l.room.LocalParticipant.GetTrackPublication(livekit.TrackSource_CAMERA)
	if err := l.room.LocalParticipant.UnpublishTrack(pub.SID()); err != nil {
		log.Printf("failed to unpublish camera track: %s", err.Error())
	}

	l.camera = nil
}

func (l *LiveChat) PublishWebcam() error {
	if !l.Connected() {
		return fmt.Errorf("cannot publish cam: not connected to a room")
	}

	if l.camera != nil {
		return fmt.Errorf("cannot publish cam: cam already on")
	}

	track, err := lksdk.NewLocalSampleTrack(
		webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
		},
	)
	if err != nil {
		return err
	}

	cam, err := NewWebcam(track)
	if err != nil {
		return err
	}

	_, err = l.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
		Name:   "cam",
		Source: livekit.TrackSource_CAMERA,
	})
	if err != nil {
		track.Close()
		return err
	}

	l.camera = cam
	return nil
}

func (l *LiveChat) SetMicMuted(muted bool) {
	pub := l.room.LocalParticipant.GetTrackPublication(livekit.TrackSource_MICROPHONE)
	mic := pub.(*lksdk.LocalTrackPublication)
	mic.SetMuted(muted)
}

func (l *LiveChat) onTrackSubscribed(
	track *webrtc.TrackRemote,
	publication *lksdk.RemoteTrackPublication,
	rp *lksdk.RemoteParticipant,
) {
	identity := rp.Identity()
	trackID := track.ID()

	kind := track.Kind()
	app := application.Get()
	source := publication.Source()

	if source == livekit.TrackSource_SCREEN_SHARE || source == livekit.TrackSource_SCREEN_SHARE_AUDIO {
		l.decodeScreenShare(track, publication, rp)
		return
	}
	switch kind {
	case webrtc.RTPCodecTypeVideo:
		_, err := l.decodeVideoStream(track, publication, rp)
		if err != nil {
			log.Printf("decodeVideoStream() error - %s\n", err.Error())
			return
		}

	case webrtc.RTPCodecTypeAudio:
		l.decodeAudioStream(track, publication, rp)
		return
	}

	ev := ParticipantTrackPublished{
		Identity:  identity,
		TrackID:   trackID,
		Source:    publication.Source().String(),
		Kind:      track.Kind().String(),
		TrackName: publication.Name(),
		SubscribeURL: fmt.Sprintf(
			"http://%s/stream?pid=%s&tid=%s",
			l.decoderServer.Addr(),
			identity,
			trackID,
		),
	}

	log.Printf("new remote track published: %+v", ev)
	app.Event.Emit(
		EventParticipantTrackPublished, ev,
	)
}

func (l *LiveChat) GetWindows() ([]WindowData, error) {
	return GetWindows()
}

func (l *LiveChat) UnpublishScreenShare() {
	log.Printf("Unpublish screen share requested")
	if l.screen == nil {
		return
	}

	// screen share publication
	videoPub := l.room.LocalParticipant.GetTrackPublication(livekit.TrackSource_SCREEN_SHARE)
	log.Printf("video pub: %+v", videoPub)
	err := l.room.LocalParticipant.UnpublishTrack(videoPub.SID())
	if err != nil {
		log.Printf("failed to unpublish screen video track: %s", err.Error())
		return
	}

	// audio publication
	audioPub := l.room.LocalParticipant.GetTrackPublication(livekit.TrackSource_SCREEN_SHARE_AUDIO)
	err = l.room.LocalParticipant.UnpublishTrack(audioPub.SID())
	if err != nil {
		log.Printf("failed to unpublish screen audio track: %s", err.Error())
		return
	}
	// stop pipeline
	err = l.screen.SetState(gst.StateNull)
	if err != nil {
		log.Printf("failed to stop screen share: %s", err.Error())
		return
	}
	l.screen = nil
	application.Get().Event.Emit(EventScreenShareClosed, ScreenShareClosed{})
}

func (l *LiveChat) PublishScreenShare(ID uint32, ss ScreenShareOpts) error {
	if !l.Connected() {
		return fmt.Errorf("cannot publish screen: not connected to a room")
	}

	if l.screen != nil {
		return fmt.Errorf("cannot publish screen: screen already on")
	}

	windows, _ := GetWindows()
	var w *WindowData
	for _, win := range windows {
		if win.ID == ID {
			w = &win
			break
		}
	}

	if w == nil {
		return fmt.Errorf("invalid window ID")
	}

	videoTrack, err := lksdk.NewLocalSampleTrack(
		webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
		},
	)
	if err != nil {
		log.Printf("error creating video track: %s", err.Error())
		return err
	}

	audioTrack, err := lksdk.NewLocalSampleTrack(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
	)
	if err != nil {
		log.Printf("error creating audio track: %s", err.Error())
		return err
	}

	log.Printf("new screen share for %s - audio=%s video=%s", w.Title, audioTrack.ID(), videoTrack.ID())
	screenShare, err := NewScreenShare(w, &ss, audioTrack, videoTrack)
	if err != nil {
		log.Printf("GStreamer error: %s", err.Error())
		return err
	}

	_, err = l.room.LocalParticipant.PublishTrack(
		videoTrack,
		&lksdk.TrackPublicationOptions{
			Name:   w.Title,
			Source: livekit.TrackSource_SCREEN_SHARE,
		},
	)
	if err != nil {
		log.Printf("failed to publish screen share track: %s", err.Error())
		return fmt.Errorf("failed to publish track: %w", err)
	}

	l.screen = screenShare

	_, err = l.room.LocalParticipant.PublishTrack(
		audioTrack,
		&lksdk.TrackPublicationOptions{
			Name:   w.Title,
			Source: livekit.TrackSource_SCREEN_SHARE_AUDIO,
		},
	)
	if err != nil {
		log.Printf("failed to publish screen audio track: %s", err.Error())
		return err
	}

	// l.screenAudioPub = audioPub

	return nil
}
