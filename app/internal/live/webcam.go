package live

import (
	"log"
	"os"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
)

func NewWebcam() (*StreamedProcess, error) {
	proc, err := ffmpegWebcamCapture()
	if err != nil {
		return nil, err
	}
	proc.cmd.Stderr = os.Stderr
	return proc, nil
}

func PublishWebcam(room *lksdk.Room, proc *StreamedProcess) error {
	return proc.Publish(
		room,
		webrtc.MimeTypeVP9,
		33*time.Millisecond,
		func() { log.Println("webcam streaming ended") },
		&lksdk.TrackPublicationOptions{
			Name:   "camera",
			Source: livekit.TrackSource_CAMERA,
		},
	)
}
