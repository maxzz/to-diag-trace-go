//go:build windows

package diag

const (
	DefaultVerbosity = 4

	TracingKey        = `SOFTWARE\DigitalPersona\Tracing`
	DPKey             = `SOFTWARE\DigitalPersona`
	DPGroupPolicyKey  = `SOFTWARE\Policies\DigitalPersona\Altus`
	DPProductsPM      = `SOFTWARE\DigitalPersona\Products\Password Manager`

	RunKeyPath        = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	RunKeyValueName   = "DigitalPersona Trace Utility"

	ETWSessionName    = "DpDiagnosticTool"
	ETWLogFileName    = "dptrace.etl"

	TaskName          = "DP Trace Cleanup Task"
	TaskDelay         = "PT30S"

	MutexName         = "906BC945-EFFB-4d10-9257-D65742B131D0"

	CmdDeleteFiles      = "/DeleteFiles"
	CmdAuto             = "/Auto"
)
