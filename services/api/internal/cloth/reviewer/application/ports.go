package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Review) error
	Find(context.Context, string) (domain.Review, error)
	FindByCommand(context.Context, string) (domain.Review, error)
	Append(context.Context, domain.Review, uint64, string) error
}

type Authorizer interface {
	Require(context.Context, string, string, string) error
}

type PairPolicy interface {
	Revalidate(context.Context, string, string) error
}

type Keyer interface {
	Key(string, string) (string, error)
}

type TokenSource interface {
	Token() (string, error)
}

type Watermarker interface {
	Ref(string, string) (string, error)
}

type IDSource interface {
	NewID() string
}
