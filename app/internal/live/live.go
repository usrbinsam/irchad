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

	microphone    io.Closer
	microphonePub *lksdk.LocalTrackPublication
	camera        io.Closer
	cameraPub     *lksdk.LocalTrackPublication
	screen        io.Closer
	screenPub     *lksdk.LocalTrackPublication
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
				if track := remotePub.Track(); track != nil {
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

	pub, err := l.room.LocalParticipant.PublishTrack(
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
	l.microphonePub = pub

	return nil
}

func (l *LiveChat) UnpublishMicrophone() {
	if l.microphone == nil {
		return
	}

	err := l.microphone.Close()
	if err != nil {
		log.Printf("failed to stop microphone: %s\n", err.Error())
		return
	}

	err = l.room.LocalParticipant.UnpublishTrack(l.microphonePub.SID())
	if err != nil {
		log.Printf("failed to unpublish microphone track: %s\n", err.Error())
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

	if err := l.room.LocalParticipant.UnpublishTrack(l.cameraPub.SID()); err != nil {
		log.Printf("failed to unpublish camera track: %s", err.Error())
	}

	l.cameraPub = nil
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

	pub, err := l.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
		Name:   "cam",
		Source: livekit.TrackSource_CAMERA,
	})
	if err != nil {
		track.Close()
		return err
	}

	l.camera = cam
	l.cameraPub = pub
	return nil
}

func (l *LiveChat) SetMicMuted(muted bool) {
	l.microphonePub.SetMuted(muted)
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
			log.Printf("decodeVideoStream() error - %s\n", err.Error())
			return
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

func (l *LiveChat) GetWindows() ([]WindowData, error) {
	return GetWindows()
}

func (l *LiveChat) Thumbnail(w WindowData) ([]byte, error) {
	return w.Thumbnail()
}

func (l *LiveChat) UnpublishScreenShare() {
	log.Printf("Unpublish screen share requested")
	if l.screen == nil {
		return
	}
	err := l.screen.Close()
	if err != nil {
		log.Printf("failed to stop screen share: %s", err.Error())
		return
	}
	l.screen = nil
	err = l.room.LocalParticipant.UnpublishTrack(l.screenPub.SID())
	if err != nil {
		log.Printf("failed to unpublish screen share track: %s", err.Error())
		return
	}
	application.Get().Event.Emit(EventScreenShareClosed)
}

func (l *LiveChat) PublishScreenShare(ID uint32) error {
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

	track, err := lksdk.NewLocalSampleTrack(
		webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
		},
	)

	log.Printf("starting screen share for: %+v", w)
	screenShare, err := NewScreenShare(w, track)
	if err != nil {
		fmt.Errorf("GStreamer error: %s", err.Error())
		return err
	}

	_, err = l.room.LocalParticipant.PublishTrack(
		track,
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
	return nil
}
