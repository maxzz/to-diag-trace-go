//go:build windows

package diag

import (
	"os"
	"path/filepath"
	"strings"
)

func addTrailingBackslash(path string) string {
	if path == "" {
		return path
	}
	if strings.HasSuffix(path, `\`) {
		return path
	}
	return path + `\`
}

func isDirExist(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func defaultTracePath() (string, error) {
	programData := os.Getenv("ProgramData")
	if programData == "" {
		return "", os.ErrNotExist
	}
	return filepath.Join(programData, "DigitalPersona", "Tracing"), nil
}

func recurseCreateDir(path string) error {
	if path == "" {
		return os.ErrInvalid
	}
	if isDirExist(path) {
		return nil
	}

	// Create parent directories first.
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	return nil
}

func openDestinationFolder(filename string) {
	openExplorerForFile(filename)
}
