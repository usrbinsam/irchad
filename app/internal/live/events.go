package live

import "github.com/wailsapp/wails/v3/pkg/application"

const (
	EventParticipantConnected      = "live:participant-connected"
	EventParticipantDisconnected   = "live:participant-dissconnected"
	EventParticipantTrackPublished = "live:participant-track-published"
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
}

func RegisterEvents() {
	application.RegisterEvent[ParticipantConnected](EventParticipantConnected)
	application.RegisterEvent[ParticipantDisconnected](EventParticipantDisconnected)
	application.RegisterEvent[ParticipantTrackPublished](EventParticipantTrackPublished)
}
