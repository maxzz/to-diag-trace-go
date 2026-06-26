//go:build windows

package diag

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type otsModule struct {
	name  string
	value uint32
}

type otsSettings struct {
	modules []otsModule
}

func newOTSSettings() *otsSettings {
	return &otsSettings{
		modules: []otsModule{
			{name: "ots_dpfbview", value: 0},
			{name: "ots_dpofeedb", value: 0},
			{name: "ots_dpffcli", value: 0},
			{name: "ots_dpfillin", value: 0},
			{name: "ots_dpocache", value: 0},
		},
	}
}

func (o *otsSettings) setTraceFlags(enabled bool) {
	flags := []uint32{0, 0, 0, 0, 0}
	if enabled {
		flags = []uint32{0x1fff003f, 0x000003ff, 0x000007ff, 0x003f003f, 0x000001ff}
	}
	for i := range o.modules {
		o.modules[i].value = flags[i]
	}
}

type tracingKey struct {
	key registry.Key
	ots *otsSettings
}

func openTracingKey(access uint32) (*tracingKey, error) {
	ots := newOTSSettings()
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, TracingKey, access)
	if err == registry.ErrNotExist && (access&registry.SET_VALUE) != 0 {
		key, _, err = registry.CreateKey(registry.LOCAL_MACHINE, TracingKey, access)
	}
	if err != nil {
		return nil, err
	}
	return &tracingKey{key: key, ots: ots}, nil
}

func (t *tracingKey) Close() error {
	return t.key.Close()
}

func (t *tracingKey) isDPTraceEnabled() (bool, bool) {
	val, _, err := t.key.GetIntegerValue("DoTrace")
	if err != nil {
		return false, false
	}
	return val == 1, true
}

func (t *tracingKey) isTraceToNewFile() (bool, bool) {
	val, _, err := t.key.GetIntegerValue("NewFileAlways")
	if err != nil {
		return false, false
	}
	return val == 1, true
}

func (t *tracingKey) isOTSTraceEnabled() (bool, bool) {
	val, _, err := t.key.GetIntegerValue("OtsTrace")
	if err != nil || val != 1 {
		return false, false
	}
	for _, m := range t.ots.modules {
		v, _, err := t.key.GetIntegerValue(m.name)
		if err != nil || v == 0 {
			return false, false
		}
	}
	return true, true
}

func (t *tracingKey) getVerbosity() (uint, bool) {
	val, _, err := t.key.GetIntegerValue("Verbosity")
	if err != nil {
		return 0, false
	}
	return uint(val), true
}

func (t *tracingKey) getTracePath() (string, bool) {
	val, _, err := t.key.GetStringValue("TracePath")
	if err != nil || val == "" {
		return "", false
	}
	return val, true
}

func (t *tracingKey) getDeleteAtEnd() bool {
	val, _, err := t.key.GetIntegerValue("DeleteAtEnd")
	return err == nil && val != 0
}

func (t *tracingKey) setDoTrace(enabled bool) error {
	var v uint32
	if enabled {
		v = 1
	}
	return t.key.SetDWordValue("DoTrace", v)
}

func (t *tracingKey) setTraceToFile(enabled bool) error {
	var v uint32
	if enabled {
		v = 1
	}
	return t.key.SetDWordValue("TraceToFile", v)
}

func (t *tracingKey) setNewFileAlways(enabled bool) error {
	var v uint32
	if enabled {
		v = 1
	}
	return t.key.SetDWordValue("NewFileAlways", v)
}

func (t *tracingKey) setVerbosity(v uint) error {
	return t.key.SetDWordValue("Verbosity", uint32(v))
}

func (t *tracingKey) setTracePath(path string) error {
	return t.key.SetStringValue("TracePath", path)
}

func (t *tracingKey) setDeleteAtEnd() error {
	return t.key.SetDWordValue("DeleteAtEnd", 1)
}

func (t *tracingKey) setOTSTrace(enabled bool) error {
	t.ots.setTraceFlags(enabled)
	if err := t.key.SetDWordValue("OtsTrace", boolToDWord(enabled)); err != nil {
		return err
	}
	for _, m := range t.ots.modules {
		if err := t.key.SetDWordValue(m.name, m.value); err != nil {
			return err
		}
	}
	return nil
}

