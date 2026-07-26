package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/domain"
)

var (
	ErrUnder18            = errors.New("member is below the minimum age")
	ErrPurgePending       = errors.New("personal-data purge remains pending")
	ErrServiceUnavailable = errors.New("safeguarding service unavailable")
)

type Service struct {
	store  RestrictionStore
	purger ArtifactPurger
	keyer  Keyer
	ids    IDSource
	now    func() time.Time
}

func NewService(store RestrictionStore, purger ArtifactPurger, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{store: store, purger: purger, keyer: keyer, ids: ids, now: now}
}

type Assessment struct {
	CommandID   string
	SubjectID   string
	SourceRef   string
	DateOfBirth time.Time
}

type Decision struct {
	Allowed     bool
	Restriction domain.Restriction
	Replayed    bool
}

// Assess uses server time exclusively. Under-18 subjects are persisted as
// blocked before any purge attempt, so provider/purge failure can never turn
// into access.
func (service Service) Assess(ctx context.Context, assessment Assessment) (Decision, error) {
	if service.store == nil || service.purger == nil || service.keyer == nil ||
		service.ids == nil || service.now == nil {
		return Decision{}, ErrServiceUnavailable
	}
	now := service.now().UTC()
	subjectID := strings.TrimSpace(assessment.SubjectID)
	sourceRef := strings.TrimSpace(assessment.SourceRef)
	commandID := strings.TrimSpace(assessment.CommandID)
	subjectKey, err := service.keyer.Key(subjectID)
	if err != nil {
		return Decision{}, ErrServiceUnavailable
	}
	sourceKey, err := service.keyer.Key(sourceRef)
	if err != nil {
		return Decision{}, ErrServiceUnavailable
	}
	existing, findErr := service.store.FindBySubjectKey(ctx, subjectKey)
	if findErr == nil {
		return Decision{Allowed: false, Restriction: existing, Replayed: true}, ErrUnder18
	}
	if !errors.Is(findErr, ErrRestrictionNotFound) {
		return Decision{}, ErrServiceUnavailable
	}
	eligible, err := domain.Eligible(assessment.DateOfBirth, now)
	if err != nil {
		return Decision{}, err
	}
	if eligible {
		return Decision{Allowed: true}, nil
	}
	restriction, err := domain.NewRestriction(service.ids.NewID(), commandID, subjectKey, sourceKey, now)
	if err != nil {
		return Decision{}, err
	}
	job := PurgeJob{
		RestrictionID: restriction.ID(), SubjectID: subjectID,
		SourceRef: sourceRef, PurgeDueAt: restriction.PurgeDueAt(),
	}
	restriction, replayed, err := service.store.CreateBlocked(ctx, restriction, job)
	if err != nil {
		if errors.Is(err, ErrCommandConflict) {
			return Decision{Allowed: false}, ErrCommandConflict
		}
		return Decision{Allowed: false}, ErrServiceUnavailable
	}
	decision := Decision{Allowed: false, Restriction: restriction, Replayed: replayed}
	if restriction.PurgeStatus() == domain.PurgeCompleted {
		return decision, ErrUnder18
	}
	if err := service.purge(ctx, job, restriction); err != nil {
		return decision, errors.Join(ErrUnder18, ErrPurgePending)
	}
	updated, findErr := service.store.FindByID(ctx, restriction.ID())
	if findErr == nil {
		decision.Restriction = updated
	}
	return decision, ErrUnder18
}

// PurgePending is the worker entrypoint. dueBefore should include the next
// worker interval, ensuring a retry is attempted before the 24-hour deadline.
func (service Service) PurgePending(ctx context.Context, dueBefore time.Time, limit int) (int, error) {
	if service.store == nil || service.purger == nil || service.now == nil || limit < 1 {
		return 0, ErrServiceUnavailable
	}
	jobs, err := service.store.FindPending(ctx, dueBefore.UTC(), limit)
	if err != nil {
		return 0, ErrServiceUnavailable
	}
	completed := 0
	for _, job := range jobs {
		restriction, findErr := service.store.FindByID(ctx, job.RestrictionID)
		if findErr != nil {
			return completed, ErrServiceUnavailable
		}
		if purgeErr := service.purge(ctx, job, restriction); purgeErr != nil {
			return completed, errors.Join(ErrPurgePending, purgeErr)
		}
		completed++
	}
	return completed, nil
}

func (service Service) purge(ctx context.Context, job PurgeJob, restriction domain.Restriction) error {
	if err := service.purger.Purge(ctx, job.SubjectID, job.SourceRef); err != nil {
		return err
	}
	purged, err := restriction.MarkPurged(service.now().UTC(), restriction.Version())
	if err != nil {
		return err
	}
	if err := service.store.CompletePurge(ctx, purged, restriction.Version()); err != nil {
		if errors.Is(err, ErrOptimisticConflict) {
			current, findErr := service.store.FindByID(ctx, restriction.ID())
			if findErr == nil && current.PurgeStatus() == domain.PurgeCompleted {
				return nil
			}
		}
		return err
	}
	return nil
}
