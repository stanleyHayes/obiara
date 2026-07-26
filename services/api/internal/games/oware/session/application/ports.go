package application

import (
	"context"
	session "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, session.Session) error
	Find(context.Context, string) (session.Session, error)
	FindByCommand(context.Context, string) (session.Session, error)
	Append(context.Context, session.Session, uint64, string) error
}
type RoomEmbedding interface {
	Revalidate(context.Context, string, string, string) error
}
type Authorizer interface {
	RequireParticipant(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
