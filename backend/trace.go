package backend

import (
	"context"
	"fmt"
	"sync"

	"to-diag-trace-go/diag"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TraceStartOptions struct {
	AccumulateTraces bool `json:"accumulateTraces"`
	EnableOtsTrace   bool `json:"enableOtsTrace"`
	Verbosity        uint `json:"verbosity"`
}

type TraceProgressEvent struct {
	Collected int `json:"collected"`
	Total     int `json:"total"`
}

type TraceGatherDoneEvent struct {
	ZipPath        string   `json:"zipPath"`
	FailedFiles    []string `json:"failedFiles"`
	DeleteOnReboot bool     `json:"deleteOnReboot"`
	CleanupMessage string   `json:"cleanupMessage,omitempty"`
}

func (a *App) ensureDiag() error {
	if a.diag == nil {
		svc, err := diag.NewService()
		if err != nil {
			return err
		}
		a.diag = svc
	}
	return nil
}

func (a *App) GetTraceSettings() (diag.TraceSettings, error) {
	if err := a.ensureDiag(); err != nil {
		return diag.TraceSettings{}, err
	}
	return a.diag.GetTraceSettings()
}

func (a *App) IsTracingActive() (bool, error) {
	if err := a.ensureDiag(); err != nil {
		return false, err
	}
	return a.diag.IsTracingActive()
}

func (a *App) IsElevated() bool {
	if err := a.ensureDiag(); err != nil {
		return false
	}
	return a.diag.IsElevated()
}

func (a *App) RequestElevation() error {
	if err := a.ensureDiag(); err != nil {
		return err
	}
	if a.diag.IsElevated() {
		return nil
	}
	if err := a.diag.RelaunchElevated(""); err != nil {
		return err
	}
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
	return nil
}

func (a *App) StartTracing(opts TraceStartOptions) error {
	if err := a.ensureDiag(); err != nil {
		return err
	}
	if !a.diag.IsElevated() {
		return fmt.Errorf("administrator privileges required")
	}
	err := a.diag.StartTracing(diag.TraceOptions{
		AccumulateTraces: opts.AccumulateTraces,
		EnableOTSTrace:   opts.EnableOtsTrace,
		Verbosity:        opts.Verbosity,
	})
	if err == nil && a.ctx != nil {
		runtime.EventsEmit(a.ctx, "trace:state", map[string]bool{"isTracing": true})
		a.minimizeAfterTracingStart(a.ctx)
	}
	return err
}

func (a *App) StopTracing() error {
	if err := a.ensureDiag(); err != nil {
		return err
	}
	err := a.diag.StopTracing()
	if err == nil && a.ctx != nil {
		runtime.EventsEmit(a.ctx, "trace:state", map[string]bool{"isTracing": false})
	}
	return err
}

func (a *App) PickGatherDestination() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not ready")
	}
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save diagnostic archive",
		DefaultFilename: "DigitalPersona-Diagnostics.zip",
		Filters: []runtime.FileFilter{
			{DisplayName: "Zip archive (*.zip)", Pattern: "*.zip"},
		},
	})
}

var gatherMu sync.Mutex

func (a *App) StopTracingAndGather(destZip string) error {
	if destZip == "" {
		return fmt.Errorf("destination path required")
	}
	if err := a.ensureDiag(); err != nil {
		return err
	}
	if !a.diag.IsElevated() {
		return fmt.Errorf("administrator privileges required")
	}

	gatherMu.Lock()
	defer gatherMu.Unlock()

	if err := a.diag.StopTracing(); err != nil {
		return err
	}

	go func() {
		result, err := a.diag.Gather(destZip, func(p diag.GatherProgress) {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "trace:progress", TraceProgressEvent{
					Collected: p.Collected,
					Total:     p.Total,
				})
			}
		})
		if a.ctx == nil {
			return
		}
		if err != nil {
			runtime.EventsEmit(a.ctx, "trace:gather-error", map[string]string{"message": err.Error()})
			return
		}
		runtime.EventsEmit(a.ctx, "trace:gather-done", TraceGatherDoneEvent{
			ZipPath:        result.ZipPath,
			FailedFiles:    result.FailedFiles,
			DeleteOnReboot: result.DeleteOnReboot,
			CleanupMessage: result.CleanupMessage,
		})
		runtime.EventsEmit(a.ctx, "trace:state", map[string]bool{"isTracing": false})
	}()

	return nil
}

func (a *App) CancelGather() {
	if a.diag != nil {
		a.diag.CancelGather()
	}
}

func (a *App) emitInitialTraceState(ctx context.Context) {
	if a.diag == nil {
		return
	}
	active, err := a.diag.IsTracingActive()
	if err != nil {
		return
	}
	runtime.EventsEmit(ctx, "trace:state", map[string]bool{"isTracing": active})
}
