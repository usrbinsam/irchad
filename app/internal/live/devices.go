package live

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4/pkg/media"
)

var rmsRegex = regexp.MustCompile(`rms=\(GValueArray\)<\s*([^>]+)\s*>`)

func NewGstTrackWriter(track *lksdk.LocalTrack, plstr string, duration time.Duration) (*GstTrackWriter, error) {
	gst.Init(nil)

	log.Printf("gst pipeline: %s", plstr)
	pipeline, err := gst.NewPipelineFromString(plstr)
	if err != nil {
		return nil, fmt.Errorf("pipeline error: %s", err.Error())
	}

	writer := &GstTrackWriter{
		pipeline,
		track,
		duration,
		127,
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
			}, &lksdk.SampleWriteOptions{AudioLevel: &writer.audioLevel})
			if err != nil {
				return gst.FlowError
			}

			return gst.FlowOK
		},
	})

	bus := pipeline.GetBus()
	bus.AddWatch(writer.setAudioLevel)

	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		return nil, err
	}

	return writer, nil
}

type GstTrackWriter struct {
	pipeline   *gst.Pipeline
	track      *lksdk.LocalTrack
	duration   time.Duration
	audioLevel uint8
}

func (w *GstTrackWriter) Close() error {
	return w.pipeline.SetState(gst.StateNull)
}

const SensitivityBoost = 18.0

func (w *GstTrackWriter) setAudioLevel(msg *gst.Message) bool {
	if msg.Type() == gst.MessageElement {
		if structure := msg.GetStructure(); structure != nil && structure.Name() == "level" {
			match := rmsRegex.FindStringSubmatch(structure.String())

			if len(match) == 2 {
				rmsStrs := strings.Split(match[1], ",")
				highestDB := -127.0

				for _, rmsStr := range rmsStrs {
					db, err := strconv.ParseFloat(strings.TrimSpace(rmsStr), 64)
					if err == nil && db > highestDB {
						highestDB = db
					}
				}

				newLevel := uint8(math.Abs(math.Max(-127, math.Min(0, highestDB))))
				w.audioLevel = newLevel - SensitivityBoost
				// log.Println(rmsStrs)
				// log.Printf("audioLevel = %d", newLevel)
			}
		}
	}
	return true
}
