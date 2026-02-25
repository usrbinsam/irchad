package live

import (
	"fmt"
	"io"
	"log"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

const maxVideoLate = 1000

func (l *LiveChat) decodeVideoStream(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) (*VideoStream, error) {
	gst.Init(nil)

	participantID := rp.Identity()
	trackID := track.ID()

	trackCodec := track.Codec().MimeType
	if trackCodec != webrtc.MimeTypeH264 {
		return nil, fmt.Errorf("can only decode h264")
	}

	pipelineStr := "appsrc name=irchad format=time is-live=true ! " +
		"h264parse ! " +
		"avdec_h264 !" +
		"videoconvert ! " +
		"queue ! " +
		"jpegenc quality=85 ! " +
		"appsink name=sink sync=false emit-signals=true drop=true max-buffers=1"

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		return nil, fmt.Errorf("pipeline error: %s", err.Error())
	}

	mjpegReader, mjpegWriter := io.Pipe()
	sinkElem, err := pipeline.GetElementByName("sink")
	if sinkElem == nil {
		return nil, fmt.Errorf("sinkElem == nil")
	}
	if err != nil {
		log.Printf("couldn't find sink element: %s", err.Error())
		return nil, err
	}
	log.Printf("sinkElement: %+v\n", sinkElem)

	sink := app.SinkFromElement(sinkElem)
	log.Printf("sink: %+v", sink)
	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
			sample := s.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			data := sample.GetBuffer().Bytes()
			fmt.Fprintf(mjpegWriter, "--irchad\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(data))
			_, _ = mjpegWriter.Write(data)
			fmt.Fprintf(mjpegWriter, "\r\n")
			return gst.FlowOK
		},
	})

	err = pipeline.SetState(gst.StatePlaying)
	if err != nil {
		log.Printf("couldn't set pipeline to playing: %s", err.Error())
		return nil, err
	}

	vStream := &VideoStream{
		Name:   pub.Name(),
		stream: mjpegReader,
	}

	l.registry.Add(
		participantID,
		trackID,
		&VideoTrackHandler{stream: vStream},
	)

	go func() {
		defer mjpegWriter.Close()
		defer pipeline.SetState(gst.StateNull)
		srcElem, err := pipeline.GetElementByName("irchad")
		if err != nil {
			log.Printf("failed to get src 'irchad' - %s", err.Error())
			return
		}
		src := app.SrcFromElement(srcElem)
		sb := samplebuilder.New(
			maxVideoLate,
			&codecs.H264Packet{},
			track.Codec().ClockRate,
		)
		for {
			packet, _, err := track.ReadRTP()
			if err != nil {
				break
			}

			sb.Push(packet)

			for {
				sample := sb.Pop()
				if sample == nil {
					break
				}

				buffer := gst.NewBufferFromBytes(sample.Data)
				buffer.SetDuration(sample.Duration)
				if src.PushBuffer(buffer) != gst.FlowOK {
					break
				}
			}
		}

		l.registry.Remove(participantID, trackID)
	}()

	return vStream, nil
}

func (l *LiveChat) decodeAudioStream(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	stream := &OpusStream{}

	participantID := rp.Identity()
	trackID := track.ID()

	log.Printf("decoding audio track from %s: %s %s", rp.Identity(), pub.Name(), pub.Source())
	l.registry.Add(
		participantID,
		trackID,
		&AudioTrackHandler{stream: stream},
	)
	go func() {
		ogg, err := oggwriter.NewWith(stream, 48000, track.Codec().Channels)
		if err != nil {
			log.Printf("ogg writer error: %s", err.Error())
			return
		}
		defer ogg.Close()

		for {
			packet, _, err := track.ReadRTP()
			if err != nil {
				break
			}

			if err := ogg.WriteRTP(packet); err != nil {
				break
			}
		}

		log.Printf("decoding ended for %s: %s %s", rp.Identity(), pub.Name(), pub.Source())
		l.registry.Remove(participantID, trackID)
	}()
}