func boolToDWord(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func deleteTracingKey(wow64 uint32) error {
	parent, err := registry.OpenKey(registry.LOCAL_MACHINE, DPKey, registry.SET_VALUE|wow64)
	if err != nil {
		return err
	}
	defer parent.Close()
	return registry.DeleteKey(parent, "Tracing")
}

func readTraceStateFromRegistry() (*TraceState, error) {
	state := &TraceState{
		Is64Bit:       is64BitWindows(),
		TraceToFile:   true,
		Verbosity:     DefaultVerbosity,
		PasswordManagerFound: isProductPasswordManager(),
	}

	if state.Is64Bit {
		read64State(state)
	} else {
		read32State(state)
	}

	if state.TracePath32 == "" && state.TracePath64 == "" {
		def, err := defaultTracePath()
		if err != nil {
			return nil, err
		}
		state.TracePath32 = def
		state.TracePath64 = def
	} else if state.TracePath32 == "" {
		state.TracePath32 = state.TracePath64
	} else if state.TracePath64 == "" {
		state.TracePath64 = state.TracePath32
	}

	return state, nil
}

func read64State(state *TraceState) {
	k32, err := openTracingKey(registry.READ | registry.WOW64_32KEY)
	if err == nil {
		applyKeyToState(k32, state, true)
		k32.Close()
	}
	k64, err := openTracingKey(registry.READ | registry.WOW64_64KEY)
	if err == nil {
		applyKeyToState(k64, state, false)
		if p, ok := k64.getTracePath(); ok {
			state.TracePath64 = p
		}
		k64.Close()
	}
}

func read32State(state *TraceState) {
	k, err := openTracingKey(registry.READ)
	if err != nil {
		return
	}
	defer k.Close()
	applyKeyToState(k, state, true)
	if p, ok := k.getTracePath(); ok {
		state.TracePath32 = p
		state.TracePath64 = p
	}
}

func applyKeyToState(k *tracingKey, state *TraceState, is32Path bool) {
	if enabled, ok := k.isDPTraceEnabled(); ok && enabled {
		state.DoTrace = true
	}
	if newFile, ok := k.isTraceToNewFile(); ok && !newFile {
		state.NewFileAlways = false
	}
	if ots, ok := k.isOTSTraceEnabled(); ok && ots {
		state.DoTraceOTS = true
	}
	if v, ok := k.getVerbosity(); ok && v > state.Verbosity {
		state.Verbosity = v
	}
	if is32Path {
		if p, ok := k.getTracePath(); ok {
			state.TracePath32 = p
		}
	}
}

func writeTraceStateToRegistry(state *TraceState) error {
	if err := writeKey(state, registry.WOW64_32KEY, state.TracePath32); err != nil {
		return fmt.Errorf("32-bit registry: %w", err)
	}
	if state.Is64Bit {
		if err := writeKey(state, registry.WOW64_64KEY, state.TracePath64); err != nil {
			return fmt.Errorf("64-bit registry: %w", err)
		}
	}
	k, err := openTracingKey(registry.SET_VALUE)
	if err == nil {
		_ = k.setDeleteAtEnd()
		k.Close()
	}
	return nil
}

func writeKey(state *TraceState, wow64 uint32, tracePath string) error {
	k, err := openTracingKey(registry.READ | registry.SET_VALUE | wow64)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.setTracePath(tracePath); err != nil {
		return err
	}
	if err := k.setDoTrace(state.DoTrace); err != nil {
		return err
	}
	if err := k.setOTSTrace(state.DoTraceOTS); err != nil {
		return err
	}
	if err := k.setTraceToFile(state.TraceToFile); err != nil {
		return err
	}
	if err := k.setNewFileAlways(state.NewFileAlways); err != nil {
		return err
	}
	return k.setVerbosity(state.Verbosity)
}

func isDeleteAtEnd() bool {
	k, err := openTracingKey(registry.READ)
	if err != nil {
		return false
	}
	defer k.Close()
	return k.getDeleteAtEnd()
}

func deleteRegistryTracingKeys(is64 bool) {
	_ = deleteTracingKey(registry.WOW64_32KEY)
	if is64 {
		_ = deleteTracingKey(registry.WOW64_64KEY)
	}
}
