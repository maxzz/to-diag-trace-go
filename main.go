package main

import (
	"embed"
	"os"
	"strings"

	"to-diag-trace-go/backend"
	"to-diag-trace-go/diag"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	_ = appIcon
	args := os.Args[1:]

	if len(args) > 0 && strings.HasPrefix(args[0], diag.CmdDeleteFiles) {
		if err := diag.HandleDeleteFilesCLI(args); err != nil {
			println("Error:", err.Error())
		}
		return
	}

	release, ok, err := diag.AcquireSingleInstance()
	if err != nil {
		println("Error:", err.Error())
		return
	}
	if !ok {
		diag.ActivateExistingInstance()
		return
	}
	defer release()

	autoMode := len(args) > 0 && strings.EqualFold(strings.TrimSpace(args[0]), diag.CmdAuto)
	if autoMode {
		svc, err := diag.NewService()
		if err != nil {
			println("Error:", err.Error())
			return
		}
		active, err := svc.IsTracingActive()
		if err != nil || !active {
			return
		}
	} else if backend.RequireElevationAtLaunchDefault(backend.LoadIniFileOptionsOrDefault()) && !diag.IsElevated() {
		if err := diag.RelaunchElevated(""); err != nil {
			println("Error:", err.Error())
		}
		return
	}

	app := backend.NewApp()
	app.SetLaunchMode(autoMode)

	initialWidth := 420
	initialHeight := 520

	opts, err := backend.LoadIniFileOptions()
	if err == nil && opts != nil && opts.Bounds != nil {
		bounds := backend.FixBounds(opts.Bounds)
		if bounds != nil {
			initialWidth = bounds.Width
			initialHeight = bounds.Height
		}
	}

	openInspector := false
	if err == nil && opts != nil {
		openInspector = opts.DevTools
	}

	err = wails.Run(&options.App{
		Title:  "DigitalPersona Diagnostic Tool",
		Width:  initialWidth,
		Height: initialHeight,
		MinWidth:  380,
		MinHeight: 400,
		Assets:    assets,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.Startup,
		OnDomReady:       app.DomReady,
		OnBeforeClose:    app.BeforeClose,
		StartHidden:      autoMode,
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
