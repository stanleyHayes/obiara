// Package application dispatches due rituals through the preferences
// decision boundary (E13-S01 caps, quiet hours, mutes) and the durable
// outbox, deduplicated per member per ritual per day via the inbox store.
package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/domain"
	ritualdomain "github.com/stanleyHayes/obiara/internal/notifications/ritual/domain"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
)

var ErrNoPreferences = errors.New("member has no notification preferences")

// MemberSource lists active members and their preference time zones.
type MemberSource interface {
	ListActiveIDs(context.Context, int) ([]string, error)
}

// PreferenceReader reads a member's preference time zone.
type PreferenceReader interface {
	TimezoneFor(ctx context.Context, memberID string) (string, error)
}

// Decider is the E13-S01 decision boundary.
type Decider interface {
	Decide(ctx context.Context, memberID string, category domain.Category) (domain.Decision, error)
}

// FireSource lists fires starting inside a herald window with their
// going attendees.
type FireSource interface {
	StartingWithin(ctx context.Context, from, until time.Time) ([]FireWindow, error)
}

// FireWindow is a fire plus its going attendees.
type FireWindow struct {
	FireID    string
	StartsAt  time.Time
	Attendees []string
}

// Dispatcher dispatches due rituals exactly once per member per day.
type Dispatcher struct {
	members     MemberSource
	preferences PreferenceReader
	decider     Decider
	fires       FireSource
	outbox      *outbox.Store
	inbox       *inbox.Store
	now         func() time.Time
}

func NewDispatcher(members MemberSource, preferences PreferenceReader, decider Decider, fires FireSource, outboxStore *outbox.Store, inboxStore *inbox.Store, now func() time.Time) Dispatcher {
	return Dispatcher{
		members:     members,
		preferences: preferences,
		decider:     decider,
		fires:       fires,
		outbox:      outboxStore,
		inbox:       inboxStore,
		now:         now,
	}
}

// DispatchCalendar dispatches every calendar ritual whose local time has
// passed and that has not been dispatched today.
func (dispatcher Dispatcher) DispatchCalendar(ctx context.Context) error {
	members, err := dispatcher.members.ListActiveIDs(ctx, 10000)
	if err != nil {
		return err
	}
	now := dispatcher.now()

	for _, memberID := range members {
		timezone, err := dispatcher.preferences.TimezoneFor(ctx, memberID)
		if err != nil {
			continue // members without preferences keep defaults on next read
		}
		location, err := time.LoadLocation(timezone)
		if err != nil {
			location = time.UTC
		}
		localDate := now.In(location).Format("2006-01-02")

		for _, kind := range ritualdomain.CalendarKinds() {
			dueAt, ok := ritualdomain.DueAt(kind, now, location)
			if !ok || now.Before(dueAt) {
				continue
			}
			if err := dispatcher.dispatch(ctx, memberID, kind, localDate, nil); err != nil {
				return fmt.Errorf("dispatch %s for %s: %w", kind, memberID, err)
			}
		}
	}
	return nil
}

// DispatchHeralds dispatches fire heralds for fires entering the herald
// window.
func (dispatcher Dispatcher) DispatchHeralds(ctx context.Context) error {
	now := dispatcher.now()
	// Look a full day ahead; the dedup key per fire makes re-runs safe.
	fires, err := dispatcher.fires.StartingWithin(ctx, now, now.Add(24*time.Hour))
	if err != nil {
		return err
	}
	for _, fire := range fires {
		if now.Before(ritualdomain.HeraldAt(fire.StartsAt)) {
			continue
		}
		for _, memberID := range fire.Attendees {
			if err := dispatcher.dispatch(ctx, memberID, ritualdomain.KindFireHerald, fire.FireID, map[string]string{"fireId": fire.FireID}); err != nil {
				return fmt.Errorf("dispatch herald %s for %s: %w", fire.FireID, memberID, err)
			}
		}
	}
	return nil
}

func (dispatcher Dispatcher) dispatch(ctx context.Context, memberID string, kind ritualdomain.Kind, discriminator string, extra map[string]string) error {
	// Decide first: a suppressed ritual is not deduplicated, so it still
	// delivers later the same day once quiet hours pass.
	decision, err := dispatcher.decider.Decide(ctx, memberID, domain.CategoryRitual)
	if err != nil {
		return err
	}
	if !decision.Allowed {
		return nil
	}

	key := ritualdomain.DedupKey(memberID, kind, discriminator)
	seen, err := dispatcher.inbox.AlreadyProcessed(ctx, "ritual.dispatcher", key)
	if err != nil || seen {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"memberId":      memberID,
		"kind":          string(kind),
		"discriminator": discriminator,
		"extra":         extra,
	})
	if err != nil {
		return err
	}
	return dispatcher.outbox.Append(ctx, outbox.Record{
		ID:            "ritual_" + string(kind) + "_" + memberID + "_" + discriminator,
		AggregateType: "member",
		AggregateID:   memberID,
		EventType:     "notification.ritual." + string(kind),
		Payload:       payload,
		OccurredAt:    dispatcher.now(),
	})
}
