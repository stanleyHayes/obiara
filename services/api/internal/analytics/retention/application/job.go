package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/domain"
	"time"
)

var (
	ErrUnavailable = errors.New("analytics retention unavailable")
	ErrApplied     = errors.New("analytics retention action applied")
	ErrConflict    = errors.New("analytics retention conflict")
)

type Job struct {
	policies   PolicyCatalog
	store      Store
	pseudonyms Pseudonymizer
	ids        IDSource
	clock      Clock
}

func New(p PolicyCatalog, s Store, k Pseudonymizer, ids IDSource, c Clock) Job {
	return Job{p, s, k, ids, c}
}

type Result struct{ Claimed, Pseudonymized, Aggregated int }

func (j Job) Run(ctx context.Context) (Result, error) {
	policy, e := j.policies.Current(ctx)
	if e != nil {
		return Result{}, ErrUnavailable
	}
	now := j.clock.Now().UTC()
	lease := j.ids.NewID()
	candidates, e := j.store.ClaimDue(ctx, now, policy.Spec().BatchSize, lease, now.Add(5*time.Minute))
	if e != nil {
		return Result{}, ErrUnavailable
	}
	result := Result{Claimed: len(candidates)}
	for _, candidate := range candidates {
		decision, e := domain.Decide(candidate, policy, now)
		if e != nil {
			return result, ErrUnavailable
		}
		current, e := j.policies.Current(ctx)
		if e != nil || current.Spec().ID != policy.Spec().ID || current.Spec().Version != policy.Spec().Version {
			return result, ErrUnavailable
		}
		receipt := j.ids.NewID()
		switch decision.Action {
		case domain.ActionPseudonymize:
			newRef, e := j.pseudonyms.Derive(candidate.SubjectRef, decision.PseudonymKeyVersion)
			if e != nil {
				return result, ErrUnavailable
			}
			if e = j.store.Pseudonymize(ctx, candidate, decision, newRef, receipt); e != nil && !errors.Is(e, ErrApplied) {
				return result, e
			}
			result.Pseudonymized++
		case domain.ActionAggregateErase:
			if e = j.store.AggregateErase(ctx, candidate, decision, receipt); e != nil && !errors.Is(e, ErrApplied) {
				return result, e
			}
			result.Aggregated++
		}
	}
	return result, nil
}
