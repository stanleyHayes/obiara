// Package application is the consent-aware analytics pipeline (E15-S02).
// Events validate against the producer registry, respect the product
// analytics consent-map row (opt-out), and persist pseudonymized —
// subject references are hashed, and nothing free-text can ever validate.
package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/analytics/domain"
)

var ErrAnalyticsOptedOut = errors.New("member opted out of product analytics")

// EventSink persists pseudonymized analytics events.
type EventSink interface {
	Append(context.Context, Event) error
}

// ConsentGate answers the product-analytics consent-map row (opt-out,
// Doc 08 §8). Until the consent registry ships, the composition bridge
// reads consent_records; absent records mean not opted out.
type ConsentGate interface {
	AllowsAnalytics(ctx context.Context, memberID string) (bool, error)
}

// Event is one validated, pseudonymized analytics record.
type Event struct {
	Name       string
	Props      map[string]any
	SubjectRef string
	OccurredAt time.Time
}

// AnalyticsService is the producer boundary.
type AnalyticsService struct {
	sink    EventSink
	consent ConsentGate
	now     func() time.Time
}

func NewAnalyticsService(sink EventSink, consent ConsentGate, now func() time.Time) AnalyticsService {
	return AnalyticsService{sink: sink, consent: consent, now: now}
}

// Emit validates and records one event. Unregistered events, undeclared
// props and free-text values fail here — before any persistence.
func (service AnalyticsService) Emit(ctx context.Context, memberID, name string, props map[string]any) error {
	if err := domain.ValidateProps(name, props); err != nil {
		return err
	}
	if service.consent != nil {
		allowed, err := service.consent.AllowsAnalytics(ctx, memberID)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrAnalyticsOptedOut
		}
	}
	return service.sink.Append(ctx, Event{
		Name:       name,
		Props:      props,
		SubjectRef: Pseudonym(memberID),
		OccurredAt: service.now().UTC(),
	})
}

// Pseudonym derives the pseudonymous subject reference (NFR-402; Doc 08
// retention: pseudonymized at 90 days, aggregated at 13 months).
func Pseudonym(memberID string) string {
	sum := sha256.Sum256([]byte("obiara.analytics.v0:" + memberID))
	return hex.EncodeToString(sum[:])[:32]
}
