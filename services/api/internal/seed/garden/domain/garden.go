package domain

import (
	"errors"
	"regexp"
	"time"
)

type State string

const (
	StateQueued    State = "queued"
	StateDelivered State = "delivered"
	StateHeard     State = "heard"
	StateSprouted  State = "sprouted"
	StateDeclined  State = "declined"
	StateExpired   State = "expired"
)

var (
	ErrInvalidProjection = errors.New("invalid garden projection")
	ErrInvalidTransition = errors.New("invalid garden transition")
	ErrStaleRevision     = errors.New("stale garden revision")
)

var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Item struct {
	SeedKey, OwnerKey string
	State             State
	ExpiresAt         time.Time
	UpdatedAt         time.Time
	Revision          uint64
}

func New(seedKey, ownerKey string, expiresAt, at time.Time) (Item, error) {
	if !opaque.MatchString(seedKey) || !opaque.MatchString(ownerKey) || at.IsZero() ||
		!expiresAt.After(at) || expiresAt.After(at.Add(90*24*time.Hour)) {
		return Item{}, ErrInvalidProjection
	}
	return Item{
		SeedKey: seedKey, OwnerKey: ownerKey, State: StateQueued,
		ExpiresAt: expiresAt.UTC(), UpdatedAt: at.UTC(), Revision: 1,
	}, nil
}

func Rehydrate(item Item) (Item, error) {
	if !opaque.MatchString(item.SeedKey) || !opaque.MatchString(item.OwnerKey) ||
		!validState(item.State) || item.ExpiresAt.IsZero() || item.UpdatedAt.IsZero() ||
		item.Revision == 0 {
		return Item{}, ErrInvalidProjection
	}
	item.ExpiresAt = item.ExpiresAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

func (item Item) Transition(next State, expected uint64, at time.Time) (Item, error) {
	if expected != item.Revision {
		return Item{}, ErrStaleRevision
	}
	if at.Before(item.UpdatedAt) || at.IsZero() || !allowed(item.State, next, at, item.ExpiresAt) {
		return Item{}, ErrInvalidTransition
	}
	item.State = next
	item.UpdatedAt = at.UTC()
	item.Revision++
	return item, nil
}

func allowed(current, next State, at, expiresAt time.Time) bool {
	if next == StateExpired {
		return !terminal(current) && !at.Before(expiresAt)
	}
	if !at.Before(expiresAt) {
		return false
	}
	switch current {
	case StateQueued:
		return next == StateDelivered || next == StateDeclined
	case StateDelivered:
		return next == StateHeard || next == StateDeclined
	case StateHeard:
		return next == StateSprouted || next == StateDeclined
	default:
		return false
	}
}

func validState(state State) bool {
	return state == StateQueued || state == StateDelivered || state == StateHeard ||
		state == StateSprouted || state == StateDeclined || state == StateExpired
}

func terminal(state State) bool {
	return state == StateSprouted || state == StateDeclined || state == StateExpired
}

type Summary struct {
	AsOf                   time.Time
	MovingQuietly, Sprouts int
	Message                string
}

func Summarize(items []Item, asOf time.Time) (Summary, error) {
	if asOf.IsZero() {
		return Summary{}, ErrInvalidProjection
	}
	result := Summary{AsOf: asOf.UTC()}
	for _, item := range items {
		if _, err := Rehydrate(item); err != nil {
			return Summary{}, err
		}
		switch item.State {
		case StateQueued, StateDelivered, StateHeard:
			result.MovingQuietly++
		case StateSprouted:
			result.Sprouts++
		}
	}
	switch {
	case result.Sprouts > 0:
		result.Message = "A doorway is ready when you are."
	case result.MovingQuietly > 0:
		result.Message = "Your seeds are moving quietly."
	default:
		result.Message = "Nothing needs your attention today."
	}
	return result, nil
}
