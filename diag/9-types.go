package diag

// TraceOptions mirrors UI settings passed when starting tracing.
type TraceOptions struct {
	AccumulateTraces   bool `json:"accumulateTraces"`
	EnableOTSTrace     bool `json:"enableOtsTrace"`
	Verbosity          uint `json:"verbosity"`
}

// TraceSettings is returned to the frontend on load.
type TraceSettings struct {
	IsTracing              bool   `json:"isTracing"`
	AccumulateTraces       bool   `json:"accumulateTraces"`
	EnableOTSTrace         bool   `json:"enableOtsTrace"`
	Verbosity              uint   `json:"verbosity"`
	TracePath              string `json:"tracePath"`
	PasswordManagerFound   bool   `json:"passwordManagerFound"`
	Is64BitWindows         bool   `json:"is64BitWindows"`
}

// GatherProgress reports ZIP collection progress.
type GatherProgress struct {
	Collected int `json:"collected"`
	Total     int `json:"total"`
}

// GatherResult is emitted when gathering completes.
type GatherResult struct {
	ZipPath         string   `json:"zipPath"`
	FailedFiles     []string `json:"failedFiles"`
	DeleteOnReboot  bool     `json:"deleteOnReboot"`
	CleanupMessage  string   `json:"cleanupMessage,omitempty"`
}
