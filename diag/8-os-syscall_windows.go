//go:build windows

package diag

import "syscall"

func syscallSysProcAttrHideWindow() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}
