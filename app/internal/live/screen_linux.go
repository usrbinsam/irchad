package live

/*
#cgo LDFLAGS: -lX11
#include <stdlib.h>
#include <string.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>

typedef struct {
    unsigned long id;
		unsigned int pid;
    int x, y;
    unsigned int width, height;
    char* title;
		char* wm_class;
} WinInfo;

static Atom atom_client_list = None;
static Atom atom_wm_name = None;
static Atom atom_wm_pid = None;
static Atom atom_utf8 = None;

void init_atoms(Display* disp) {
    if (atom_client_list == None) {
        atom_client_list = XInternAtom(disp, "_NET_CLIENT_LIST", False);
        atom_wm_name     = XInternAtom(disp, "_NET_WM_NAME", False);
				atom_wm_pid      = XInternAtom(disp, "_NET_WM_PID", False);
        atom_utf8        = XInternAtom(disp, "UTF8_STRING", False);
    }
}

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

WinInfo get_win_info(Display* disp, Window win) {
    init_atoms(disp);
    WinInfo info = {0};
    info.id = (unsigned long)win;

    Atom type;
    int format;
    unsigned long nitems, after;
    unsigned char* title_data = NULL;

    if (XGetWindowProperty(disp, win, atom_wm_name, 0, 1024, False, atom_utf8,
                           &type, &format, &nitems, &after, &title_data) == Success && title_data) {
        info.title = (char*)title_data;
    } else {
        info.title = NULL;
    }

		XClassHint ch;
    if (XGetClassHint(disp, win, &ch)) {
        if (ch.res_class) {
            info.wm_class = strdup(ch.res_class);
            XFree(ch.res_class);
        } else if (ch.res_name) {
            info.wm_class = strdup(ch.res_name);
        }

        if (ch.res_name) {
            XFree(ch.res_name);
        }
    } else {
        info.wm_class = NULL;
    }

		unsigned char* pid_data = NULL;
		if (XGetWindowProperty(disp, win, atom_wm_pid, 0, 1, False, XA_CARDINAL,
                           &type, &format, &nitems, &after, &pid_data) == Success && pid_data) {
        if (type == XA_CARDINAL && format == 32 && nitems > 0) {
            info.pid = (unsigned int)(*((unsigned long*)pid_data));
        }
        XFree(pid_data); // Free immediately, we just needed the integer
    }

    Window root, child;
    unsigned int border, depth;
    int local_x, local_y;

    if (XGetGeometry(disp, win, &root, &local_x, &local_y, &info.width, &info.height, &border, &depth) != 0) {
        XTranslateCoordinates(disp, win, root, 0, 0, &info.x, &info.y, &child);
    }

    return info;
}
*/
import "C"

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"
	"unsafe"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4/pkg/media"
)

// this and the CGO was made by Google Gemini
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

		if title == "" {
			title = "(unnamed)"
		}

		var wmClass string
		if info.wm_class != nil {
			wmClass = C.GoString(info.wm_class)
			C.XFree(unsafe.Pointer(info.wm_class))
		}

		results = append(results, WindowData{
			ID:      uint32(info.id),
			Title:   title,
			X:       int(info.x),
			Y:       int(info.y),
			W:       uint(info.width),
			H:       uint(info.height),
			PID:     uint(info.pid),
			WMClass: wmClass,
		})
	}

	return results, nil
}

// func hasElement(name string) bool {
// 	factory := gst.Find(name)
// 	return factory != nil
// }

func preferredEncoder(w *WindowData, ss *ScreenShareOpts) string {
	return fmt.Sprintf(
		"ximagesrc xid=%d name=video_src use-damage=0 ! "+
			"videoconvert ! "+
			"videoscale add-borders=true ! "+
			"video/x-raw,format=NV12,width=%d,height=%d,framerate=%d/1,pixel-aspect-ratio=1/1 ! "+
			"vah264enc bitrate=%d ! "+
			"h264parse config-interval=-1 ! "+
			"video/x-h264,stream-format=byte-stream,alignment=au ! ",
		w.ID, w.W&^1, w.H&^1, ss.FrameRate, ss.BitRate,
	)
}

func NewGstScreenShare(w *WindowData, track *lksdk.LocalTrack, ss *ScreenShareOpts) (*GstTrackWriter, error) {
	pipelineStr := preferredEncoder(w, ss)
	log.Printf("selected encoder: %s\n", pipelineStr)

	return NewGstTrackWriter(track, pipelineStr, time.Second/30)
}

type PwNode struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
	Info struct {
		Props map[string]any `json:"props"`
	} `json:"info"`
}

