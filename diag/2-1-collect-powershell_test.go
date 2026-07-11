//go:build windows

package diag

import (
	"os"
	"testing"
)

func TestEmbeddedPowerShellScriptMatchesResourceFile(t *testing.T) {
	onDisk, err := os.ReadFile("res/PowerShell.ps1")
	if err != nil {
		t.Skip("res/PowerShell.ps1 not on disk")
	}
	if len(onDisk) != len(powerShellScript) {
		t.Fatalf("embedded script length %d != file length %d", len(powerShellScript), len(onDisk))
	}
	for i := range onDisk {
		if onDisk[i] != powerShellScript[i] {
			t.Fatalf("embedded script differs from res/PowerShell.ps1 at byte %d", i)
		}
	}
}

func TestSystemPowerShellPath(t *testing.T) {
	path, err := systemPowerShellPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("powershell.exe not found at %q: %v", path, err)
	}
}
