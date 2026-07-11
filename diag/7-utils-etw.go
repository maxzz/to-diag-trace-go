//go:build windows

package diag

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	eventTraceControlQuery = 0
	eventTraceControlStop  = 1
	wnodeFlagTracedGUID    = 0x00020000
	eventTraceFileModeCircular        = 0x00000200
	eventTraceDelayOpenFileMode       = 0x00001000
	eventTraceAddHeaderMode           = 0x01000000
	errorWMIInstanceNotFound          = 4201
)

var (
	modAdvapi32    = syscall.NewLazyDLL("advapi32.dll")
	procStartTrace = modAdvapi32.NewProc("StartTraceW")
	procControlTrace = modAdvapi32.NewProc("ControlTraceW")
	procEnableTrace  = modAdvapi32.NewProc("EnableTrace")
)

var (
	dpTraceGUID  = syscall.GUID{Data1: 0x653c88e1, Data2: 0x427c, Data3: 0x5bd6, Data4: [8]byte{0x45, 0x11, 0x3c, 0x9a, 0x8b, 0x19, 0x44, 0xa4}}
	chromeGUID   = syscall.GUID{Data1: 0x7FE69228, Data2: 0x633E, Data3: 0x4F06, Data4: [8]byte{0x80, 0xC1, 0x52, 0x7F, 0xEA, 0x23, 0xE3, 0xA7}}
)

type eventTraceProperties struct {
	Wnode               wnodeHeader
	BufferSize          uint32
	MinimumBuffers      uint32
	MaximumBuffers      uint32
	MaximumFileSize     uint32
	LogFileMode         uint32
	FlushTimer          uint32
	EnableFlags         uint32
	AgeLimit            int32
	NumberOfBuffers     uint32
	FreeBuffers         uint32
	EventsLost          uint32
	BuffersWritten      uint32
	LogBuffersLost      uint32
	RealTimeBuffersLost uint32
	LoggerThreadId      uintptr
	LogFileNameOffset   uint32
	LoggerNameOffset    uint32
}

type wnodeHeader struct {
	BufferSize  uint32
	ProviderId  uint32
	HistoricalContext uint64
	TimeStamp   int64
	Guid        syscall.GUID
	ClientContext uint32
	Flags       uint32
}

func etwIsEnabled(sessionName string) bool {
	props, err := allocateTraceProperties("")
	if err != nil {
		return false
	}
	defer props.free()

	name, _ := syscall.UTF16PtrFromString(sessionName)
	ret, _, _ := procControlTrace.Call(
		0,
		uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(props.ptr)),
		eventTraceControlQuery,
	)
	return ret == 0
}

