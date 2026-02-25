package live

import (
	"fmt"
	"log"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

func NewGstTrackWriter(track *lksdk.LocalTrack, plstr string, duration time.Duration) (*GstTrackWriter, error) {
	gst.Init(nil)

	log.Printf("gst pipeline: %s", plstr)
	pipeline, err := gst.NewPipelineFromString(plstr)
	if err != nil {
		return nil, fmt.Errorf("pipeline error: %s", err.Error())
	}

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

			err := track.WriteSample(media.Sample{
				Data:     data,
				Duration: duration,
			}, nil)
			if err != nil {
				return gst.FlowError
			}

			return gst.FlowOK
		},
	})

	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		return nil, err
	}

	return &GstTrackWriter{
		pipeline,
		track,
		duration,
	}, nil
}

type GstTrackWriter struct {
	pipeline *gst.Pipeline
	track    *lksdk.LocalTrack
	duration time.Duration
}

func (w *GstTrackWriter) Close() error {
	return w.pipeline.SetState(gst.StateNull)
}
