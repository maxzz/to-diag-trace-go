//go:build windows

package diag

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

//go:embed res/PowerShell.ps1
var powerShellScript []byte

const createNoWindow = 0x08000000

// runPowerShellScript mirrors PsTraceUtils::Execute in the original C++ tool:
// embed BINARY resource → temp .ps1 → powershell.exe -File ... → wait → delete temp.
func runPowerShellScript(state string, outputPath string, verbosity uint) error {
	tmpPath, err := createTempScriptPath()
	if err != nil {
		return err
	}

	if err := writeTempScript(tmpPath, powerShellScript); err != nil {
		os.Remove(tmpPath)
		return err
	}

	psPath, err := systemPowerShellPath()
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Same argument list as C++:
	// -NonInteractive -WindowStyle Hidden -NoLogo -NoProfile -ExecutionPolicy Bypass
	// -File "<tmp>" <state> "<outputPath>" <verbosity>
	cmd := exec.Command(
		psPath,
		"-NonInteractive",
		"-WindowStyle", "Hidden",
		"-NoLogo",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", tmpPath,
		state,
		outputPath,
		fmt.Sprintf("%d", verbosity),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}

	if err := cmd.Start(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("CreateProcess powershell: %w", err)
	}

	_ = cmd.Wait()
	_ = os.Remove(tmpPath)

	// C++ Execute does not propagate the child exit code; only setup failures return HRESULT.
	return nil
}

// createTempScriptPath matches GetTempFileName(..., "DP", ...) + ".ps1" from PsTraceUtils.cpp.
func createTempScriptPath() (string, error) {
	tempDir := os.TempDir()
	base, err := os.CreateTemp(tempDir, "DP")
	if err != nil {
		return "", err
	}
	basePath := base.Name()
	base.Close()
	if err := os.Remove(basePath); err != nil {
		return "", err
	}
	return basePath + ".ps1", nil
}

func writeTempScript(path string, content []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.Write(content)
	if err != nil {
		return err
	}
	if n != len(content) {
		return fmt.Errorf("short write writing PowerShell script (%d/%d bytes)", n, len(content))
	}
	return nil
}

// systemPowerShellPath returns %SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe
// via SHGetKnownFolderPath(FOLDERID_System), same as PsTraceUtils.cpp.
func systemPowerShellPath() (string, error) {
	system32, err := knownFolderSystem32()
	if err != nil {
		// Fallback when Known Folder API is unavailable.
		root := os.Getenv("SystemRoot")
		if root == "" {
			return "", err
		}
		return filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"), nil
	}
	return filepath.Join(system32, "WindowsPowerShell", "v1.0", "powershell.exe"), nil
}

// FOLDERID_System = {1AC14E77-02E7-4E5D-B745-2EB1AE5198B7}
var folderIDSystem = syscall.GUID{
	Data1: 0x1AC14E77,
	Data2: 0x02E7,
	Data3: 0x4E5D,
	Data4: [8]byte{0xB7, 0x45, 0x2E, 0xB1, 0xAE, 0x51, 0x98, 0xB7},
}

func knownFolderSystem32() (string, error) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shGetKnownFolderPath := shell32.NewProc("SHGetKnownFolderPath")

	const kfFlagDefault = 0
	var pathPtr uintptr
	ret, _, err := shGetKnownFolderPath.Call(
		uintptr(unsafe.Pointer(&folderIDSystem)),
		uintptr(kfFlagDefault),
		0,
		uintptr(unsafe.Pointer(&pathPtr)),
	)
	if ret != 0 {
		return "", fmt.Errorf("SHGetKnownFolderPath: %w", err)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(pathPtr))

	path := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(pathPtr)))
	if path == "" {
		return "", fmt.Errorf("SHGetKnownFolderPath returned empty path")
	}
	return path, nil
}

func psOnEnabled(outputPath string, verbosity uint) error {
	return runPowerShellScript("start", outputPath, verbosity)
}

func psOnDisabled(outputPath string, verbosity uint) error {
	return runPowerShellScript("stop", outputPath, verbosity)
}
