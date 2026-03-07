package live

import (
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
)

func NewMicrophone(track *lksdk.LocalTrack) (*GstTrackWriter, error) {
	pipelineStr := "pulsesrc buffer-time=20000 latency-time=5000 do-timestamp=true ! " +
		"queue max-size-time=500000000 leaky=downstream ! " +
		"audioconvert ! " +
		"audioresample ! " +
		"audiornnoise ! " +
		"level name=vadel ! " +
		"audioconvert ! " +
		"audioresample ! " +
		"audio/x-raw,format=S16LE,layout=interleaved,rate=48000,channels=2 ! " +
		"opusenc dtx=true bitrate=64000 frame-size=20 bitrate-type=vbr bandwidth=fullband ! " +
		"appsink name=sink sync=false emit-signals=true drop=true max-buffers=25"

	return NewGstTrackWriter(
		track,
		pipelineStr,
		20*time.Millisecond,
	)
}
