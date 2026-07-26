// Package application runs fire scheduling and attendance (E09-S01).
// Capacity is race-safe: the fire counter and the attendance record commit
// in one transaction, and waitlist promotion on cancellation is atomic.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
)

var (
	ErrFireNotFound  = errors.New("fire not found")
	ErrRSVPNotFound  = errors.New("rsvp not found")
	ErrAlreadyRSVPed = errors.New("member already has an rsvp")
)

// FireRepository persists fires and attendance with transactional
// capacity/promotion semantics.
type FireRepository interface {
	Create(context.Context, domain.Fire) error
	FindByID(context.Context, string) (domain.Fire, error)
	// AdmitTx atomically claims a seat or waitlist position and inserts the
	// attendance record. It implements the domain capacity rule: open
	// fires only, going until capacity, then waitlist in order.
	AdmitTx(ctx context.Context, fireID, memberID string, now time.Time) (domain.RSVP, error)
	// CancelTx removes the RSVP and atomically promotes the first
	// waitlisted member when a going seat freed. It returns the promotion
	// when one happened.
	CancelTx(ctx context.Context, fireID, memberID string, now time.Time) (*domain.RSVP, error)
	FindRSVP(context.Context, string, string) (domain.RSVP, error)
	ListUpcoming(context.Context, time.Time, int) ([]domain.Fire, error)
}

// FireService schedules fires and manages attendance.
type FireService struct {
	fires FireRepository
	now   func() time.Time
	newID func() string
}

func NewFireService(fires FireRepository, now func() time.Time, newID func() string) FireService {
	return FireService{fires: fires, now: now, newID: newID}
}

// Schedule creates a fire.
func (service FireService) Schedule(ctx context.Context, hostID, circleID, title string, startsAt time.Time, capacity int) (domain.Fire, error) {
	fire, err := domain.NewFire(service.newID(), hostID, circleID, title, startsAt, capacity, service.now())
	if err != nil {
		return domain.Fire{}, err
	}
	if err := service.fires.Create(ctx, fire); err != nil {
		return domain.Fire{}, err
	}
	return fire, nil
}

// RSVP admits a Tier 1+ member (FR-401). Capacity races converge
// atomically in the repository; no application retry loop is needed.
func (service FireService) RSVP(ctx context.Context, fireID, memberID string, tier int) (domain.RSVP, error) {
	if tier < 1 {
		return domain.RSVP{}, domain.ErrTierTooLow
	}
	return service.fires.AdmitTx(ctx, fireID, memberID, service.now())
}

// Cancel removes an RSVP and promotes the first waitlisted member when a
// seat freed.
func (service FireService) Cancel(ctx context.Context, fireID, memberID string) (promoted *domain.RSVP, err error) {
	return service.fires.CancelTx(ctx, fireID, memberID, service.now())
}

// Upcoming lists scheduled fires from now onward.
func (service FireService) Upcoming(ctx context.Context, limit int) ([]domain.Fire, error) {
	return service.fires.ListUpcoming(ctx, service.now(), limit)
}
