package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/room/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Room) error
	Find(context.Context, string) (domain.Room, error)
	FindByCommand(context.Context, string) (domain.Room, error)
	Append(context.Context, domain.Room, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}
type Membership interface {
	RevalidatePair(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
