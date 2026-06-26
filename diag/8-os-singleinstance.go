//go:build windows

package diag

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const bringToFrontMessage = "016ED575-15D2-4a78-BFB1-EE0B1930F310"

var registeredBringMessage uint32

func init() {
	user32 := syscall.NewLazyDLL("user32.dll")
	registerWindowMessage := user32.NewProc("RegisterWindowMessageW")
	msg, _ := syscall.UTF16PtrFromString(bringToFrontMessage)
	ret, _, _ := registerWindowMessage.Call(uintptr(unsafe.Pointer(msg)))
	registeredBringMessage = uint32(ret)
}

func AcquireSingleInstance() (release func(), ok bool, err error) {
	name, _ := syscall.UTF16PtrFromString(MutexName)
	handle, _, callErr := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateMutexW").Call(
		0, 0, uintptr(unsafe.Pointer(name)),
	)
	if handle == 0 {
		return func() {}, false, callErr
	}
	lastErr := syscall.GetLastError()
	if lastErr == syscall.ERROR_ALREADY_EXISTS {
		syscall.CloseHandle(syscall.Handle(handle))
		return func() {}, false, nil
	}
	return func() {
		syscall.CloseHandle(syscall.Handle(handle))
	}, true, nil
}

// ActivateExistingInstance asks a running instance to come to the foreground.
func ActivateExistingInstance() {
	if registeredBringMessage != 0 {
		user32 := syscall.NewLazyDLL("user32.dll")
		broadcastSystemMessage := user32.NewProc("BroadcastSystemMessageW")
		const bsfIgnoreCurrentTask = 0x00000004
		const bsfPostMessage = 0x00000010
		const bsfForceIfHung = 0x00000020
		const bsmApplications = 0x00000008
		recipients := uint32(bsmApplications)
		_, _, _ = broadcastSystemMessage.Call(
			bsfIgnoreCurrentTask|bsfPostMessage|bsfForceIfHung,
			uintptr(unsafe.Pointer(&recipients)),
			uintptr(registeredBringMessage),
			0,
			0,
		)
	}
	activateExistingWindowByTitle()
}

func activateExistingWindowByTitle() {
	const windowTitle = "DigitalPersona Diagnostic Tool"
	title, _ := syscall.UTF16PtrFromString(windowTitle)
	user32 := syscall.NewLazyDLL("user32.dll")
	findWindow := user32.NewProc("FindWindowW")
	showWindow := user32.NewProc("ShowWindow")
	setForegroundWindow := user32.NewProc("SetForegroundWindow")

	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd == 0 {
		return
	}
	const swRestore = 9
	_, _, _ = showWindow.Call(hwnd, swRestore)
	_, _, _ = setForegroundWindow.Call(hwnd)
}

func getTracingKeyLastWriteTime() time.Time {
	var best time.Time
	for _, wow64 := range []uint32{registry.WOW64_32KEY, registry.WOW64_64KEY} {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, TracingKey, registry.READ|wow64)
		if err != nil {
			continue
		}
		info, err := key.Stat()
		key.Close()
		if err != nil {
			continue
		}
		t := info.ModTime()
		if best.IsZero() || t.Before(best) {
			best = t
		}
	}
	if best.IsZero() {
		return time.Now().UTC()
	}
	return best
}
