package live

import (
	"fmt"
	"io"
	"log"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

type GstStreamer struct {
	pipeline *gst.Pipeline
	track    *lksdk.LocalTrack
	reader   *io.PipeReader
	writer   *io.PipeWriter
}

func NewGstScreenShare(w *WindowData, track *lksdk.LocalTrack) (*GstStreamer, error) {
	gst.Init(nil)

	pipelineStr := fmt.Sprintf(
		"ximagesrc xid=%d use-damage=0 ! "+
			"videoconvert ! "+
			"videoscale ! "+
			"video/x-raw,format=I420,framerate=30/1,width=%d,height=%d ! "+
			"x264enc tune=zerolatency speed-preset=ultrafast key-int-max=30 ! "+
			"h264parse config-interval=-1 ! "+
			"video/x-h264,stream-format=byte-stream,alignment=au ! "+ // FORCE ANNEX-B
			"appsink name=sink sync=false emit-signals=true drop=true max-buffers=1",
		w.ID, w.W&^1, w.H&^1,
	)

	log.Println(pipelineStr)

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()

	sinkElement, _ := pipeline.GetElementByName("sink")
	sink := app.SinkFromElement(sinkElement)

	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
			sample := s.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			buffer := sample.GetBuffer()
			data := buffer.Bytes()
			log.Printf("GST pipeline produced %d bytes", len(data))

			err := track.WriteSample(media.Sample{
				Data:     data,
				Duration: time.Second / 30,
			}, nil)
			if err != nil {
				return gst.FlowError
			}

			return gst.FlowOK
		},
	})

	return &GstStreamer{
		pipeline: pipeline,
		reader:   pr,
		writer:   pw,
	}, nil
}

func (g *GstStreamer) Read(p []byte) (n int, err error) {
	log.Println("--> GST READ")
	return g.reader.Read(p)
}

func (g *GstStreamer) Start() error {
	log.Println("GstStreamer.Start")
	return g.pipeline.SetState(gst.StatePlaying)
}

func (g *GstStreamer) Stop() error {
	// 1. Close the pipe to unblock any pending Read() calls
	g.writer.Close()
	g.reader.Close()

	// 2. Transition GStreamer state
	return g.pipeline.SetState(gst.StateNull)
}

func (g *GstStreamer) Close() error {
	return g.Stop()
}
