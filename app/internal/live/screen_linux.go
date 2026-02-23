package live

/*
#cgo LDFLAGS: -lX11
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>

typedef struct {
    unsigned long id;
    int x, y;
    unsigned int width, height;
    char* title; // Holds raw Xlib pointer, Go is responsible for calling XFree
} WinInfo;

// Cache atoms statically to avoid X11 server round-trips in loops
static Atom atom_client_list = None;
static Atom atom_wm_name = None;
static Atom atom_utf8 = None;

void init_atoms(Display* disp) {
    if (atom_client_list == None) {
        atom_client_list = XInternAtom(disp, "_NET_CLIENT_LIST", False);
        atom_wm_name     = XInternAtom(disp, "_NET_WM_NAME", False);
        atom_utf8        = XInternAtom(disp, "UTF8_STRING", False);
    }
}

// Helper to get the window list from the Root Window
Window* get_window_list(Display* disp, unsigned long* len) {
    init_atoms(disp);
    Atom actual_type;
    int actual_format;
    unsigned long nitems, bytes_after;
    unsigned char* data = NULL;

    if (XGetWindowProperty(disp, XDefaultRootWindow(disp), atom_client_list, 0, 1024, False, XA_WINDOW,
                           &actual_type, &actual_format, &nitems, &bytes_after, &data) == Success && data) {
        *len = nitems;
        return (Window*)data;
    }
    return NULL;
}

// Helper to get title and geometry for a specific window
WinInfo get_win_info(Display* disp, Window win) {
    init_atoms(disp);
    WinInfo info = {0};
    info.id = (unsigned long)win;

    // 1. Get Title (UTF-8)
    Atom type;
    int format;
    unsigned long nitems, after;
    unsigned char* title_data = NULL;

    if (XGetWindowProperty(disp, win, atom_wm_name, 0, 1024, False, atom_utf8,
                           &type, &format, &nitems, &after, &title_data) == Success && title_data) {
        info.title = (char*)title_data; // Pass directly to Go to avoid strdup
    } else {
        info.title = NULL;
    }

    // 2. Get Geometry & Translate to Absolute Coordinates
    Window root, child;
    unsigned int border, depth;
    int local_x, local_y;

    // Check return value (returns 0 on failure, e.g., if window just closed)
    if (XGetGeometry(disp, win, &root, &local_x, &local_y, &info.width, &info.height, &border, &depth) != 0) {
        XTranslateCoordinates(disp, win, root, 0, 0, &info.x, &info.y, &child);
    }

    return info;
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

func GetWindows() ([]WindowData, error) {
	display := C.XOpenDisplay(nil)
	if display == nil {
		return nil, fmt.Errorf("could not open X display")
	}
	defer C.XCloseDisplay(display)

	var length C.ulong
	winListPtr := C.get_window_list(display, &length)
	if winListPtr == nil {
		return nil, nil
	}
	defer C.XFree(unsafe.Pointer(winListPtr))

	cWindows := unsafe.Slice((*C.Window)(unsafe.Pointer(winListPtr)), length)
	results := make([]WindowData, 0, length)

	for _, win := range cWindows {
		info := C.get_win_info(display, win)

		var title string
		if info.title != nil {
			title = C.GoString(info.title)
			C.XFree(unsafe.Pointer(info.title))
		}

		if info.x < 0 || info.y < 0 {
			continue
		}

		if title == "" {
			title = "(unnamed)"
		}

		results = append(results, WindowData{
			ID:    uint32(info.id),
			Title: title,
			X:     int(info.x),
			Y:     int(info.y),
			W:     uint(info.width),
			H:     uint(info.height),
		})
	}

	return results, nil
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
