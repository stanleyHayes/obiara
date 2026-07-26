package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Item) error
	Find(context.Context, string, string) (domain.Item, error)
	Save(context.Context, domain.Item, uint64) error
	ListOwner(context.Context, string) ([]domain.Item, error)
	ExpireDue(context.Context, string, time.Time, int) (int64, error)
}

type Keyer interface {
	Key(string, string) (string, error)
}

var ErrNotFound = domain.ErrInvalidProjection
