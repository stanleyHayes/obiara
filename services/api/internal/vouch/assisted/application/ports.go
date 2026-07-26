package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Repository interface {
	Create(context.Context, domain.Request) error
	Find(context.Context, string) (domain.Request, error)
	FindByCommand(context.Context, string) (domain.Request, error)
	Save(context.Context, domain.Request, uint64, string) error
}

type Authorizer interface {
	Require(ctx context.Context, actorID, action, requestID string) error
}

type Keyer interface {
	Key(namespace, value string) (string, error)
}

type IDSource interface {
	NewID() string
}
