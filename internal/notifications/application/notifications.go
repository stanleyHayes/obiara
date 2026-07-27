// Package application enforces notification preferences and the daily cap
// (E13-S01). Dispatchers in other contexts call Decide before sending;
// the cap counter increments atomically so concurrent sends cannot
// overshoot (agent_plan.md §7.4).
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/domain"
)

var ErrPreferencesNotFound = errors.New("notification preferences not found")

// PreferencesRepository persists preferences.
type PreferencesRepository interface {
	Upsert(context.Context, domain.Preferences) error
	FindByMember(context.Context, string) (domain.Preferences, error)
}

// DeliveryCounter tracks per-day delivery counts with atomic claims.
type DeliveryCounter interface {
	// ClaimSlot atomically records a delivery for the member-local date
	// and reports whether it fit under the cap. Concurrent claims cannot
	// exceed the cap.
	ClaimSlot(ctx context.Context, memberID, localDate string, cap int) (claimed bool, err error)
}

// QuieteningChecker reports an active care quietening window (Doc 09 §5:
// 72-hour notification quietening after closure).
type QuieteningChecker interface {
	QuietUntil(ctx context.Context, memberID string, now time.Time) (bool, error)
}

// NotificationService manages preferences and delivery decisions.
type NotificationService struct {
	preferences PreferencesRepository
	counter     DeliveryCounter
	quietening  QuieteningChecker
	now         func() time.Time
}

func NewNotificationService(preferences PreferencesRepository, counter DeliveryCounter, now func() time.Time) NotificationService {
	return NotificationService{preferences: preferences, counter: counter, now: now}
}

// NewNotificationServiceWithQuietening adds the care quietening check.
func NewNotificationServiceWithQuietening(preferences PreferencesRepository, counter DeliveryCounter, quietening QuieteningChecker, now func() time.Time) NotificationService {
	return NotificationService{preferences: preferences, counter: counter, quietening: quietening, now: now}
}

// Get returns the member's preferences, creating defaults on first read.
func (service NotificationService) Get(ctx context.Context, memberID string) (domain.Preferences, error) {
	preferences, err := service.preferences.FindByMember(ctx, memberID)
	if errors.Is(err, ErrPreferencesNotFound) {
		preferences, err = domain.NewPreferences(memberID, service.now())
		if err != nil {
			return domain.Preferences{}, err
		}
		if err := service.preferences.Upsert(ctx, preferences); err != nil {
			return domain.Preferences{}, err
		}
		return preferences, nil
	}
	return preferences, err
}

// Configure validates and stores a member's configuration.
func (service NotificationService) Configure(ctx context.Context, memberID string, muted map[domain.Category]bool, quietStart, quietEnd int, timezone string) (domain.Preferences, error) {
	preferences, err := service.Get(ctx, memberID)
	if err != nil {
		return domain.Preferences{}, err
	}
	if err := preferences.Configure(muted, quietStart, quietEnd, timezone, service.now()); err != nil {
		return domain.Preferences{}, err
	}
	if err := service.preferences.Upsert(ctx, preferences); err != nil {
		return domain.Preferences{}, err
	}
	return preferences, nil
}

// Decide rules on one delivery and, when allowed, claims a daily slot.
// Safety notifications bypass the cap counter entirely (Doc 09 §1).
func (service NotificationService) Decide(ctx context.Context, memberID string, category domain.Category) (domain.Decision, error) {
	preferences, err := service.Get(ctx, memberID)
	if err != nil {
		return domain.Decision{}, err
	}

	decision := preferences.Allows(category, service.now())
	if !decision.Allowed || category == domain.CategorySafety {
		return decision, nil
	}
	if service.quietening != nil {
		quiet, err := service.quietening.QuietUntil(ctx, memberID, service.now())
		if err != nil {
			return domain.Decision{}, err
		}
		if quiet {
			return domain.Decision{Allowed: false, Reason: "care_quietening"}, nil
		}
	}

	claimed, err := service.counter.ClaimSlot(ctx, memberID, preferences.LocalDate(service.now()), domain.DailyCap)
	if err != nil {
		return domain.Decision{}, err
	}
	if !claimed {
		return domain.Decision{Allowed: false, Reason: "daily_cap"}, nil
	}
	return decision, nil
}
