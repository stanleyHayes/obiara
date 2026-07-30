package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/domain"
)

type Catalog interface {
	SaveApproved(context.Context, domain.Prompt) error
	ListApproved(context.Context, int) ([]domain.Prompt, error)
}

type StoredDuel struct {
	Duel    domain.Duel
	RoomKey string
}

type DuelRepository interface {
	Create(context.Context, StoredDuel, string) error
	Find(context.Context, string) (StoredDuel, error)
	FindByCommand(context.Context, string) (StoredDuel, error)
	Append(context.Context, StoredDuel, uint64, string) error
}

type PairAuthority interface {
	Revalidate(context.Context, string, string, string) error
}

type ReviewerAuthority interface {
	RequireReviewer(context.Context, string) error
}

type Keyer interface {
	Key(string, string) (string, error)
}

type IDSource interface{ NewID() string }