func getPipewireNodeID(applicationPID uint) (uint, error) {
	out := bytes.Buffer{}
	cmd := exec.Command("pw-dump")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	dec := json.NewDecoder(&out)
	var nodes []PwNode
	err = dec.Decode(&nodes)
	if err != nil {
		return 0, err
	}

	searchPID := strconv.Itoa(int(applicationPID))
	for _, node := range nodes {
		if node.Type != "PipeWire:Interface:Node" {
			continue
		}

		mediaClass, ok := node.Info.Props["media.class"].(string)
		if !ok || mediaClass != "Stream/Output/Audio" {
			continue
		}

		nodePIDf64, ok := node.Info.Props["application.process.id"].(float64)
		if !ok {
			continue
		}

		nodePID := strconv.Itoa(int(nodePIDf64))

		if nodePID == searchPID {
			serial, ok := node.Info.Props["object.serial"].(float64)
			if !ok {
				return 0, fmt.Errorf("pw node missing `object.serial' prop")
			}
			return uint(serial), nil
		}
	}

	return 0, fmt.Errorf("no pipewire node found")
}

func listChildPIDs(parentPID uint) ([]uint, error) {
	out := bytes.Buffer{}
	parentPIDStr := strconv.Itoa(int(parentPID))
	cmd := exec.Command("pgrep", "-P", parentPIDStr)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(&out)
	pids := make([]uint, 0)
	for scanner.Scan() {
		pidStr := scanner.Text()
		pidInt, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Printf("invalid output from pgrep %q - %s", pidStr, err.Error())
			continue
		}
		pids = append(pids, uint(pidInt))
	}
	return pids, nil
}

func NewGstAppAudioShare(w *WindowData, track *lksdk.LocalTrack) (*GstTrackWriter, error) {
	log.Printf("grabbing audio for window %+v\n", w)
	nodeID, err := getPipewireNodeID(w.PID)
	if err != nil {
		log.Printf("no pipewire node for pid %d. searching child PIDs", w.PID)
		childPIDs, err := listChildPIDs(w.PID)
		if err != nil {
			log.Printf("error looking up child PIDs for %d: %s", w.PID, err.Error())
			return nil, err
		}
		for _, pid := range childPIDs {
			childNodeID, err := getPipewireNodeID(pid)
			if err != nil {
				log.Printf("no pipewire node for child PID %d", childNodeID)
				continue
			}
			nodeID = childNodeID
			break
		}
	}

	log.Printf("found pipewire node %d", nodeID)
	pipelineStr := fmt.Sprintf(
		"pipewiresrc target-object=%d ! "+
			"audioconvert ! "+
			"audioresample ! "+
			"audio/x-raw,format=S16LE,layout=interleaved,rate=48000,channels=2 ! "+
			"opusenc bitrate=64000 frame-size=20 bitrate-type=vbr bandwidth=fullband ! "+
			"appsink name=sink sync=false emit-signals=true drop=true max-buffers=1",
		nodeID,
	)
	return NewGstTrackWriter(track, pipelineStr, 20*time.Millisecond)
}

func huntPipeWireTarget(w *WindowData) (uint, error) {
	nodeID, err := getPipewireNodeID(w.PID)
	if err != nil {
		log.Printf("no pipewire node for pid %d. searching child PIDs", w.PID)
		childPIDs, err := listChildPIDs(w.PID)
		if err != nil {
			log.Printf("error looking up child PIDs for %d: %s", w.PID, err.Error())
			return 0, err
		}
		for _, pid := range childPIDs {
			childNodeID, err := getPipewireNodeID(pid)
			if err != nil {
				log.Printf("no pipewire node for child PID %d", childNodeID)
				continue
			}
			nodeID = childNodeID
			break
		}
	}

	return nodeID, nil
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
		// log.Printf("wrote %d bytes to from sink %s to track %s", len(data), name, track.ID())
	}
}

func NewScreenShare(w *WindowData, opts *ScreenShareOpts, audioTrack, videoTrack *lksdk.LocalSampleTrack) (*gst.Pipeline, error) {
	videoEncoder := preferredEncoder(w, opts)
	pwTarget, err := huntPipeWireTarget(w)
	if err != nil {
		return nil, err
	}
	pipelineStr := fmt.Sprintf(
		"%s"+
			"appsink name=video_sink sync=false emit-signals=true drop=true max-buffers=1 "+
			"pipewiresrc target-object=%d ! "+
			"audioconvert ! "+
			"audioresample ! "+
			"audio/x-raw,format=S16LE,layout=interleaved,rate=48000,channels=2 ! "+
			"opusenc bitrate=64000 frame-size=20 bitrate-type=vbr bandwidth=fullband ! "+
			"appsink name=audio_sink sync=false emit-signals=true drop=true max-buffers=1",
		videoEncoder, pwTarget,
	)

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		log.Fatalf("pipeline err: %s", err.Error())
	}

	videoElem, err := pipeline.GetElementByName("video_sink")
	if err != nil {
		log.Fatalf("videoBin.GetElements(): %s", err.Error())
	}
	videoSink := app.SinkFromElement(videoElem)

	audioElem, err := pipeline.GetElementByName("audio_sink")
	if err != nil {
		log.Fatalf("audioBin.GetElements(): %s", err.Error())
	}
	audioSink := app.SinkFromElement(audioElem)

	go pushTrack(videoSink, videoTrack)
	go pushTrack(audioSink, audioTrack)

	err = pipeline.SetState(gst.StatePlaying)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}
