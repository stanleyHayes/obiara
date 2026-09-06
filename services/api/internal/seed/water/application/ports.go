package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Water) error
	Find(context.Context, string) (domain.Water, error)
	FindByCommand(context.Context, string) (domain.Water, error)
	Append(context.Context, domain.Water, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}

// PairConsent asks whether these two people may still be brought together —
// a block, a withdrawal, anything that has happened since the water started.
//
// Both arguments are RAW member ids, never the water's keys. The water keys
// its members under its own secret, so a key handed to this port would be
// compared against nothing and every call would silently pass. That is what
// used to happen on the mutual step (agent_plan.md §62): Start passed raw ids
// and Water passed keys, so a block placed after the water began was never
// honoured.
type PairConsent interface {
	Revalidate(ctx context.Context, memberID, otherMemberID string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
