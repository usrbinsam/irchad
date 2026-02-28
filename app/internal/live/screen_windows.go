package live

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/go-gst/go-gst/gst"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// Win32 API definitions
var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

func GetWindows() ([]WindowData, error) {
	var results []WindowData

	// The callback function passed to Win32 EnumWindows
	cb := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		// Filter out hidden system/background windows
		vis, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
		if vis == 0 {
			return 1 // 1 tells Win32 to continue enumerating the next window
		}

		// 1. Get Title
		var title string
		tLen, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
		if tLen > 0 {
			buf := make([]uint16, tLen+1)
			procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(tLen+1))
			title = syscall.UTF16ToString(buf)
		}
		if title == "" {
			title = "(unnamed)"
		}

		// 2. Get Class Name (The closest Win32 equivalent to X11 WM_CLASS)
		var wmClass string
		cBuf := make([]uint16, 256)
		cLen, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&cBuf[0])), uintptr(len(cBuf)))
		if cLen > 0 {
			wmClass = syscall.UTF16ToString(cBuf)
		}

		// 3. Get Process ID
		var pid uint32
		procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))

		// 4. Get Window Geometry
		var rect RECT
		procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))

		results = append(results, WindowData{
			ID:      uint32(hwnd), // Cast HWND pointer to uint32 to match X11
			Title:   title,
			X:       int(rect.Left),
			Y:       int(rect.Top),
			W:       uint(rect.Right - rect.Left),
			H:       uint(rect.Bottom - rect.Top),
			PID:     uint(pid),
			WMClass: wmClass,
		})

		return 1 // Continue enumerating
	})

	// Fire the enumeration
	ret, _, err := procEnumWindows.Call(cb, 0)

	// Go syscall quirk: Windows often sets a harmless error even on success.
	if ret == 0 {
		if err != nil && err.Error() != "The operation completed successfully." {
			return nil, fmt.Errorf("EnumWindows failed: %w", err)
		}
	}

	return results, nil
}

func NewScreenShare(w *WindowData, ss *ScreenShareOpts, audioTrack, videoTrack *lksdk.LocalSampleTrack) (*gst.Pipeline, error) {
	panic("not implemented")
}
