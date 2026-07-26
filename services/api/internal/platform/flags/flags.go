// Package flags provides the transport-neutral feature-flag and kill-switch
// kernel. It deliberately starts every capability disabled and publishes
// immutable snapshots so request paths can evaluate flags without locks.
package flags

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Flag is a canonical product capability that can be released independently.
type Flag string

const (
	FlagSow      Flag = "sow"
	FlagFires    Flag = "fires"
	FlagAI       Flag = "ai"
	FlagPayments Flag = "payments"
	FlagGate     Flag = "gate"
)

var canonicalFlags = [...]Flag{
	FlagSow,
	FlagFires,
	FlagAI,
	FlagPayments,
	FlagGate,
}

// Source identifies the configuration layer that produced a decision.
type Source string

const (
	SourceDefault         Source = "default"
	SourceEnvironment     Source = "environment"
	SourceRuntime         Source = "runtime"
	SourceEnvironmentKill Source = "environment_kill_switch"
	SourceRuntimeKill     Source = "runtime_kill_switch"
	SourceUnknown         Source = "unknown"
)

var (
	ErrUnknownFlag = errors.New("unknown feature flag")
	ErrEmptyChange = errors.New("feature flag change has no fields")
)

type layer struct {
	enabled map[Flag]bool
	killed  map[Flag]bool
}

// Configuration is a validated configuration layer. NewConfiguration copies
// all input, so later caller mutation cannot alter a live registry.
type Configuration struct {
	enabled map[Flag]bool
	killed  map[Flag]bool
}

// NewConfiguration creates environment-level overrides and kill switches.
// Omitted flags retain their safe disabled default.
func NewConfiguration(enabled, killed map[Flag]bool) (Configuration, error) {
	if err := validateFlags(enabled); err != nil {
		return Configuration{}, err
	}
	if err := validateFlags(killed); err != nil {
		return Configuration{}, err
	}
	return Configuration{
		enabled: cloneMap(enabled),
		killed:  cloneMap(killed),
	}, nil
}

func validateFlags(values map[Flag]bool) error {
	for flag := range values {
		if !IsCanonical(flag) {
			return fmt.Errorf("%w: %q", ErrUnknownFlag, flag)
		}
	}
	return nil
}

// IsCanonical reports whether flag belongs to the fixed product registry.
func IsCanonical(flag Flag) bool {
	switch flag {
	case FlagSow, FlagFires, FlagAI, FlagPayments, FlagGate:
		return true
	default:
		return false
	}
}

// Decision is the effective state of one flag in a specific snapshot.
type Decision struct {
	Flag    Flag
	Enabled bool
	Killed  bool
	Known   bool
	Source  Source
	Version uint64
}

// Snapshot is an immutable point-in-time configuration view.
type Snapshot struct {
	version     uint64
	environment layer
	runtime     layer
}

// Version returns the monotonically increasing snapshot version.
func (snapshot *Snapshot) Version() uint64 {
	if snapshot == nil {
		return 0
	}
	return snapshot.version
}

// Evaluate resolves a flag with this precedence:
// runtime kill > environment kill > runtime override > environment override >
// safe disabled default. Kill switches therefore always deny the capability.
func (snapshot *Snapshot) Evaluate(flag Flag) Decision {
	if snapshot == nil || !IsCanonical(flag) {
		return Decision{Flag: flag, Source: SourceUnknown}
	}

	decision := Decision{
		Flag:    flag,
		Known:   true,
		Source:  SourceDefault,
		Version: snapshot.version,
	}
	if enabled, ok := snapshot.environment.enabled[flag]; ok {
		decision.Enabled = enabled
		decision.Source = SourceEnvironment
	}
	if enabled, ok := snapshot.runtime.enabled[flag]; ok {
		decision.Enabled = enabled
		decision.Source = SourceRuntime
	}
	if snapshot.environment.killed[flag] {
		decision.Enabled = false
		decision.Killed = true
		decision.Source = SourceEnvironmentKill
	}
	if snapshot.runtime.killed[flag] {
		decision.Enabled = false
		decision.Killed = true
		decision.Source = SourceRuntimeKill
	}
	return decision
}

