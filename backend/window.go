package backend

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) SetLaunchMode(autoMode bool) {
	a.autoMode = autoMode
}

func (a *App) restoreLaunchWindow(ctx context.Context) {
	if a.autoMode {
		runtime.WindowShow(ctx)
	}
}

func (a *App) minimizeAfterTracingStart(ctx context.Context) {
	runtime.WindowMinimise(ctx)
}
