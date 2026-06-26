//go:build windows

package diag

import (
	"fmt"
	"os/exec"
	"strings"
)

func createBootCleanupTask(taskArgs string) error {
	appPath, err := quotedAppPath()
	if err != nil {
		return err
	}
	// Match C++ TaskSchedulerStuff: PT2M boot delay, HID Global author (schtasks has no author field).
	tr := appPath + " " + taskArgs
	cmd := exec.Command("schtasks",
		"/Create",
		"/TN", TaskName,
		"/TR", tr,
		"/SC", "ONSTART",
		"/DELAY", "0002:00",
		"/RU", "LOCAL SERVICE",
		"/F",
	)
	cmd.SysProcAttr = syscallSysProcAttrHideWindow()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks create: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func deleteBootCleanupTask() error {
	cmd := exec.Command("schtasks", "/Delete", "/TN", TaskName, "/F")
	cmd.SysProcAttr = syscallSysProcAttrHideWindow()
	out, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(out), "ERROR: The system cannot find") {
		return fmt.Errorf("schtasks delete: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func scheduleCleanupIfNeeded(state *TraceState) error {
	if !isDeleteAtEnd() {
		return nil
	}
	args := CmdDeleteFiles + " " + state.quotedTracePaths()
	return createBootCleanupTask(args)
}
