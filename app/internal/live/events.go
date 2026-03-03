package live

import "github.com/wailsapp/wails/v3/pkg/application"

const (
	EventParticipantConnected        = "live:participant-connected"
	EventParticipantDisconnected     = "live:participant-disconnected"
	EventParticipantTrackPublished   = "live:participant-track-published"
	EventParticipantTrackUnpublished = "live:participant-track-unpublished"
	EventScreenShareClosed           = "live:screen-share-closed"
	EventSpeakingChanged             = "live:speaking-changed"
)

type ParticipantConnected struct {
	Identity string
	Channel  string
}

type ParticipantDisconnected struct {
	Identity string
	Channel  string
}

type ParticipantTrackPublished struct {
	Identity     string
	TrackID      string
	SubscribeURL string
	Source       string
	Kind         string
	Show         bool
	TrackName    string
}

type ParticipantTrackUnpublished struct {
	Identity string
	TrackID  string
}

type (
	ScreenShareClosed struct{}
	SpeakingChanged   struct {
		Identity   string
		IsSpeaking bool
	}
)

func RegisterEvents() {
	application.RegisterEvent[ParticipantConnected](EventParticipantConnected)
	application.RegisterEvent[ParticipantDisconnected](EventParticipantDisconnected)
	application.RegisterEvent[ParticipantTrackPublished](EventParticipantTrackPublished)
	application.RegisterEvent[ParticipantTrackUnpublished](EventParticipantTrackUnpublished)
	application.RegisterEvent[ScreenShareClosed](EventScreenShareClosed)
	application.RegisterEvent[SpeakingChanged](EventSpeakingChanged)
}
