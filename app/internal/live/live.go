package live

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/livekit/protocol/auth"
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

	microphone *StreamedProcess
	camera     *StreamedProcess
	screen     *StreamedProcess
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
		serverCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		l.decoderServer.Shutdown(serverCtx)
		l.decoderServer = nil
	}

	l.UnpublishMic()

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

	microphone, err := NewMicrophone()
	if err != nil {
		return err
	}

	err = microphone.Start()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			return
		}
		log.Println("failed to publish microphone, cleaning up mic stream")
		_ = microphone.Close()
	}()

	err = PublishMicrophone(l.room, microphone)
	if err != nil {
		return err
	}

	l.microphone = microphone

	return nil
}

func (l *LiveChat) UnpublishMic() {
	if l.microphone == nil {
		return
	}

	sid := l.microphone.SID()
	if sid != "" {
		l.room.LocalParticipant.UnpublishTrack(sid)
	}
	_ = l.microphone.Close()
	l.microphone = nil
}

func (l *LiveChat) UnpublishWebcam() {
	if l.camera == nil {
		return
	}

	sid := l.camera.SID()
	if sid != "" {
		l.room.LocalParticipant.UnpublishTrack(sid)
	}
	_ = l.camera.Close()
	l.camera = nil
}

func (l *LiveChat) PublishWebcam() error {
	if !l.Connected() {
		return fmt.Errorf("cannot publish cam: not connected to a room")
	}

	if l.camera != nil {
		return fmt.Errorf("cannot publish cam: cam already on")
	}

	cam, err := NewWebcam()
	if err != nil {
		return err
	}

	err = cam.Start()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			return
		}
		log.Println("failed to publish webcam, cleaning up webcam stream")
		_ = cam.Close()
	}()

	err = PublishWebcam(l.room, cam)
	if err != nil {
		return err
	}

	l.camera = cam

	return nil
}

func (l *LiveChat) SetMicMuted(muted bool) {
	l.microphone.SetMuted(muted)
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

func (l *LiveChat) GetWindows() ([]WindowData, error) {
	return GetWindows()
}

func (l *LiveChat) Thumbnail(w WindowData) ([]byte, error) {
	return w.Thumbnail()
}

func (l *LiveChat) UnpublishScreenShare() {
	if l.screen == nil {
		return
	}

	sid := l.screen.SID()
	if sid != "" {
		l.room.LocalParticipant.UnpublishTrack(sid)
	}
	_ = l.screen.Close()
	l.screen = nil
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

	screenShare, err := NewScreenShare(w)
	if err != nil {
		return err
	}

	err = screenShare.Start()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			return
		}
		log.Println("failed to publish screenshare, cleaning up screenshare stream")
		_ = screenShare.Close()
	}()

	err = PublishScreenShare(l.room, screenShare, func() {
		// screenShare.Close()
		application.Get().Event.Emit(EventScreenShareClosed)
	})
	if err != nil {
		return err
	}

	l.screen = screenShare

	return nil
}
