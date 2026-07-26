package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/games/conduct/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Signal) error
	Find(context.Context, string) (domain.Signal, error)
	FindByCommand(context.Context, string) (domain.Signal, error)
	Append(context.Context, domain.Signal, uint64, string) error
}
type Authority interface {
	RequireSubject(context.Context, string, string, string) error
	RequireAppealReviewer(context.Context, string) error
}
type EventVerifier interface {
	Revalidate(context.Context, string, string, string, domain.GameEvent) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
