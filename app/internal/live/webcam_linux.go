package live

import (
	"fmt"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
)

func NewWebcam(track *lksdk.LocalTrack) (*GstTrackWriter, error) {
	pipelineStr := fmt.Sprintf(
		"v4l2src device=/dev/video0 ! " +
			"image/jpeg,width=1280,height=720,framerate=30/1 !" +
			// "jpegparse ! " +
			"jpegdec ! " +
			"videoconvert ! " +
			"videoscale ! " +
			"video/x-raw,format=I420,width=1280,height=720,framerate=30/1 ! " +
			"x264enc tune=zerolatency speed-preset=ultrafast key-int-max=30 ! " +
			"h264parse config-interval=-1 ! " +
			"video/x-h264,stream-format=byte-stream,alignment=au ! " +
			"appsink name=sink sync=false emit-signals=true drop=true max-buffers=1",
	)

	return NewGstTrackWriter(
		track,
		pipelineStr,
		time.Second/30,
	)
}
