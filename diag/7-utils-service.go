//go:build windows

package diag

import (
	"fmt"
	"sync"
)
type Service struct {
	state         *TraceState
	gatherer      *Gatherer
	mu            sync.Mutex
	tracingActive bool
}

func NewService() (*Service, error) {
	state, err := readTraceStateFromRegistry()
	if err != nil {
		return nil, err
	}
	return &Service{state: state, tracingActive: state.IsTracing()}, nil
}

func (s *Service) refreshState() error {
	state, err := readTraceStateFromRegistry()
	if err != nil {
		return err
	}
	s.state = state
	return nil
}

func (s *Service) GetTraceSettings() (TraceSettings, error) {
	if err := s.refreshState(); err != nil {
		return TraceSettings{}, err
	}
	return TraceSettings{
		IsTracing:            s.tracingActive,
		AccumulateTraces:     !s.state.NewFileAlways,
		EnableOTSTrace:       s.state.DoTraceOTS,
		Verbosity:            s.state.Verbosity,
		TracePath:            s.state.TracePath32,
		PasswordManagerFound: s.state.PasswordManagerFound,
		Is64BitWindows:       s.state.Is64Bit,
	}, nil
}

func (s *Service) IsTracingActive() (bool, error) {
	if err := s.refreshState(); err != nil {
		return false, err
	}
	return s.tracingActive, nil
}

func (s *Service) StartTracing(opts TraceOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.refreshState(); err != nil {
		return err
	}

	s.state.DoTrace = true
	s.state.NewFileAlways = !opts.AccumulateTraces
	s.state.DoTraceOTS = opts.EnableOTSTrace && s.state.PasswordManagerFound
	s.state.Verbosity = opts.Verbosity
	if s.state.Verbosity == 0 {
		s.state.Verbosity = DefaultVerbosity
	}
	s.state.TraceToFile = true

	if err := s.state.createTraceFolders(); err != nil {
		return fmt.Errorf("create trace folder: %w", err)
	}
	if err := writeTraceStateToRegistry(s.state); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}
	if err := addToRunKey(); err != nil {
		return fmt.Errorf("run key: %w", err)
	}

	logPath := etwLogPathForState(s.state)
	if err := etwEnable(ETWSessionName, logPath); err != nil {
		return fmt.Errorf("ETW: %w", err)
	}
	if err := psOnEnabled(s.state.TracePath64, s.state.Verbosity); err != nil {
		return fmt.Errorf("PowerShell start: %w", err)
	}

	s.tracingActive = true
	return nil
}

func (s *Service) StopTracing() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.refreshState(); err != nil {
		return err
	}

	s.state.DoTrace = false
	_ = etwDisable(ETWSessionName)
	if err := psOnDisabled(s.state.TracePath64, s.state.Verbosity); err != nil {
		return fmt.Errorf("PowerShell stop: %w", err)
	}
	s.tracingActive = false
	return nil
}

func (s *Service) Gather(destZip string, onProgress func(GatherProgress)) (GatherResult, error) {
	s.mu.Lock()
	if err := s.refreshState(); err != nil {
		s.mu.Unlock()
		return GatherResult{}, err
	}
	stateCopy := *s.state
	s.mu.Unlock()

	g := NewGatherer(&stateCopy, destZip, onProgress)
	s.mu.Lock()
	s.gatherer = g
	s.mu.Unlock()

	result, err := g.Run()

	s.mu.Lock()
	s.gatherer = nil
	s.mu.Unlock()

	if err != nil {
		return result, err
	}

	deleteFromRunKey()
	openExplorerForFile(destZip)

	if isDeleteAtEnd() {
		deleteRegistryTracingKeys(stateCopy.Is64Bit)
		if err := scheduleCleanupIfNeeded(&stateCopy); err != nil {
			result.CleanupMessage = "Failed to schedule cleanup task: " + err.Error()
		} else {
			result.DeleteOnReboot = true
			result.CleanupMessage = "Trace files will be deleted on the next reboot."
		}
	} else {
		_ = deleteBootCleanupTask()
		result.CleanupMessage = "Trace registry settings were left in place. Remove them manually when finished."
	}

	return result, nil
}

func (s *Service) CancelGather() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gatherer != nil {
		s.gatherer.Cancel()
	}
}

func (s *Service) IsElevated() bool {
	return IsElevated()
}

func (s *Service) RelaunchElevated(extraArgs string) error {
	return RelaunchElevated(extraArgs)
}

func HandleDeleteFilesCLI(args []string) error {
	cmdLine := stringsJoinArgs(args)
	if !stringsHasPrefix(cmdLine, CmdDeleteFiles) {
		return fmt.Errorf("unexpected delete command")
	}

	svc, err := NewService()
	if err != nil {
		return err
	}
	if svc.state.IsTracing() {
		return nil
	}

	path32, path64 := parseDeletePaths(cmdLine)
	if deleteTraceFiles(path32, path64) {
		deleteFromRunKey()
		_ = deleteBootCleanupTask()
	}
	return nil
}
