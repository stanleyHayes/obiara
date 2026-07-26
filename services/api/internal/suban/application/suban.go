// Package application runs the suban ledger (E15-S04): append-only events
// with anti-gaming period caps, and marks recomputed from the full ledger
// on every read — there is no cached authority (Doc 08 §4).
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
)

var ErrPeriodCapReached = errors.New("suban event period cap reached")

// EventStore is the append-only ledger.
type EventStore interface {
	Append(context.Context, domain.Event) error
	ListForSubject(context.Context, string) ([]domain.Event, error)
	CountForSubjectSince(context.Context, string, domain.Kind, time.Time) (int, error)
}

// SubanService records events and recomputes marks.
type SubanService struct {
	events EventStore
	now    func() time.Time
	newID  func() string
}

func NewSubanService(events EventStore, now func() time.Time, newID func() string) SubanService {
	return SubanService{events: events, now: now, newID: newID}
}

// Record appends one event, enforcing the anti-gaming period cap.
func (service SubanService) Record(ctx context.Context, subjectID string, kind domain.Kind, provenance domain.Provenance) error {
	since := service.now().Add(-domain.CapWindow)
	count, err := service.events.CountForSubjectSince(ctx, subjectID, kind, since)
	if err != nil {
		return err
	}
	if count >= domain.PeriodCap {
		return ErrPeriodCapReached
	}
	return service.events.Append(ctx, domain.Event{
		ID:         service.newID(),
		SubjectID:  subjectID,
		Kind:       kind,
		Provenance: provenance,
		OccurredAt: service.now().UTC(),
	})
}

// Marks recomputes the member's thresholded marks from the ledger.
func (service SubanService) Marks(ctx context.Context, subjectID string) ([]domain.Mark, error) {
	events, err := service.events.ListForSubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	return domain.ComputeMarks(events, service.now()), nil
}

// Events returns the member-visible ledger (Doc 08 §4: members can view
// every event behind their own marks).
func (service SubanService) Events(ctx context.Context, subjectID string) ([]domain.Event, error) {
	return service.events.ListForSubject(ctx, subjectID)
}
