package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/domain"
)

var (
	ErrInvalid     = errors.New("invalid P0 gate request")
	ErrUnavailable = errors.New("P0 gate unavailable")
	ErrApplied     = errors.New("P0 report already applied")
	ErrConflict    = errors.New("P0 report conflict")
)

type Service struct {
	definitions DefinitionCatalog
	source      AggregateSource
	repo        Repository
	ids         IDSource
	clock       Clock
}

func New(d DefinitionCatalog, s AggregateSource, r Repository, ids IDSource, c Clock) Service {
	return Service{d, s, r, ids, c}
}
func (s Service) Project(ctx context.Context, windowKey string, expectedDefinition, expectedSnapshot uint64) (domain.Report, error) {
	d, e := s.definitions.Current(ctx)
	if e != nil || d.Spec().Version != expectedDefinition {
		return domain.Report{}, ErrUnavailable
	}
	snapshot, e := s.source.Current(ctx, windowKey)
	if e != nil || snapshot.WindowKey != windowKey || snapshot.Version != expectedSnapshot {
		return domain.Report{}, ErrUnavailable
	}
	report, e := domain.Evaluate(s.ids.NewID(), d, snapshot, s.clock.Now())
	if e != nil {
		return domain.Report{}, ErrInvalid
	}
	current, e := s.definitions.Current(ctx)
	if e != nil || current.Spec().ID != d.Spec().ID || current.Spec().Version != d.Spec().Version {
		return domain.Report{}, ErrUnavailable
	}
	if e = s.repo.Insert(ctx, report.Canonical()); e != nil {
		return domain.Report{}, e
	}
	return report.Canonical(), nil
}