func etwDisable(sessionName string) error {
	props, err := allocateTraceProperties("")
	if err != nil {
		return err
	}
	defer props.free()

	name, _ := syscall.UTF16PtrFromString(sessionName)
	_, _, _ = procControlTrace.Call(
		0,
		uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(props.ptr)),
		eventTraceControlStop,
	)

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\WMI\Autologger`, registry.SET_VALUE|registry.WOW64_64KEY)
	if err == nil {
		_ = registry.DeleteKey(key, sessionName)
		key.Close()
	}
	return nil
}

func etwEnable(sessionName, logFilePath string) error {
	if err := etwDisable(sessionName); err != nil {
		// continue
	}

	const (
		clockType       = 1
		bufferSize      = 256
		flushTimer      = 60
		maxBuffers      = 256
		minBuffers      = 16
		maxFileSize     = 200
	)
	mode := uint32(eventTraceFileModeCircular | eventTraceDelayOpenFileMode | eventTraceAddHeaderMode)

	props, err := allocateTraceProperties(logFilePath)
	if err != nil {
		return err
	}
	defer props.free()

	props.setParams(clockType, bufferSize, flushTimer, maxBuffers, minBuffers, maxFileSize, mode)

	sessionNamePtr, _ := syscall.UTF16PtrFromString(sessionName)
	var handle uintptr
	ret, _, _ := procStartTrace.Call(
		uintptr(unsafe.Pointer(&handle)),
		uintptr(unsafe.Pointer(sessionNamePtr)),
		uintptr(unsafe.Pointer(props.ptr)),
	)
	if ret != 0 {
		return fmt.Errorf("StartTrace failed: %d", ret)
	}

	if err := enableTraceProvider(&dpTraceGUID, handle, 0); err != nil {
		_ = etwDisable(sessionName)
		return err
	}
	_ = enableTraceProvider(&chromeGUID, handle, 4)

	return configureAutologger(sessionName, logFilePath, clockType, bufferSize, flushTimer, maxBuffers, minBuffers, maxFileSize, mode)
}

func enableTraceProvider(guid *syscall.GUID, sessionHandle uintptr, level uint8) error {
	ret, _, _ := procEnableTrace.Call(
		1,
		0,
		uintptr(level),
		uintptr(unsafe.Pointer(guid)),
		sessionHandle,
	)
	if ret != 0 {
		return fmt.Errorf("EnableTrace failed: %d", ret)
	}
	return nil
}

func configureAutologger(sessionName, logFilePath string, clockType, bufferSize, flushTimer, maxBuffers, minBuffers, maxFileSize uint32, mode uint32) error {
	sessionPath := `SYSTEM\CurrentControlSet\Control\WMI\Autologger\` + sessionName
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, sessionPath, registry.SET_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return err
	}
	_ = key.SetDWordValue("BufferSize", bufferSize)
	_ = key.SetDWordValue("FlushTimer", flushTimer)
	_ = key.SetDWordValue("MaximumBuffers", maxBuffers)
	_ = key.SetDWordValue("MinimumBuffers", minBuffers)
	_ = key.SetDWordValue("ClockType", clockType)
	_ = key.SetDWordValue("MaxFileSize", maxFileSize)
	_ = key.SetDWordValue("LogFileMode", mode)
	_ = key.SetStringValue("Guid", "{00000000-0000-0000-0000-000000000000}")
	_ = key.SetStringValue("FileName", logFilePath)
	_ = key.SetDWordValue("Start", 1)
	_ = key.SetDWordValue("FileMax", 6)
	_ = key.SetDWordValue("Status", 0)
	key.Close()

	providerPath := sessionPath + `\{653C88E1-427C-5BD6-4511-3C9A8B1944A4}`
	pkey, _, err := registry.CreateKey(registry.LOCAL_MACHINE, providerPath, registry.SET_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return err
	}
	_ = pkey.SetDWordValue("Enabled", 1)
	_ = pkey.SetDWordValue("EnableLevel", 0)
	_ = pkey.SetDWordValue("EnableProperty", 0)
	_ = pkey.SetQWordValue("MatchAllKeyword", 0)
	_ = pkey.SetQWordValue("MatchAnyKeyword", 0)
	_ = pkey.SetDWordValue("Status", 0)
	return pkey.Close()
}

func etwLogPathForState(state *TraceState) string {
	dir := state.TracePath64
	if !isDirExist(dir) {
		dir = state.TracePath32
	}
	return filepath.Join(dir, ETWLogFileName)
}

type tracePropertiesBuffer struct {
	ptr  *eventTraceProperties
	data []byte
}

func allocateTraceProperties(logFileName string) (*tracePropertiesBuffer, error) {
	const maxSessionName = 256
	size := int(unsafe.Sizeof(eventTraceProperties{})) + maxSessionName + syscall.MAX_PATH
	data := make([]byte, size)
	ptr := (*eventTraceProperties)(unsafe.Pointer(&data[0]))
	ptr.Wnode.BufferSize = uint32(size)
	ptr.Wnode.Flags = wnodeFlagTracedGUID
	ptr.LoggerNameOffset = uint32(unsafe.Sizeof(eventTraceProperties{}))
	ptr.LogFileNameOffset = ptr.LoggerNameOffset + maxSessionName

	if logFileName != "" {
		logOffset := ptr.LogFileNameOffset
		dest := (*[syscall.MAX_PATH]uint16)(unsafe.Pointer(&data[logOffset]))
		copy(dest[:], syscall.StringToUTF16(logFileName))
	}

	return &tracePropertiesBuffer{ptr: ptr, data: data}, nil
}

func (b *tracePropertiesBuffer) setParams(clockType, bufferSize, flushTimer, maxBuffers, minBuffers, maxFileSize, mode uint32) {
	b.ptr.Wnode.ClientContext = clockType
	b.ptr.BufferSize = bufferSize
	b.ptr.FlushTimer = flushTimer
	b.ptr.MaximumBuffers = maxBuffers
	b.ptr.MinimumBuffers = minBuffers
	b.ptr.MaximumFileSize = maxFileSize
	b.ptr.LogFileMode = mode
}

func (b *tracePropertiesBuffer) free() {}
