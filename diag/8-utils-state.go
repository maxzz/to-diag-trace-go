//go:build windows

package diag

import (
	"os"
	"path/filepath"
)

// TraceState holds tracing configuration mirrored from traceState_t.
type TraceState struct {
	Is64Bit              bool
	DoTrace              bool
	DoTraceOTS           bool
	TraceToFile          bool
	NewFileAlways        bool
	Verbosity            uint
	PasswordManagerFound bool
	TracePath32          string
	TracePath64          string
}

func (s *TraceState) IsTracing() bool {
	return s.DoTrace
}

func (s *TraceState) TraceFolder32() string {
	return s.TracePath32
}

func (s *TraceState) TraceFolder64() string {
	return s.TracePath64
}

func (s *TraceState) pathsDifferent() bool {
	return s.TracePath32 != s.TracePath64
}

func (s *TraceState) createTraceFolders() error {
	if err := recurseCreateDir(s.TracePath32); err != nil {
		return err
	}
	if s.pathsDifferent() {
		return recurseCreateDir(s.TracePath64)
	}
	return nil
}

func (s *TraceState) quotedTracePaths() string {
	var parts []string
	if isDirExist(s.TracePath64) {
		parts = append(parts, `"`+s.TracePath64+`"`)
	}
	if s.pathsDifferent() && isDirExist(s.TracePath32) {
		parts = append(parts, `"`+s.TracePath32+`"`)
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

func is64BitWindows() bool {
	return os.Getenv("PROCESSOR_ARCHITECTURE") == "AMD64" || os.Getenv("PROCESSOR_ARCHITEW6432") == "AMD64"
}

func quotedAppPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return "", err
	}
	return `"` + exe + `"`, nil
}
