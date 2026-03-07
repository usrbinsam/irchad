package live

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

type (
	FrameRate  int
	Resolution int
)

type WindowData struct {
	ID           uint32
	Title        string
	X, Y         int
	W, H         uint
	PID          uint
	WMClass      string
	MonitorIndex int
}

type ScreenShareOpts struct {
	FrameRate FrameRate
	BitRate   int
}

type ScreenSharer interface {
	Stop() error
}

func pushTrack(appSink *app.Sink, track *lksdk.LocalSampleTrack) {
	sinkProp, err := appSink.GetProperty("name")
	if err != nil {
		log.Fatalf("sink has no name?")
	}

	name := sinkProp.(string)

	for {
		sample := appSink.PullSample()
		if sample == nil {
			if appSink.IsEOS() {
				log.Printf("end of stream. track=%s sink=%s", track.ID(), name)
				return
			}
			continue
		}
		buffer := sample.GetBuffer()
		if buffer == nil {
			continue
		}

		data := buffer.Bytes()
		dur := buffer.Duration().AsDuration()

		webrtcSample := media.Sample{
			Data:     data,
			Duration: *dur,
		}
		if err := track.WriteSample(webrtcSample, nil); err != nil {
			log.Printf("write sample error: %s", err.Error())
			return
		}
	}
}

func hasElement(name string) bool {
	factory := gst.Find(name)
	return factory != nil
}

func preferredEncoder(w *WindowData, ss *ScreenShareOpts) string {
	// Ensure width and height are even numbers (macroblock requirement for H.264)
	width := w.W &^ 1
	height := w.H &^ 1

	// Base scaling and framerate (Common to all encoders)
	basePipeline := fmt.Sprintf(
		"videoconvert ! "+
			"videoscale add-borders=true ! "+
			"video/x-raw,width=%d,height=%d,framerate=%d/1,pixel-aspect-ratio=1/1 ! ",
		width, height, ss.FrameRate,
	)

	var encoder string

	// 1. NVIDIA Hardware Encoding
	if hasElement("nvh264enc") {
		encoder = fmt.Sprintf(
			"videoconvert ! video/x-raw,format=NV12 ! nvh264enc bitrate=%d zerolatency=true ! ",
			ss.BitRate,
		)
	} else if hasElement("vah264enc") {
		// 2. Modern VAAPI Hardware Encoding (AMD/Intel)
		encoder = fmt.Sprintf(
			"videoconvert ! video/x-raw,format=NV12 ! vah264enc bitrate=%d ! ",
			ss.BitRate,
		)
	} else if hasElement("vaapih264enc") {
		// 3. Legacy VAAPI Hardware Encoding
		encoder = fmt.Sprintf(
			"videoconvert ! video/x-raw,format=NV12 ! vaapih264enc bitrate=%d ! ",
			ss.BitRate,
		)
	} else {
		// 4. CPU Software Encoding (Fallback)
		encoder = fmt.Sprintf(
			"videoconvert ! video/x-raw,format=I420 ! x264enc bitrate=%d tune=zerolatency speed-preset=ultrafast ! ",
			ss.BitRate,
		)
	}

	tail := "h264parse config-interval=-1 ! video/x-h264,stream-format=byte-stream,alignment=au ! "
	return basePipeline + encoder + tail
}

type ScreenShare struct {
	VideoTrack *lksdk.LocalSampleTrack
	AudioTrack *lksdk.LocalSampleTrack

	pipeline *gst.Pipeline
}

func NewScreenShare(w *WindowData, opts *ScreenShareOpts) (*ScreenShare, error) {
	pipelineStr := screenCaptureSourceElement(w) +
		"tee name=vtee vtee. ! " +

		// webrtc branch
		"queue ! " +
		preferredEncoder(w, opts) +
		"appsink name=video_sink sync=false emit-signals=true drop=true max-buffers=1 " +

		// video preview
		"vtee. ! queue ! videoconvert ! jpegenc ! appsink name=preview_sink sync=false emit-signals=true drop=true max-buffers=1 " +

		// audio
		screenAudioSourceElement(w) +
		"queue max-size-time=1000000000 ! " +
		"audioconvert ! " +
		"audioresample ! " +
		"audiorate ! " +
		"audio/x-raw,format=S16LE,layout=interleaved,rate=48000,channels=2 ! " +
		"opusenc bitrate=64000 frame-size=20 bitrate-type=vbr bandwidth=fullband ! " +
		"appsink name=audio_sink sync=false emit-signals=true drop=true max-buffers=100"

	log.Printf("screen share pipeline: %s", pipelineStr)
	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		log.Fatalf("pipeline err: %s", err.Error())
	}

	videoElem, err := pipeline.GetElementByName("video_sink")
	if err != nil {
		log.Fatalf("pipeline err: %s", err.Error())
	}
	videoSink := app.SinkFromElement(videoElem)

	audioElem, err := pipeline.GetElementByName("audio_sink")
	if err != nil {
		log.Fatalf("pipeline err: %s", err.Error())
	}
	audioSink := app.SinkFromElement(audioElem)

	audioTrack, err := lksdk.NewLocalSampleTrack(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
	)
	if err != nil {
		log.Printf("error creating audio track: %s", err.Error())
		return nil, err
	}

	videoTrack, err := lksdk.NewLocalSampleTrack(
		webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
		},
		lksdk.WithRTCPHandler(func(packet rtcp.Packet) {
			switch packet.(type) {
			case *rtcp.PictureLossIndication:
				log.Println("got picture loss indicater, sending Key Frame")
				st := gst.NewStructure("GstForceKeyUnit")
				event := gst.NewCustomEvent(gst.EventTypeCustomUpstream, st)
				if !videoSink.SendEvent(event) {
					log.Println("failed to send GstForceKeyUnit event")
				}
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	go pushTrack(videoSink, videoTrack)
	go pushTrack(audioSink, audioTrack)

	err = pipeline.SetState(gst.StatePlaying)
	if err != nil {
		return nil, err
	}

	return &ScreenShare{
		AudioTrack: audioTrack,
		VideoTrack: videoTrack,
		pipeline:   pipeline,
	}, nil
}

func (s *ScreenShare) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "multipart/x-mixed-replace; boundary=irchad")
	w.Header().Set("connection", "keep-alive")
	w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")

	flusher := w.(http.Flusher)
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	previewElem, err := s.pipeline.GetElementByName("preview_sink")
	if err != nil {
		log.Fatalf("pipeline missing 'preview_sink' element")
	}

	sink := app.SinkFromElement(previewElem)

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			sample := sink.PullSample()
			if sample == nil {
				if sink.IsEOS() {
					return
				}

				continue
			}

			frame := sample.GetBuffer().Bytes()
			_, err = fmt.Fprintf(
				w,
				"--irchad\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n",
				len(frame),
			)
			if err != nil {
				return
			}

			_, err := w.Write(frame)
			if err != nil {
				log.Printf("err writing to browser: %s", err.Error())
				return
			}

			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				return
			}

			flusher.Flush()
		}
	}
}

func (s *ScreenShare) Close() {
	s.pipeline.SetState(gst.StateNull)
}
