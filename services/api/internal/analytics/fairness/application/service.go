package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/domain"
)

var (
	ErrInvalid     = errors.New("invalid fairness projection request")
	ErrUnavailable = errors.New("fairness projection unavailable")
	ErrApplied     = errors.New("fairness projection already applied")
	ErrConflict    = errors.New("fairness projection conflict")
	ErrNotFound    = errors.New("fairness projection not found")
)

type Service struct {
	definitions DefinitionCatalog
	source      AggregateSource
	repo        Repository
	authority   Authority
	ids         IDSource
	clock       Clock
}

func New(d DefinitionCatalog, s AggregateSource, r Repository, a Authority, ids IDSource, c Clock) Service {
	return Service{d, s, r, a, ids, c}
}
func (s Service) Project(ctx context.Context, actor, quarterKey string, expectedDefinition, expectedSnapshot uint64) (domain.Report, error) {
	if err := s.authority.RequireProjector(ctx, actor); err != nil {
		return domain.Report{}, err
	}
	d, err := s.definitions.Current(ctx)
	if err != nil || d.Spec().Version != expectedDefinition {
		return domain.Report{}, ErrUnavailable
	}
	snapshot, err := s.source.CurrentQuarter(ctx, quarterKey)
	if err != nil || snapshot.QuarterKey != quarterKey || snapshot.Version != expectedSnapshot {
		return domain.Report{}, ErrUnavailable
	}
	report, err := domain.Evaluate(s.ids.NewID(), d, snapshot, s.clock.Now())
	if err != nil {
		return domain.Report{}, ErrInvalid
	}
	current, err := s.definitions.Current(ctx)
	if err != nil || current.Spec().ID != d.Spec().ID || current.Spec().Version != d.Spec().Version {
		return domain.Report{}, ErrUnavailable
	}
	if err = s.authority.RequireProjector(ctx, actor); err != nil {
		return domain.Report{}, err
	}
	if err = s.repo.Insert(ctx, report); err != nil {
		if !errors.Is(err, ErrApplied) {
			return domain.Report{}, err
		}
		existing, findErr := s.repo.Find(ctx, quarterKey, d.Spec().Version)
		if findErr != nil {
			return domain.Report{}, findErr
		}
		if existing.Fingerprint != report.Fingerprint {
			return domain.Report{}, ErrConflict
		}
		return existing, nil
	}
	return report, nil
}
