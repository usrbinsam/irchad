package main

import (
	"embed"

	"IrChad/internal/live"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

func init() {
	live.RegisterEvents()
}

func main() {
	app := application.New(
		application.Options{
			Name: "IrChad",
			Services: []application.Service{
				application.NewService(live.NewLiveChat("ws://127.0.0.1:7880")),
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

	// TODO: this should only run on Linux
	// go func() {
	// 	fmt.Printf("hacking window\n")
	// 	for {
	// 		time.Sleep(500 * time.Millisecond)
	// 		ptr := window.NativeWindow()
	// 		if ptr == nil {
	// 			continue
	// 		}
	// 		hack.HackAllowGetUserMedia(ptr)
	// 		break
	// 	}
	// }()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
