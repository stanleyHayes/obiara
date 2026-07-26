package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/anomaly/domain"
	"time"
)

var ErrInvalid = errors.New("invalid anomaly request")

type Service struct {
	rules     RuleCatalog
	consent   ConsentVerifier
	authority Authority
	cases     CaseRouter
	now       func() time.Time
}

func NewService(r RuleCatalog, c ConsentVerifier, a Authority, h CaseRouter, n func() time.Time) Service {
	return Service{r, c, a, h, n}
}

type Request struct {
	Actor     string
	Aggregate domain.Aggregate
}

func (s Service) Evaluate(ctx context.Context, q Request) (domain.Decision, error) {
	if s.rules == nil || s.consent == nil || s.authority == nil || s.cases == nil || s.now == nil {
		return domain.Decision{}, ErrInvalid
	}
	if err := s.authority.RequireOfflineEvaluator(ctx, q.Actor); err != nil {
		return domain.Decision{}, err
	}
	r, err := s.rules.Current(ctx, q.Aggregate.Shape)
	if err != nil {
		return domain.Decision{Outcome: domain.OutcomeNoAction}, nil
	}
	if err = s.consent.Revalidate(ctx, q.Aggregate.EvidenceRef); err != nil {
		return domain.Decision{Outcome: domain.OutcomeNoAction}, nil
	}
	d, err := domain.Evaluate(r, q.Aggregate, s.now().UTC())
	if err != nil {
		return domain.Decision{}, err
	}
	if d.Outcome != domain.OutcomeHumanReview {
		return d, nil
	}
	if err = s.authority.RequireHumanRoute(ctx, q.Actor, q.Aggregate.EvidenceRef); err != nil {
		return domain.Decision{Outcome: domain.OutcomeNoAction}, nil
	}
	if err = s.cases.OpenHumanCase(ctx, d); err != nil {
		return domain.Decision{Outcome: domain.OutcomeNoAction}, nil
	}
	return d, nil
}
