//go:build windows

package diag

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed res/PowerShell.ps1
var powerShellScript []byte

func runPowerShellScript(state string, outputPath string, verbosity uint) error {
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "DP*.ps1")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(powerShellScript); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	psPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	args := []string{
		"-NonInteractive", "-WindowStyle", "Hidden", "-NoLogo", "-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", tmpPath,
		state, outputPath, fmt.Sprintf("%d", verbosity),
	}

	cmd := exec.Command(psPath, args...)
	attr := syscallSysProcAttrHideWindow()
	cmd.SysProcAttr = attr
	return cmd.Run()
}

func psOnEnabled(outputPath string, verbosity uint) error {
	return runPowerShellScript("start", outputPath, verbosity)
}

func psOnDisabled(outputPath string, verbosity uint) error {
	return runPowerShellScript("stop", outputPath, verbosity)
}
