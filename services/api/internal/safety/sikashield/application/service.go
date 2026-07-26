package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/sikashield/domain"
	"time"
)

var ErrInvalid = errors.New("invalid sika shield request")

type Service struct {
	catalog   Catalog
	metrics   MetricsGate
	evidence  EvidenceVerifier
	cases     CaseRouter
	authority Authority
	now       func() time.Time
}

func NewService(c Catalog, m MetricsGate, e EvidenceVerifier, r CaseRouter, a Authority, n func() time.Time) Service {
	return Service{c, m, e, r, a, n}
}

type Request struct {
	Actor  string
	Signal domain.Signal
}

func (s Service) Evaluate(ctx context.Context, q Request) (domain.Decision, error) {
	if s.catalog == nil || s.metrics == nil || s.evidence == nil || s.cases == nil || s.authority == nil || s.now == nil {
		return domain.Decision{}, ErrInvalid
	}
	if err := s.authority.RequireOfflineEvaluator(ctx, q.Actor); err != nil {
		return domain.Decision{}, err
	}
	p, err := s.catalog.Current(ctx, q.Signal.PatternKey)
	if err != nil {
		return domain.Decision{}, err
	}
	m, err := s.metrics.Current(ctx, p.Key, p.Version)
	if err != nil || !m.Pass() {
		return domain.Decision{Outcome: domain.OutcomeNoAction}, err
	}
	if err = s.evidence.Revalidate(ctx, q.Signal.EvidenceRef, q.Signal.Source); err != nil {
		return domain.Decision{Outcome: domain.OutcomeNoAction}, nil
	}
	d, err := domain.Evaluate([]domain.Pattern{p}, q.Signal, s.now().UTC())
	if err != nil {
		return domain.Decision{}, err
	}
	if d.Outcome == domain.OutcomeHumanReview {
		if err = s.cases.OpenHumanCase(ctx, d); err != nil {
			return domain.Decision{Outcome: domain.OutcomeNoAction}, nil
		}
	}
	return d, nil
}