// Change is an atomic runtime configuration mutation. Pointer fields
// distinguish "set false" from "leave unchanged".
type Change struct {
	Flag    Flag
	Enabled *bool
	Killed  *bool
}

// Registry owns runtime configuration and lock-free immutable snapshots.
type Registry struct {
	changeMu sync.Mutex
	auditMu  sync.Mutex
	current  atomic.Pointer[Snapshot]
	audit    AuditSink
	clock    func() time.Time
}

// New creates a registry at version one. A nil audit sink is safely discarded;
// a nil clock uses time.Now.
func New(configuration Configuration, audit AuditSink, clock func() time.Time) *Registry {
	if audit == nil {
		audit = DiscardAudit{}
	}
	if clock == nil {
		clock = time.Now
	}
	registry := &Registry{audit: audit, clock: clock}
	registry.current.Store(&Snapshot{
		version: 1,
		environment: layer{
			enabled: cloneMap(configuration.enabled),
			killed:  cloneMap(configuration.killed),
		},
		runtime: layer{
			enabled: make(map[Flag]bool),
			killed:  make(map[Flag]bool),
		},
	})
	return registry
}

// Snapshot returns the current immutable snapshot. Callers may retain it to
// keep a consistent view throughout one operation.
func (registry *Registry) Snapshot() *Snapshot {
	return registry.current.Load()
}

// Evaluate resolves against the latest snapshot and emits a metadata-only
// audit record. Unknown flags are denied rather than falling open.
func (registry *Registry) Evaluate(flag Flag) Decision {
	decision := registry.Snapshot().Evaluate(flag)
	registry.record(evaluationRecord(registry.clock().UTC(), decision))
	return decision
}

// Apply validates and atomically publishes a batch of runtime changes. If one
// change is invalid, none are applied.
func (registry *Registry) Apply(changes ...Change) error {
	registry.changeMu.Lock()
	defer registry.changeMu.Unlock()

	if len(changes) == 0 {
		registry.record(rejectedRecord(registry.clock().UTC(), "", RejectEmptyChange))
		return ErrEmptyChange
	}
	for _, change := range changes {
		if !IsCanonical(change.Flag) {
			registry.record(rejectedRecord(registry.clock().UTC(), change.Flag, RejectUnknownFlag))
			return fmt.Errorf("%w: %q", ErrUnknownFlag, change.Flag)
		}
		if change.Enabled == nil && change.Killed == nil {
			registry.record(rejectedRecord(registry.clock().UTC(), change.Flag, RejectEmptyChange))
			return fmt.Errorf("%w: %q", ErrEmptyChange, change.Flag)
		}
	}

	previous := registry.current.Load()
	next := &Snapshot{
		version:     previous.version + 1,
		environment: previous.environment,
		runtime: layer{
			enabled: cloneMap(previous.runtime.enabled),
			killed:  cloneMap(previous.runtime.killed),
		},
	}
	for _, change := range changes {
		if change.Enabled != nil {
			next.runtime.enabled[change.Flag] = *change.Enabled
		}
		if change.Killed != nil {
			next.runtime.killed[change.Flag] = *change.Killed
		}
	}
	registry.current.Store(next)

	for _, change := range changes {
		registry.record(changeRecord(
			registry.clock().UTC(),
			next.Evaluate(change.Flag),
			change,
		))
	}
	return nil
}

func (registry *Registry) record(record AuditRecord) {
	registry.auditMu.Lock()
	defer registry.auditMu.Unlock()
	registry.audit.Record(record)
}

func cloneMap(source map[Flag]bool) map[Flag]bool {
	result := make(map[Flag]bool, len(source))
	for flag, value := range source {
		result[flag] = value
	}
	return result
}
