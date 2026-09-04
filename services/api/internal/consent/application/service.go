package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

type Service struct {
	records  Repository
	purposes PurposeCatalog
	now      func() time.Time
}

func NewService(records Repository, purposes PurposeCatalog, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{records: records, purposes: purposes, now: now}
}

type Command struct {
	CommandID        string
	SubjectID        string
	PurposeID        string
	PurposeVersion   uint64
	ExpectedRevision uint64
	ActorID          string
	ActorKind        domain.ActorKind
	Source           domain.Source
	Evidence         domain.Evidence
}

type Result struct {
	Record   domain.Record
	Replayed bool
}

func (service Service) Grant(ctx context.Context, command Command) (Result, error) {
	return service.execute(ctx, command, domain.ActionGranted)
}

func (service Service) Withdraw(ctx context.Context, command Command) (Result, error) {
	return service.execute(ctx, command, domain.ActionWithdrawn)
}

func (service Service) execute(ctx context.Context, command Command, action domain.Action) (Result, error) {
	if service.records == nil || service.purposes == nil || service.now == nil {
		return Result{}, ErrRepositoryUnavailable
	}
	command.CommandID = strings.TrimSpace(command.CommandID)
	command.SubjectID = strings.TrimSpace(command.SubjectID)
	command.PurposeID = strings.TrimSpace(command.PurposeID)

	purpose, err := service.purposes.FindVersion(ctx, command.PurposeID, command.PurposeVersion)
	if err != nil {
		return Result{}, domain.ErrInvalidPurpose
	}
	record, err := service.records.Find(ctx, Key{SubjectID: command.SubjectID, PurposeID: command.PurposeID})
	if errors.Is(err, ErrNotFound) {
		record, err = domain.NewRecord(command.SubjectID, command.PurposeID)
	}
	if err != nil {
		return Result{}, ErrRepositoryUnavailable
	}
	change := domain.Change{
		CommandID:        command.CommandID,
		ExpectedRevision: command.ExpectedRevision,
		Purpose:          purpose,
		ActorID:          command.ActorID,
		ActorKind:        command.ActorKind,
		Source:           command.Source,
		Evidence:         command.Evidence,
		RecordedAt:       service.now().UTC(),
	}
	if record.HasCommand(command.CommandID) {
		if !record.ReplayMatches(change, action) {
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{Record: record, Replayed: true}, nil
	}
	var next domain.Record
	if action == domain.ActionGranted {
		next, err = record.Grant(change)
	} else {
		next, err = record.Withdraw(change)
	}
	if err != nil {
		return Result{}, err
	}
	if err = service.records.Save(ctx, next, record.Revision(), command.CommandID); err == nil {
		return Result{Record: next}, nil
	}
	if errors.Is(err, ErrCommandAlreadyApplied) {
		reloaded, findErr := service.records.Find(ctx, Key{SubjectID: command.SubjectID, PurposeID: command.PurposeID})
		if findErr == nil && reloaded.ReplayMatches(change, action) {
			return Result{Record: reloaded, Replayed: true}, nil
		}
		if findErr == nil && reloaded.HasCommand(command.CommandID) {
			return Result{}, domain.ErrCommandMismatch
		}
	}
	if errors.Is(err, ErrOptimisticConflict) {
		return Result{}, domain.ErrStaleRevision
	}
	return Result{}, ErrRepositoryUnavailable
}

// Revision reports the subject's current revision for a purpose, and zero
// when they have no record of it yet.
//
// A writer that assumes zero can only ever succeed on a subject's first
// attempt: the domain refuses a change whose ExpectedRevision does not match
// what is stored, so a second pass over an existing receipt is rejected as
// stale rather than recognised as a repeat.
func (service Service) Revision(ctx context.Context, subjectID, purposeID string) (uint64, error) {
	if service.records == nil {
		return 0, ErrRepositoryUnavailable
	}
	record, err := service.records.Find(ctx, Key{
		SubjectID: strings.TrimSpace(subjectID),
		PurposeID: strings.TrimSpace(purposeID),
	})
	if errors.Is(err, ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, ErrRepositoryUnavailable
	}
	return record.Revision(), nil
}

func (service Service) Effective(ctx context.Context, subjectID, purposeID string, version uint64) (bool, error) {
	if service.records == nil || service.purposes == nil || service.now == nil {
		return false, ErrRepositoryUnavailable
	}
	purpose, err := service.purposes.FindVersion(ctx, strings.TrimSpace(purposeID), version)
	if err != nil {
		return false, domain.ErrInvalidPurpose
	}
	record, err := service.records.Find(ctx, Key{
		SubjectID: strings.TrimSpace(subjectID),
		PurposeID: strings.TrimSpace(purposeID),
	})
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, ErrRepositoryUnavailable
	}
	return record.Effective(purpose, service.now().UTC()), nil
}
