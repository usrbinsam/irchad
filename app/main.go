package main

import (
	"embed"
	"os"
	"runtime"

	"IrChad/internal/live"

	"github.com/go-gst/go-gst/gst"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

func init() {
	live.RegisterEvents()
}

func main() {
	if runtime.GOOS == "linux" {
		_ = os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}
	gst.Init(nil)
	app := application.New(
		application.Options{
			Name: "IrChad",
			Services: []application.Service{
				application.NewService(live.NewLiveChat()),
			},
			Assets: application.AssetOptions{
				Handler: application.AssetFileServerFS(assets),
			},
		},
	)

	/*window :=*/
	app.Window.NewWithOptions(
		application.WebviewWindowOptions{
			Title:  "IrChad",
			Width:  1024,
			Height: 768,
		},
	)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
