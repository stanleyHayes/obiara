// Package application is the sole E11-S07 counsel egress boundary.
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/counsel/isolation/domain"
)

var (
	ErrInvalidRequest = errors.New("invalid counsel safety escalation")
	ErrUnavailable    = errors.New("counsel safety escalation unavailable")
)

type Request struct {
	SessionKey           string
	ActorKey             string
	SubjectKey           string
	ConsentVersion       uint64
	ExplicitConfirmation bool
}

type Service struct {
	scope     Scope
	consent   Consent
	authority Authority
	safety    SafetySink
	ids       IDSource
	now       func() time.Time
}

func New(scope Scope, consent Consent, authority Authority, safety SafetySink, ids IDSource, now func() time.Time) Service {
	return Service{scope: scope, consent: consent, authority: authority, safety: safety, ids: ids, now: now}
}

// Escalate is the only counsel egress operation. Current scope, consent, and
// authority are revalidated for each call immediately before a minimal event
// is published.
func (service Service) Escalate(ctx context.Context, request Request) (domain.SafetyEvent, error) {
	request.SessionKey = strings.TrimSpace(request.SessionKey)
	request.ActorKey = strings.TrimSpace(request.ActorKey)
	request.SubjectKey = strings.TrimSpace(request.SubjectKey)
	if service.scope == nil || service.consent == nil || service.authority == nil ||
		service.safety == nil || service.ids == nil || service.now == nil ||
		!opaque(request.SessionKey) || !opaque(request.ActorKey) || !opaque(request.SubjectKey) ||
		request.ConsentVersion == 0 || !request.ExplicitConfirmation {
		return domain.SafetyEvent{}, ErrInvalidRequest
	}
	withinScope, err := service.scope.ContainsBoth(ctx, request.SessionKey, request.ActorKey, request.SubjectKey)
	if err != nil || !withinScope {
		return domain.SafetyEvent{}, ErrUnavailable
	}
	allowed, err := service.consent.CurrentAllows(
		ctx,
		request.SubjectKey,
		SafetyEscalationPurpose,
		request.ConsentVersion,
	)
	if err != nil || !allowed {
		return domain.SafetyEvent{}, ErrUnavailable
	}
	if err := service.authority.AuthorizeEscalation(ctx, request.ActorKey, request.SubjectKey); err != nil {
		return domain.SafetyEvent{}, ErrUnavailable
	}
	event, err := domain.NewSafetyEvent(
		service.ids.NewID(),
		request.SubjectKey,
		domain.ReasonExplicitSafetySupport,
		service.now(),
		request.ConsentVersion,
	)
	if err != nil {
		return domain.SafetyEvent{}, ErrUnavailable
	}
	if err := service.safety.Publish(ctx, event); err != nil {
		return domain.SafetyEvent{}, ErrUnavailable
	}
	return event, nil
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
