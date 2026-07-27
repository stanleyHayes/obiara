package drrehearsal

import (
	"context"
	"errors"
	"fmt"
	"time"
)

//go:generate mockgen -source=service.go -destination=mock_service_test.go -package=drrehearsal

type Inspector interface {
	Capture(context.Context, string, time.Time) (Snapshot, error)
}

type Restorer interface {
	Restore(context.Context, Plan) error
	Cleanup(context.Context, string) error
}

type Authority interface {
	Approve(context.Context, string, string) (Approval, error)
}

type EvidenceStore interface {
	Append(context.Context, Evidence, string) error
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	inspector Inspector
	restorer  Restorer
	authority Authority
	store     EvidenceStore
	clock     Clock
}

func NewService(i Inspector, r Restorer, a Authority, s EvidenceStore, c Clock) *Service {
	return &Service{inspector: i, restorer: r, authority: a, store: s, clock: c}
}

func (s *Service) Run(ctx context.Context, plan Plan) (e Evidence, err error) {
	now := s.clock.Now().UTC()
	if err := plan.Validate(now); err != nil {
		return Evidence{}, err
	}
	source, err := s.inspector.Capture(ctx, plan.Source, plan.PointInTime)
	if err != nil {
		return Evidence{}, fmt.Errorf("capture source: %w", err)
	}
	if !source.Watermark.Equal(plan.PointInTime) {
		return Evidence{}, ErrIntegrityMismatch
	}
	started := s.clock.Now().UTC()
	if err := s.restorer.Restore(ctx, plan); err != nil {
		return Evidence{}, fmt.Errorf("restore isolated target: %w", err)
	}
	cleanupNeeded := true
	defer func() {
		if cleanupNeeded {
			err = errors.Join(err, s.restorer.Cleanup(ctx, plan.Target))
		}
	}()
	target, err := s.inspector.Capture(ctx, plan.Target, plan.PointInTime)
	if err != nil {
		return Evidence{}, fmt.Errorf("inspect isolated target: %w", err)
	}
	validationDigest, err := Verify(source, target)
	if err != nil {
		return Evidence{}, err
	}
	completed := s.clock.Now().UTC()
	rpo := started.Sub(plan.PointInTime)
	rto := completed.Sub(started)
	if rpo < 0 || rpo > MaxRPO || rto < 0 || rto > MaxRTO {
		return Evidence{}, ErrObjectiveMissed
	}
	if err := s.restorer.Cleanup(ctx, plan.Target); err != nil {
		return Evidence{}, fmt.Errorf("cleanup isolated target: %w", err)
	}
	cleanupNeeded = false

	dataOwner, err := s.authority.Approve(ctx, "data-owner", validationDigest)
	if err != nil {
		return Evidence{}, fmt.Errorf("data-owner approval: %w", err)
	}
	security, err := s.authority.Approve(ctx, "security", validationDigest)
	if err != nil {
		return Evidence{}, fmt.Errorf("security approval: %w", err)
	}
	if dataOwner.PrincipalRef == "" || security.PrincipalRef == "" ||
		dataOwner.PrincipalRef == security.PrincipalRef ||
		len(dataOwner.EvidenceRef) != 64 || len(security.EvidenceRef) != 64 {
		return Evidence{}, ErrApprovalConflict
	}
	e = Evidence{
		SchemaVersion: 1, Environment: plan.Environment, SyntheticOnly: true,
		RehearsalID: plan.RehearsalID, SourceSnapshotAt: source.Watermark.UTC(),
		RestoreStartedAt: started, RestoreCompletedAt: completed, Target: plan.Target,
		PointInTime: true, ValidationDigest: validationDigest,
		DataOwnerApprovalRef: dataOwner.EvidenceRef, SecurityApprovalRef: security.EvidenceRef,
		Result: "pass", RPOMinutesObserved: int64(rpo / time.Minute),
		RTOMinutesObserved: int64(rto / time.Minute),
	}
	digest, err := e.Digest()
	if err != nil {
		return Evidence{}, err
	}
	if err := s.store.Append(ctx, e, digest); err != nil {
		return Evidence{}, fmt.Errorf("append immutable evidence: %w", err)
	}
	return e, nil
}
