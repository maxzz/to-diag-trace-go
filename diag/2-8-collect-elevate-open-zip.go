//go:build windows

package diag

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

func IsElevated() bool {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, RunKeyPath, registry.SET_VALUE)
	if err != nil {
		return false
	}
	key.Close()
	return true
}

func RelaunchElevated(extraArgs string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	var params *uint16
	if extraArgs != "" {
		params, _ = syscall.UTF16PtrFromString(extraArgs)
	}
	ret, _, _ := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW").Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)),
		0,
		1, // SW_SHOWNORMAL
	)
	if ret <= 32 {
		return os.ErrPermission
	}
	return nil
}

func addToRunKey() error {
	appPath, err := quotedAppPath()
	if err != nil {
		return err
	}
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, RunKeyPath, registry.SET_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(RunKeyValueName, appPath+" "+CmdAuto)
}

func deleteFromRunKey() {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, RunKeyPath, registry.SET_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return
	}
	defer key.Close()
	_ = key.DeleteValue(RunKeyValueName)
}

func deleteTraceFiles(path32, path64 string) bool {
	paths := []string{path32, path64}
	ok := true
	for _, p := range paths {
		if p == "" || !isDirExist(p) {
			continue
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			ok = false
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if err := os.Remove(filepath.Join(p, e.Name())); err != nil {
				ok = false
			}
		}
		if err := os.Remove(p); err != nil {
			ok = false
		}
	}
	return ok
}

func openExplorerForFile(path string) {
	exec.Command("explorer.exe", "/select,", path).Start()
}
