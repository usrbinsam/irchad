package live

import (
	"fmt"
	"syscall"
	"unsafe"
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

	cb := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
		if vis == 0 {
			return 1 // 1 tells Win32 to continue enumerating the next window
		}

		var title string
		tLen, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
		if tLen > 0 {
			buf := make([]uint16, tLen+1)
			procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(tLen+1))
			title = syscall.UTF16ToString(buf)
		}
		if title == "" {
			return 1
		}

		var wmClass string
		cBuf := make([]uint16, 256)
		cLen, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&cBuf[0])), uintptr(len(cBuf)))
		if cLen > 0 {
			wmClass = syscall.UTF16ToString(cBuf)
		}

		var pid uint32
		procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))

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

		return 1
	})

	ret, _, err := procEnumWindows.Call(cb, 0)

	// Go syscall quirk: Windows often sets a harmless error even on success.
	if ret == 0 {
		if err != nil && err.Error() != "The operation completed successfully." {
			return nil, fmt.Errorf("EnumWindows failed: %w", err)
		}
	}

	return results, nil
}

func screenCaptureSourceElement(w *WindowData) string {
	common := "do-timestamp=true show-cursor=true "
	if w.MonitorIndex != 0 {
		return fmt.Sprintf(
			"d3d12screencapturesrc %s monitor-index=%d ! ",
			common, w.MonitorIndex,
		)
	}
	return fmt.Sprintf(
		"d3d12screencapturesrc %s crop-x=%d crop-y=%d crop-width=%d crop-height=%d ! ",
		common, w.X, w.Y, w.W, w.H,
	)
}

func screenAudioSourceElement(w *WindowData) string {
	if w.PID != 0 {
		return fmt.Sprintf("wasapi2src loopback=true loopback-mode=include-process-tree loopback-target-pid=%d do-timestamp=true ! ", w.PID)
	}
	// send silence if we're sharing the whole desktop
	return "audiotestsrc wave=silence is-live=true ! "
}
