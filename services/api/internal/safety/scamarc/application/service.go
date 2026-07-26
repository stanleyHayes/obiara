// Package application coordinates current consent, reviewed rules, and the
// mandatory human route for E11-S11. It exposes no enforcement port.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/safety/scamarc/domain"
)

var (
	ErrInvalidRequest = errors.New("invalid scam-arc evaluation request")
	ErrUnavailable    = errors.New("scam-arc evaluation unavailable")
)

type Request struct {
	PairKey        string
	ConsentVersion uint64
}

type Result struct {
	Signal *domain.Signal
}

type Service struct {
	consent   Consent
	authority Authority
	rules     RuleCatalog
	events    EventSource
	human     HumanRoute
	ids       IDSource
	clock     Clock
}

func New(consent Consent, authority Authority, rules RuleCatalog, events EventSource, human HumanRoute, ids IDSource, clock Clock) Service {
	return Service{consent: consent, authority: authority, rules: rules, events: events, human: human, ids: ids, clock: clock}
}

func (service Service) Evaluate(ctx context.Context, request Request) (Result, error) {
	request.PairKey = strings.TrimSpace(request.PairKey)
	if service.consent == nil || service.authority == nil || service.rules == nil ||
		service.events == nil || service.human == nil || service.ids == nil || service.clock == nil ||
		!opaque(request.PairKey) || request.ConsentVersion == 0 {
		return Result{}, ErrInvalidRequest
	}
	if err := service.authority.AuthorizeEvaluation(ctx, request.PairKey); err != nil {
		return Result{}, ErrUnavailable
	}
	allowed, err := service.consent.CurrentAllows(ctx, request.PairKey, MonitoringPurpose, request.ConsentVersion)
	if err != nil || !allowed {
		return Result{}, ErrUnavailable
	}
	rules, err := service.rules.Current(ctx)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	sequenceKey, events, err := service.events.Current(ctx, request.PairKey, domain.MaxEvents)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	sequence, err := domain.NewSequence(sequenceKey, request.PairKey, events)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	signal, err := domain.Evaluate(service.ids.NewID(), sequence, rules, service.clock.Now())
	if errors.Is(err, domain.ErrNoPattern) {
		return Result{}, nil
	}
	if err != nil {
		return Result{}, ErrUnavailable
	}

	// Consent and authority are revalidated immediately before the only
	// outbound operation. The route is human review, never enforcement.
	allowed, err = service.consent.CurrentAllows(ctx, request.PairKey, MonitoringPurpose, request.ConsentVersion)
	if err != nil || !allowed {
		return Result{}, ErrUnavailable
	}
	if err := service.authority.AuthorizeEvaluation(ctx, request.PairKey); err != nil {
		return Result{}, ErrUnavailable
	}
	if err := service.human.Route(ctx, signal); err != nil {
		return Result{}, ErrUnavailable
	}
	return Result{Signal: &signal}, nil
}

func opaque(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}
	return true
}
