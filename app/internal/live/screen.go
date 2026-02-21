package live

import (
	"os"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
)

type WindowData struct {
	ID    uint32
	Title string
	X, Y  int
	W, H  uint
}

func NewScreenShare(w *WindowData) (*StreamedProcess, error) {
	proc, err := ffmpegScreenShare(w)
	if err != nil {
		return nil, err
	}
	proc.cmd.Stderr = os.Stderr
	return proc, nil
}

func PublishScreenShare(room *lksdk.Room, proc *StreamedProcess, endCallback func()) error {
	return proc.Publish(
		room,
		webrtc.MimeTypeH264,
		33*time.Millisecond,
		endCallback,
		&lksdk.TrackPublicationOptions{
			Name:   "screen",
			Source: livekit.TrackSource_SCREEN_SHARE,
		},
	)
}
