package live

import (
	"log"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
)

func NewMicrophone() (*StreamedProcess, error) {
	return ffmpegMicCapture()
}

func PublishMicrophone(room *lksdk.Room, proc *StreamedProcess) error {
	return proc.Publish(
		room,
		webrtc.MimeTypeOpus,
		20*time.Millisecond,
		func() { log.Println("microphone streaming ended") },
		&lksdk.TrackPublicationOptions{
			Name:   "mic",
			Source: livekit.TrackSource_MICROPHONE,
		},
	)
}
