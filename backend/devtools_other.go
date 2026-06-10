//go:build !windows

package backend

func (a *App) platformIsDevToolsOpen() bool { return false }

// platformCloseDevTools is a no-op on non-Windows platforms.
func (a *App) platformCloseDevTools() {}
