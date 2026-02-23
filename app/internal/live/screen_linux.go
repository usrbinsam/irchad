package live

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

func GetWindows() ([]WindowData, error) {
	cmd := exec.Command("/home/sam/xgb-window-lister/xgb-window-lister")
	buf := bytes.Buffer{}

	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		log.Printf("wmctrl err: %s", err.Error())
		return nil, err
	}

	windows := make([]WindowData, 0)

	dec := json.NewDecoder(&buf)
	err = dec.Decode(&windows)
	if err != nil {
		log.Printf("invalid json from xgb-window-lister: %s", err.Error())
		return nil, err
	}
	return windows, nil
}

func (w *WindowData) Thumbnail() ([]byte, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "x11grab",
		"-video_size", fmt.Sprintf("%dx%d", w.W, w.H),
		"-i", fmt.Sprintf(":0.0+%d,%d", w.X, w.Y),
		"-frames:v", "1",
		"-f", "image2",
		"-vcodec", "mjpeg",
		"pipe:1",
	)

	buf := bytes.Buffer{}
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ffmpegScreenShare(w *WindowData) (*StreamedProcess, error) {
	proc, err := NewStreamedProcess(
		"ffmpeg",
		"-f",
		"x11grab",
		"-video_size", fmt.Sprintf("%dx%d", w.W, w.H),
		"-i", fmt.Sprintf(":0.0+%d,%d", w.X, w.Y),
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2", // Force even dimensions
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-pix_fmt", "yuv420p",
		"-f", "h264",
		"pipe:1",
	)
	if err != nil {
		return nil, err
	}
	proc.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	return proc, nil
}

type GstStreamer struct {
	pipeline *gst.Pipeline
	track    *lksdk.LocalTrack
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
	}, nil
}

func (g *GstStreamer) Start() error {
	return g.pipeline.SetState(gst.StatePlaying)
}
