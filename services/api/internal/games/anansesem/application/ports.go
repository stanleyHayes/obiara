package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Story) error
	Find(context.Context, string) (domain.Story, error)
	FindByCommand(context.Context, string) (domain.Story, error)
	Append(context.Context, domain.Story, uint64, string) error
}
type Authority interface {
	RevalidateAuthors(context.Context, string, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
