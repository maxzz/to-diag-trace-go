//go:build !windows

package diag

import "errors"
import "time"

var ErrUnsupported = errors.New("diagnostic tool is only supported on Windows")

type Service struct{}

func NewService() (*Service, error) { return &Service{}, nil }

func (s *Service) GetTraceSettings() (TraceSettings, error) {
	return TraceSettings{}, ErrUnsupported
}

func (s *Service) StartTracing(opts TraceOptions) error { return ErrUnsupported }

func (s *Service) StopTracing() error { return ErrUnsupported }

func (s *Service) IsTracingActive() (bool, error) { return false, ErrUnsupported }

func (s *Service) Gather(destZip string, onProgress func(GatherProgress)) (GatherResult, error) {
	return GatherResult{}, ErrUnsupported
}

func (s *Service) CancelGather() {}

func (s *Service) IsElevated() bool { return IsElevated() }

func (s *Service) RelaunchElevated(extraArgs string) error { return RelaunchElevated(extraArgs) }

func IsElevated() bool { return false }

func RelaunchElevated(extraArgs string) error { return ErrUnsupported }

func HandleDeleteFilesCLI(args []string) error { return ErrUnsupported }

func AcquireSingleInstance() (release func(), ok bool, err error) {
	return func() {}, true, nil
}

func ActivateExistingInstance() {}

func getTracingKeyLastWriteTime() time.Time { return time.Now().UTC() }
