//go:build windows

package diag

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (g *Gatherer) exportEventLogs(failed *[]string) error {
	folder := g.state.TracePath32
	g.deleteOldLogFiles(folder)

	keyTime := getTracingKeyLastWriteTime()
	dpFrom := iso8601Time(keyTime.Add(-24 * time.Hour))
	systemFrom := iso8601Time(keyTime.Add(-7 * 24 * time.Hour))

	exports := []struct {
		outPath string
		args    []string
		label   string
	}{
		{
			outPath: filepath.Join(folder, "DPLogs.evtx"),
			label:   "DPLogs",
			args: []string{
				"epl",
				"/q:" + dpEventLogQuery(dpFrom),
				filepath.Join(folder, "DPLogs.evtx"),
			},
		},
		{
			outPath: filepath.Join(folder, "GroupPolicy.evtx"),
			label:   "GroupPolicy",
			args: []string{
				"epl",
				"Microsoft-Windows-GroupPolicy/Operational",
				filepath.Join(folder, "GroupPolicy.evtx"),
				"/q:" + fmt.Sprintf(`*[System[((EventID>=6000 and EventID<=6007) or (EventID>=6017 and EventID<=6299) or (EventID>=7000 and EventID<=7007) or (EventID>=7017 and EventID<=7299)) and TimeCreated[@SystemTime > '%s']]]`, systemFrom),
			},
		},
		{
			outPath: filepath.Join(folder, "System.evtx"),
			label:   "System",
			args: []string{
				"epl",
				"System",
				filepath.Join(folder, "System.evtx"),
				"/q:" + fmt.Sprintf(`*[System[(EventID=1014 or EventID=1112 or EventID=1085) and TimeCreated[@SystemTime > '%s']]]`, systemFrom),
			},
		},
	}

	for _, item := range exports {
		cmd := exec.Command("wevtutil", item.args...)
		cmd.SysProcAttr = syscallSysProcAttrHideWindow()
		if out, err := cmd.CombinedOutput(); err != nil {
			*failed = append(*failed, fmt.Sprintf("event log %s: %v (%s)", item.label, err, strings.TrimSpace(string(out))))
		}
	}
	return nil
}

func (g *Gatherer) deleteOldLogFiles(folder string) {
	for _, name := range []string{"DPLogs.evtx", "System.evtx", "GroupPolicy.evtx"} {
		_ = os.Remove(filepath.Join(folder, name))
	}
}

func dpEventLogQuery(from string) string {
	const template = `<QueryList><Query Id='0'><Select Path='DigitalPersona-Altus-Core/Operational'>*[System[TimeCreated[@SystemTime > '%s']]]</Select></Query><Query Id='1'><Select Path='DigitalPersona-Altus-Logon/Operational'>*[System[TimeCreated[@SystemTime > '%s']]]</Select></Query><Query Id='2'><Select Path='DigitalPersona-Altus-Password Manager/Operational'>*[System[TimeCreated[@SystemTime > '%s']]]</Select></Query><Query Id='3'><Select Path='DigitalPersona-Altus-Reporter/Operational'>*[System[TimeCreated[@SystemTime > '%s']]]</Select></Query></QueryList>`
	return fmt.Sprintf(template, from, from, from, from)
}

func iso8601Time(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}
