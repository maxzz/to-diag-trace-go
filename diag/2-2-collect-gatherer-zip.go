//go:build windows

package diag

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type folderPattern struct {
	path    string
	pattern string
}

type Gatherer struct {
	state      *TraceState
	destZip    string
	cancelled  bool
	mu         sync.Mutex
	onProgress func(GatherProgress)
}

func NewGatherer(state *TraceState, destZip string, onProgress func(GatherProgress)) *Gatherer {
	return &Gatherer{state: state, destZip: destZip, onProgress: onProgress}
}

func (g *Gatherer) Cancel() {
	g.mu.Lock()
	g.cancelled = true
	g.mu.Unlock()
}

func (g *Gatherer) isCancelled() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.cancelled
}

func (g *Gatherer) Run() (GatherResult, error) {
	result := GatherResult{ZipPath: g.destZip}
	var failedFiles []string

	if err := g.exportEventLogs(&failedFiles); err != nil {
		failedFiles = append(failedFiles, err.Error())
	}

	dirs := g.collectDirectories()
	total := 0
	for _, d := range dirs {
		total += g.countFiles(d.path, d.pattern)
	}
	g.report(0, total)

	zf, err := os.Create(g.destZip)
	if err != nil {
		return result, fmt.Errorf("creating zip: %w", err)
	}
	zw := zip.NewWriter(zf)
	collected := 0

	for _, d := range dirs {
		if g.isCancelled() {
			zw.Close()
			zf.Close()
			os.Remove(g.destZip)
			return result, fmt.Errorf("gather cancelled")
		}
		errs := g.addMatchingFiles(zw, d.path, d.pattern, func() {
			collected++
			g.report(collected, total)
		})
		failedFiles = append(failedFiles, errs...)
	}

	if len(failedFiles) > 0 {
		missed := "This file contains list of trace files that Zip was not able to put into the archive.\r\n"
		missed += strings.Repeat("=", 100) + "\r\n\r\n"
		missed += strings.Join(failedFiles, "\r\n")
		w, err := zw.Create("__MissedFiles__.txt")
		if err == nil {
			_, _ = w.Write([]byte(missed))
		}
	}

	if err := zw.Close(); err != nil {
		zf.Close()
		os.Remove(g.destZip)
		return result, err
	}
	if err := zf.Close(); err != nil {
		os.Remove(g.destZip)
		return result, err
	}

	result.FailedFiles = failedFiles
	return result, nil
}

func (g *Gatherer) report(collected, total int) {
	if g.onProgress != nil {
		g.onProgress(GatherProgress{Collected: collected, Total: total})
	}
}

func (g *Gatherer) collectDirectories() []folderPattern {
	var dirs []folderPattern
	addTracePatterns := func(base string) {
		patterns := []string{
			"*.trc?", "*.log", "*.evtx", "*.etl",
			"*.001", "*.002", "*.003", "*.004", "*.005", "*.006",
			"*.007", "*.008", "*.009", "*.010",
			"*.txt", "*.json", "*.yaml", "*.config", "*.reg",
		}
		for _, p := range patterns {
			dirs = append(dirs, folderPattern{path: base, pattern: p})
		}
	}

	if isDirExist(g.state.TracePath32) {
		addTracePatterns(g.state.TracePath32)
	}
	if g.state.Is64Bit && g.state.pathsDifferent() && isDirExist(g.state.TracePath64) {
		addTracePatterns(g.state.TracePath64)
	}

	if g.state.DoTraceOTS {
		if tpl := getOTSTemplatePath(); tpl != "" && isDirExist(tpl) {
			dirs = append(dirs, folderPattern{path: tpl, pattern: "*.dpm"})
		}
	}

	if dpPath := getProductDPPath(); dpPath != "" {
		pluginPath := filepath.Join(dpPath, "PluginData")
		if isDirExist(pluginPath) {
			dirs = append(dirs, folderPattern{path: pluginPath, pattern: "*.*"})
		}
	}

	return dirs
}

func addFileToZipViaTemp(zw *zip.Writer, name, path string) error {
	tmp, err := os.CreateTemp(os.TempDir(), "dp-trace-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	dst.Close()
	return addFileToZip(zw, name, tmpPath)
}

func (g *Gatherer) countFiles(folder, pattern string) int {
	matches, err := filepath.Glob(filepath.Join(folder, pattern))
	if err != nil {
		return 0
	}
	count := 0
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && !info.IsDir() {
			count++
		}
	}
	return count
}

func (g *Gatherer) addMatchingFiles(zw *zip.Writer, folder, pattern string, onFile func()) []string {
	var failed []string
	matches, err := filepath.Glob(filepath.Join(folder, pattern))
	if err != nil {
		return []string{err.Error()}
	}
	for _, filePath := range matches {
		if g.isCancelled() {
			break
		}
		info, err := os.Stat(filePath)
		if err != nil || info.IsDir() {
			continue
		}
		rel, err := filepath.Rel(folder, filePath)
		if err != nil {
			rel = filepath.Base(filePath)
		}
		zipName := filepath.Join(filepath.Base(folder), rel)
		zipName = strings.ReplaceAll(zipName, `\`, `/`)

		if err := addFileToZip(zw, zipName, filePath); err != nil {
			if err2 := addFileToZipViaTemp(zw, zipName, filePath); err2 != nil {
				failed = append(failed, filePath+": "+err.Error())
			} else {
				onFile()
			}
		} else {
			onFile()
		}
	}
	return failed
}

func addFileToZip(zw *zip.Writer, name, path string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, src)
	return err
}
