package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/domain"
)

type Command struct {
	ID         string
	DeclinerID string
	SeedID     string
	OwnerID    string
}

type Result struct {
	Replayed bool
}

type Service struct {
	store Store
	keyer Keyer
	ids   IDSource
	now   func() time.Time
}

func New(store Store, keyer Keyer, ids IDSource, now func() time.Time) Service {
	return Service{store: store, keyer: keyer, ids: ids, now: now}
}

func (service Service) Decline(ctx context.Context, command Command) (Result, error) {
	if service.store == nil || service.keyer == nil || service.ids == nil || service.now == nil {
		return Result{}, ErrUnavailable
	}
	if strings.TrimSpace(command.ID) == "" || strings.TrimSpace(command.DeclinerID) == "" ||
		strings.TrimSpace(command.SeedID) == "" || strings.TrimSpace(command.OwnerID) == "" {
		return Result{}, domain.ErrInvalid
	}
	declinerKey, err := service.keyer.Key("decliner", command.DeclinerID)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	seedKey, err := service.keyer.Key("seed", command.SeedID)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	recipientKey, err := service.keyer.Key("notification-recipient", command.OwnerID)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	eventKey, err := service.keyer.Key("notification-event", command.ID)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	now := service.now().UTC()
	decline, err := domain.New(service.ids.NewID(), declinerKey, seedKey, recipientKey, command.ID, now)
	if err != nil {
		return Result{}, err
	}
	notification, err := domain.NewNotification(eventKey, recipientKey, now)
	if err != nil {
		return Result{}, err
	}
	_, replayed, err := service.store.Record(ctx, decline, notification)
	if err != nil {
		if errors.Is(err, domain.ErrCommandMismatch) {
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{}, ErrUnavailable
	}
	return Result{Replayed: replayed}, nil
}

// Locked reports whether a decline still shields the target from this sower.
//
// It keys each side the way Decline recorded it: the target under "decliner",
// because they are the one who said no, and the sower under
// "notification-recipient", because that is the namespace the seed owner was
// keyed under when the decline was written. Reusing those namespaces is what
// makes every decline already on record enforceable without a backfill.
//
// Like Eligible, it returns a bool and nothing else. A caller must not be
// able to learn that a decline happened, only that this reach is not
// available (FR-205).
func (service Service) Locked(ctx context.Context, sowerID, targetID string, at time.Time) (bool, error) {
	if service.store == nil || service.keyer == nil || at.IsZero() {
		return false, ErrUnavailable
	}
	declinerKey, err := service.keyer.Key("decliner", targetID)
	if err != nil {
		return false, ErrUnavailable
	}
	sowerKey, err := service.keyer.Key("notification-recipient", sowerID)
	if err != nil {
		return false, ErrUnavailable
	}
	locked, err := service.store.IsPairExcluded(ctx, declinerKey, sowerKey, at.UTC())
	if err != nil {
		return false, ErrUnavailable
	}
	return locked, nil
}

// Eligible is the only selection-facing operation. It returns no reason,
// timestamp, owner, or rejection signal.
func (service Service) Eligible(ctx context.Context, declinerID, seedID string, at time.Time) (bool, error) {
	if service.store == nil || service.keyer == nil || at.IsZero() {
		return false, ErrUnavailable
	}
	declinerKey, err := service.keyer.Key("decliner", declinerID)
	if err != nil {
		return false, ErrUnavailable
	}
	seedKey, err := service.keyer.Key("seed", seedID)
	if err != nil {
		return false, ErrUnavailable
	}
	excluded, err := service.store.IsExcluded(ctx, declinerKey, seedKey, at.UTC())
	if err != nil {
		return false, ErrUnavailable
	}
	return !excluded, nil
}
