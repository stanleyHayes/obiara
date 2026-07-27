package drrehearsal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"go.uber.org/mock/gomock"
)

type sequenceClock struct {
	values []time.Time
	index  int
}

func (c *sequenceClock) Now() time.Time {
	v := c.values[c.index]
	c.index++
	return v
}

func validSnapshot(at time.Time) Snapshot {
	return Snapshot{Watermark: at, Collections: []CollectionFact{
		{Name: "audit", Count: 2, Digest: strings.Repeat("a", 64), Indexes: []string{"_id_", "event_1"}, Transaction: true, Audit: true},
		{Name: "ledger", Count: 3, Digest: strings.Repeat("b", 64), Indexes: []string{"_id_", "entry_1"}, Transaction: true, Audit: true},
	}}
}

func TestServiceRunProducesApprovedEvidenceAndCleansTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	inspector, restorer := NewMockInspector(ctrl), NewMockRestorer(ctrl)
	authority, store := NewMockAuthority(ctrl), NewMockEvidenceStore(ctrl)
	point := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	start, done := point.Add(2*time.Minute), point.Add(7*time.Minute)
	plan := Plan{RehearsalID: "dr-20260727", Environment: "staging", Source: "staging-source", Target: "isolated-dr-20260727", PointInTime: point}
	snapshot := validSnapshot(point)

	gomock.InOrder(
		inspector.EXPECT().Capture(gomock.Any(), plan.Source, point).Return(snapshot, nil),
		restorer.EXPECT().Restore(gomock.Any(), plan).Return(nil),
		inspector.EXPECT().Capture(gomock.Any(), plan.Target, point).Return(snapshot, nil),
		restorer.EXPECT().Cleanup(gomock.Any(), plan.Target).Return(nil),
		authority.EXPECT().Approve(gomock.Any(), "data-owner", gomock.Any()).Return(Approval{PrincipalRef: "owner-opaque", EvidenceRef: strings.Repeat("c", 64)}, nil),
		authority.EXPECT().Approve(gomock.Any(), "security", gomock.Any()).Return(Approval{PrincipalRef: "security-opaque", EvidenceRef: strings.Repeat("d", 64)}, nil),
		store.EXPECT().Append(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, e Evidence, digest string) error {
			if len(digest) != 64 || !e.PointInTime || e.Result != "pass" {
				t.Fatalf("invalid immutable evidence: %+v %q", e, digest)
			}
			return nil
		}),
	)
	evidence, err := NewService(inspector, restorer, authority, store, &sequenceClock{values: []time.Time{start, start, done}}).Run(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.RPOMinutesObserved != 2 || evidence.RTOMinutesObserved != 5 {
		t.Fatalf("unexpected objectives: %+v", evidence)
	}
}

func TestServiceFailsClosedAndCleansCorruptRestore(t *testing.T) {
	ctrl := gomock.NewController(t)
	inspector, restorer := NewMockInspector(ctrl), NewMockRestorer(ctrl)
	point := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	plan := Plan{RehearsalID: "dr-corrupt", Environment: "staging", Source: "staging-source", Target: "isolated-corrupt", PointInTime: point}
	source, corrupt := validSnapshot(point), validSnapshot(point)
	corrupt.Collections[0].Count++
	gomock.InOrder(
		inspector.EXPECT().Capture(gomock.Any(), plan.Source, point).Return(source, nil),
		restorer.EXPECT().Restore(gomock.Any(), plan).Return(nil),
		inspector.EXPECT().Capture(gomock.Any(), plan.Target, point).Return(corrupt, nil),
		restorer.EXPECT().Cleanup(gomock.Any(), plan.Target).Return(nil),
	)
	neverAuthority, neverStore := NewMockAuthority(ctrl), NewMockEvidenceStore(ctrl)
	_, err := NewService(inspector, restorer, neverAuthority, neverStore, &sequenceClock{values: []time.Time{point, point}}).Run(context.Background(), plan)
	if !errors.Is(err, ErrIntegrityMismatch) {
		t.Fatalf("expected integrity failure, got %v", err)
	}
}

func TestServiceRejectsProductionBeforeAnyPortCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	now := time.Now().UTC()
	_, err := NewService(NewMockInspector(ctrl), NewMockRestorer(ctrl), NewMockAuthority(ctrl), NewMockEvidenceStore(ctrl),
		&sequenceClock{values: []time.Time{now}}).Run(context.Background(), Plan{
		RehearsalID: "production-dr", Environment: "production", Source: "prod", Target: "isolated-prod", PointInTime: now,
	})
	if !errors.Is(err, ErrUnsafePlan) {
		t.Fatalf("expected unsafe plan, got %v", err)
	}
}

func TestPlanNeverAcceptsNonIsolatedTarget(t *testing.T) {
	now := time.Now().UTC()
	property := func(target string) bool {
		if strings.HasPrefix(target, "isolated-") {
			return true
		}
		p := Plan{RehearsalID: "property-run", Environment: "staging", Source: "source", Target: target, PointInTime: now}
		return errors.Is(p.Validate(now), ErrUnsafePlan)
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

func FuzzPlanTargetIsolation(f *testing.F) {
	f.Add("production")
	f.Add("isolated-rehearsal")
	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	f.Fuzz(func(t *testing.T, target string) {
		p := Plan{RehearsalID: "fuzz-run-001", Environment: "staging", Source: "source", Target: target, PointInTime: now}
		err := p.Validate(now)
		if err == nil && !strings.HasPrefix(target, "isolated-") {
			t.Fatalf("accepted unsafe target %q", target)
		}
	})
}
