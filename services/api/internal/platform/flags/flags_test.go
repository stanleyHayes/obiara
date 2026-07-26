package flags

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCanonicalFlagsDefaultDisabled(t *testing.T) {
	t.Parallel()

	registry := New(Configuration{}, nil, nil)
	for _, flag := range canonicalFlags {
		decision := registry.Evaluate(flag)
		if !decision.Known || decision.Enabled || decision.Killed {
			t.Fatalf("%s: expected known, disabled, not killed; got %+v", flag, decision)
		}
		if decision.Source != SourceDefault || decision.Version != 1 {
			t.Fatalf("%s: unexpected metadata: %+v", flag, decision)
		}
	}
}

func TestPrecedenceAndKillSwitchFailSafe(t *testing.T) {
	t.Parallel()

	configuration, err := NewConfiguration(
		map[Flag]bool{FlagSow: true, FlagAI: true},
		map[Flag]bool{FlagAI: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	registry := New(configuration, nil, nil)

	if got := registry.Evaluate(FlagSow); !got.Enabled || got.Source != SourceEnvironment {
		t.Fatalf("environment enablement not applied: %+v", got)
	}
	if got := registry.Evaluate(FlagAI); got.Enabled || !got.Killed || got.Source != SourceEnvironmentKill {
		t.Fatalf("environment kill switch did not win: %+v", got)
	}

	disabled := false
	enabled := true
	if err := registry.Apply(
		Change{Flag: FlagSow, Enabled: &disabled},
		Change{Flag: FlagAI, Enabled: &enabled, Killed: &disabled},
	); err != nil {
		t.Fatal(err)
	}
	if got := registry.Evaluate(FlagSow); got.Enabled || got.Source != SourceRuntime {
		t.Fatalf("runtime override did not win: %+v", got)
	}
	if got := registry.Evaluate(FlagAI); got.Enabled || !got.Killed || got.Source != SourceEnvironmentKill {
		t.Fatalf("runtime change cleared stronger environment kill: %+v", got)
	}

	if err := registry.Apply(Change{Flag: FlagSow, Killed: &enabled}); err != nil {
		t.Fatal(err)
	}
	if got := registry.Evaluate(FlagSow); got.Enabled || !got.Killed || got.Source != SourceRuntimeKill {
		t.Fatalf("runtime kill did not win: %+v", got)
	}
}

func TestSnapshotIsImmutableAcrossChanges(t *testing.T) {
	t.Parallel()

	registry := New(Configuration{}, nil, nil)
	before := registry.Snapshot()
	enabled := true
	if err := registry.Apply(Change{Flag: FlagGate, Enabled: &enabled}); err != nil {
		t.Fatal(err)
	}
	after := registry.Snapshot()

	if before == after {
		t.Fatal("expected a newly published snapshot")
	}
	if before.Version() != 1 || after.Version() != 2 {
		t.Fatalf("unexpected versions: before=%d after=%d", before.Version(), after.Version())
	}
	if before.Evaluate(FlagGate).Enabled {
		t.Fatal("retained snapshot was mutated")
	}
	if !after.Evaluate(FlagGate).Enabled {
		t.Fatal("new snapshot did not contain change")
	}
}

func TestConfigurationCopiesCallerMaps(t *testing.T) {
	t.Parallel()

	enabled := map[Flag]bool{FlagFires: true}
	configuration, err := NewConfiguration(enabled, nil)
	if err != nil {
		t.Fatal(err)
	}
	registry := New(configuration, nil, nil)
	enabled[FlagFires] = false
	configuration.enabled[FlagFires] = false

	if got := registry.Evaluate(FlagFires); !got.Enabled {
		t.Fatalf("caller mutation leaked into registry: %+v", got)
	}
}

func TestUnknownFlagDeniedAndAudited(t *testing.T) {
	t.Parallel()

	audit := &memoryAudit{}
	registry := New(Configuration{}, audit, fixedClock)
	decision := registry.Evaluate(Flag("not-a-real-flag"))
	if decision.Known || decision.Enabled || decision.Source != SourceUnknown {
		t.Fatalf("unknown flag was not denied: %+v", decision)
	}

	records := audit.Records()
	if len(records) != 1 || records[0].Known || records[0].Flag != "" {
		t.Fatalf("unexpected audit records: %+v", records)
	}
}

func TestApplyRejectsInvalidBatchAtomically(t *testing.T) {
	t.Parallel()

	audit := &memoryAudit{}
	registry := New(Configuration{}, audit, fixedClock)
	enabled := true
	err := registry.Apply(
		Change{Flag: FlagSow, Enabled: &enabled},
		Change{Flag: Flag("secret@example.com"), Enabled: &enabled},
	)
	if !errors.Is(err, ErrUnknownFlag) {
		t.Fatalf("expected unknown flag error, got %v", err)
	}
	if registry.Snapshot().Version() != 1 || registry.Snapshot().Evaluate(FlagSow).Enabled {
		t.Fatal("part of an invalid batch was applied")
	}

	records := audit.Records()
	if len(records) != 1 || records[0].Rejection != RejectUnknownFlag {
		t.Fatalf("unexpected rejected-change audit: %+v", records)
	}
	if records[0].Flag != "" || strings.Contains(fmt.Sprintf("%+v", records[0]), "secret@example.com") {
		t.Fatalf("untrusted flag name leaked into audit: %+v", records[0])
	}
}

func TestApplyRejectsEmptyChange(t *testing.T) {
	t.Parallel()

	audit := &memoryAudit{}
	registry := New(Configuration{}, audit, fixedClock)
	err := registry.Apply(Change{Flag: FlagPayments})
	if !errors.Is(err, ErrEmptyChange) {
		t.Fatalf("expected empty change error, got %v", err)
	}
	records := audit.Records()
	if len(records) != 1 || records[0].Rejection != RejectEmptyChange {
		t.Fatalf("unexpected audit: %+v", records)
	}
}

func TestApplyRejectsEmptyBatch(t *testing.T) {
	t.Parallel()

	audit := &memoryAudit{}
	registry := New(Configuration{}, audit, fixedClock)
	err := registry.Apply()
	if !errors.Is(err, ErrEmptyChange) {
		t.Fatalf("expected empty change error, got %v", err)
	}
	if registry.Snapshot().Version() != 1 {
		t.Fatal("empty batch published a snapshot")
	}
	records := audit.Records()
	if len(records) != 1 || records[0].Rejection != RejectEmptyChange || records[0].Flag != "" {
		t.Fatalf("unexpected audit: %+v", records)
	}
}

func TestConfigurationAuditContainsOnlySafeMetadata(t *testing.T) {
	t.Parallel()

	audit := &memoryAudit{}
	registry := New(Configuration{}, audit, fixedClock)
	enabled := true
	killed := false
	if err := registry.Apply(Change{Flag: FlagPayments, Enabled: &enabled, Killed: &killed}); err != nil {
		t.Fatal(err)
	}

	records := audit.Records()
	if len(records) != 1 {
		t.Fatalf("expected one record, got %d", len(records))
	}
	want := AuditRecord{
		At:            fixedClock(),
		Kind:          AuditConfiguration,
		Flag:          FlagPayments,
		Known:         true,
		Enabled:       true,
		Source:        SourceRuntime,
		Version:       2,
		ChangedFields: []ChangedField{ChangedEnabled, ChangedKillSwitch},
	}
	if !reflect.DeepEqual(records[0], want) {
		t.Fatalf("audit mismatch:\n got: %+v\nwant: %+v", records[0], want)
	}

	recordType := reflect.TypeOf(AuditRecord{})
	for _, forbidden := range []string{"Actor", "Subject", "Reason", "Request", "Secret", "Value", "Payload"} {
		if _, exists := recordType.FieldByName(forbidden); exists {
			t.Fatalf("audit record must not expose %s", forbidden)
		}
	}
}

func TestConcurrentEvaluateAndApply(t *testing.T) {
	t.Parallel()

	audit := &memoryAudit{}
	registry := New(Configuration{}, audit, fixedClock)
	var wait sync.WaitGroup

	for writer := 0; writer < 8; writer++ {
		wait.Add(1)
		go func(writer int) {
			defer wait.Done()
			for iteration := 0; iteration < 250; iteration++ {
				enabled := (writer+iteration)%2 == 0
				killed := (writer+iteration)%7 == 0
				if err := registry.Apply(Change{
					Flag:    canonicalFlags[(writer+iteration)%len(canonicalFlags)],
					Enabled: &enabled,
					Killed:  &killed,
				}); err != nil {
					t.Errorf("apply: %v", err)
					return
				}
			}
		}(writer)
	}
	for reader := 0; reader < 24; reader++ {
		wait.Add(1)
		go func(reader int) {
			defer wait.Done()
			for iteration := 0; iteration < 500; iteration++ {
				decision := registry.Evaluate(canonicalFlags[(reader+iteration)%len(canonicalFlags)])
				if !decision.Known {
					t.Errorf("canonical flag became unknown: %+v", decision)
					return
				}
				if decision.Killed && decision.Enabled {
					t.Errorf("killed feature was enabled: %+v", decision)
					return
				}
			}
		}(reader)
	}
	wait.Wait()

	if got, want := registry.Snapshot().Version(), uint64(1+(8*250)); got != want {
		t.Fatalf("lost writes: version=%d want=%d", got, want)
	}
}

func TestNewConfigurationRejectsUnknownFlag(t *testing.T) {
	t.Parallel()

	_, err := NewConfiguration(map[Flag]bool{Flag("unknown"): true}, nil)
	if !errors.Is(err, ErrUnknownFlag) {
		t.Fatalf("expected unknown flag error, got %v", err)
	}
}

func TestAuditShapeCannotCarryFreeFormData(t *testing.T) {
	t.Parallel()

	record := fmt.Sprintf("%+v", AuditRecord{
		Kind: AuditConfiguration,
		Flag: FlagSow,
	})
	for _, sensitive := range []string{"email", "token", "reason", "payload"} {
		if strings.Contains(strings.ToLower(record), sensitive) {
			t.Fatalf("audit representation unexpectedly contains %q: %s", sensitive, record)
		}
	}
}

type memoryAudit struct {
	mu      sync.Mutex
	records []AuditRecord
}

func (audit *memoryAudit) Record(record AuditRecord) {
	audit.mu.Lock()
	defer audit.mu.Unlock()
	record.ChangedFields = append([]ChangedField(nil), record.ChangedFields...)
	audit.records = append(audit.records, record)
}

func (audit *memoryAudit) Records() []AuditRecord {
	audit.mu.Lock()
	defer audit.mu.Unlock()
	return append([]AuditRecord(nil), audit.records...)
}

func fixedClock() time.Time {
	return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
}
