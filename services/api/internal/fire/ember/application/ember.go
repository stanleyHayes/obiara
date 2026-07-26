// Package application issues and redeems embers (E06-S10, FR-402).
// Co-attendance is verified against fire attendance; one ember per
// attendee per fire is enforced by unique index; mutual detection opens a
// doorway through the seed-context port when it is composed.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/domain"
)

var (
	ErrEmberNotFound     = errors.New("ember not found")
	ErrEmberAlreadyGiven = errors.New("attendee already gave an ember at this fire")
	ErrNotCoAttendee     = errors.New("embers require co-attendance at the same fire")
	ErrNotRecipient      = errors.New("only the recipient can redeem an ember")
)

// EmberRepository persists embers.
type EmberRepository interface {
	Create(context.Context, domain.Ember) error
	FindByID(context.Context, string) (domain.Ember, error)
	// FindDirected returns the from→to ember for a fire when it exists.
	FindDirected(context.Context, string, string, string) (domain.Ember, error)
	Update(context.Context, domain.Ember) error
}

// AttendanceChecker verifies co-attendance (fire context read model).
type AttendanceChecker interface {
	Attended(ctx context.Context, fireID, memberID string) (bool, error)
}

// DoorwayOpener opens an instant doorway for a mutual ember pair (seed
// context port; composed when the sprout module is wired).
type DoorwayOpener interface {
	OpenForPair(ctx context.Context, fromID, toID, reference string) error
}

// EmberService issues and redeems embers.
type EmberService struct {
	embers     EmberRepository
	attendance AttendanceChecker
	opener     DoorwayOpener // nil until the sprout module composes
	now        func() time.Time
	newID      func() string
}

func NewEmberService(embers EmberRepository, attendance AttendanceChecker, opener DoorwayOpener, now func() time.Time, newID func() string) EmberService {
	return EmberService{embers: embers, attendance: attendance, opener: opener, now: now, newID: newID}
}

// Issue gives an ember from one co-attendee to another (FR-402).
func (service EmberService) Issue(ctx context.Context, fireID, fromID, toID string) (domain.Ember, error) {
	for _, memberID := range []string{fromID, toID} {
		attended, err := service.attendance.Attended(ctx, fireID, memberID)
		if err != nil {
			return domain.Ember{}, err
		}
		if !attended {
			return domain.Ember{}, ErrNotCoAttendee
		}
	}

	ember, err := domain.NewEmber(service.newID(), fireID, fromID, toID, service.now())
	if err != nil {
		return domain.Ember{}, err
	}
	if err := service.embers.Create(ctx, ember); err != nil {
		return domain.Ember{}, err
	}

	// Mutual detection: the reverse ember already exists.
	reverse, err := service.embers.FindDirected(ctx, fireID, toID, fromID)
	if errors.Is(err, ErrEmberNotFound) {
		return ember, nil
	}
	if err != nil {
		return domain.Ember{}, err
	}
	if reverse.Status() == domain.StatusIssued {
		reverse.MarkMutual()
		if err := service.embers.Update(ctx, reverse); err != nil {
			return domain.Ember{}, err
		}
	}
	ember.MarkMutual()
	if err := service.embers.Update(ctx, ember); err != nil {
		return domain.Ember{}, err
	}

	if service.opener != nil {
		if err := service.opener.OpenForPair(ctx, fromID, toID, "ember:"+ember.ID()); err != nil {
			return domain.Ember{}, err
		}
	}
	return ember, nil
}

// Redeem closes an ember inside its 24-hour window (FR-402). Only the
// recipient may redeem.
func (service EmberService) Redeem(ctx context.Context, emberID, memberID string) (domain.Ember, error) {
	ember, err := service.embers.FindByID(ctx, emberID)
	if err != nil {
		return domain.Ember{}, err
	}
	if memberID != ember.ToID() {
		return domain.Ember{}, ErrNotRecipient
	}
	redeemErr := ember.Redeem(service.now())
	if err := service.embers.Update(ctx, ember); err != nil {
		return domain.Ember{}, err
	}
	return ember, redeemErr
}
