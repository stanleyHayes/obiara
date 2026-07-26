package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Repository interface {
	Create(context.Context, domain.Request) error
	Find(context.Context, string) (domain.Request, error)
	FindByCommand(context.Context, string) (domain.Request, error)
	Append(context.Context, domain.Request, uint64, string) error
}

type Authorizer interface {
	Require(context.Context, string, string, domain.SourceType, string) error
}

type SourcePolicy interface {
	Allow(context.Context, domain.SourceType, string) error
}

// CandidateResolver is intentionally source-scoped and bounded. It must never
// expose a global member list, reverse graph, or profile payload.
type CandidateResolver interface {
	CandidateIDs(context.Context, domain.SourceType, string, int) ([]string, error)
}

type ConsentVisibility interface {
	Visible(context.Context, string, string, domain.SourceType, string) (bool, error)
}

type Keyer interface {
	Key(namespace, value string) (string, error)
}

type IDSource interface{ NewID() string }
