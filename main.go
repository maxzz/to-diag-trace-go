package main

import (
	"embed"
	"to-diag-trace-go/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Create an instance of the app structure
	app := backend.NewApp()

	// Load options on startup to get initial width/height
	initialWidth := 1200
	initialHeight := 800

	opts, err := backend.LoadIniFileOptions()
	if err == nil && opts != nil && opts.Bounds != nil {
		bounds := backend.FixBounds(opts.Bounds)
		if bounds != nil {
			initialWidth = bounds.Width
			initialHeight = bounds.Height
		}
	}

	// Create application with options
	openInspector := false
	if err == nil && opts != nil {
		openInspector = opts.DevTools
	}

	err = wails.Run(&options.App{
		Title:            "wails template",
		Width:            initialWidth,
		Height:           initialHeight,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		OnDomReady:       app.DomReady,
		OnBeforeClose:    app.BeforeClose,
		StartHidden:      true,
		Debug: options.Debug{
			OpenInspectorOnStartup: openInspector,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
