package live

type (
	FrameRate  int
	Resolution int
)

type WindowData struct {
	ID      uint32
	Title   string
	X, Y    int
	W, H    uint
	PID     uint
	WMClass string
}

type ScreenShareOpts struct {
	FrameRate FrameRate
	BitRate   int
}

type ScreenSharer interface {
	Stop() error
}
