package live

type (
	FrameRate  int
	Resolution int
)

type WindowData struct {
	ID    uint32
	Title string
	X, Y  int
	W, H  uint
}

type ScreenShareParams struct {
	FrameRate  FrameRate
	Resolution Resolution
}

type ScreenSharer interface {
	Stop() error
}
