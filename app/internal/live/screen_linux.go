package live

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"syscall"
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
