package backend

import (
	"fmt"

	"to-diag-trace-go/diag"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetAppLaunchSettings() (AppLaunchSettings, error) {
	opts := LoadIniFileOptionsOrDefault()
	return AppLaunchSettings{
		RequireElevationAtLaunch: RequireElevationAtLaunchDefault(opts),
		IsElevated:               a.IsElevated(),
	}, nil
}

func (a *App) SetRequireElevationAtLaunch(require bool) error {
	opts := LoadIniFileOptionsOrDefault()
	opts.RequireElevationAtLaunch = &require
	return saveIniFileOptions(opts)
}

func (a *App) RequestElevationForAction(action PendingAction) error {
	if a.IsElevated() {
		return a.executePendingAction(action)
	}

	opts := LoadIniFileOptionsOrDefault()
	opts.PendingAction = &action
	if err := saveIniFileOptions(opts); err != nil {
		return err
	}
	if err := diag.RelaunchElevated(""); err != nil {
		opts.PendingAction = nil
		_ = saveIniFileOptions(opts)
		return err
	}
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
	return nil
}

func (a *App) resumePendingAction() {
	opts, err := LoadIniFileOptions()
	if err != nil || opts == nil || opts.PendingAction == nil {
		return
	}
	if !a.IsElevated() {
		return
	}

	action := *opts.PendingAction
	opts.PendingAction = nil
	_ = saveIniFileOptions(opts)
	_ = a.executePendingAction(action)
}

func (a *App) executePendingAction(action PendingAction) error {
	switch action.Type {
	case "startTracing":
		if action.StartOpts == nil {
			return fmt.Errorf("missing trace options")
		}
		return a.StartTracing(*action.StartOpts)
	case "gather":
		dest, err := a.PickGatherDestination()
		if err != nil {
			return err
		}
		if dest == "" {
			return nil
		}
		return a.StopTracingAndGather(dest)
	default:
		return fmt.Errorf("unknown pending action: %s", action.Type)
	}
}
