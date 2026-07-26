package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Repository interface {
	CreateInvite(context.Context, domain.Invite) error
	FindInviteByDigest(context.Context, string) (domain.Invite, error)
	FindInviteByCommand(context.Context, string) (domain.Invite, error)
	CreateRequest(context.Context, domain.Request) error
	FindRequest(context.Context, string) (domain.Request, error)
	FindRequestByCommand(context.Context, string) (domain.Request, error)
	SaveRequest(context.Context, domain.Request, uint64, string) error
	Redeem(context.Context, domain.Invite, domain.Request, uint64, string) error
}

type Authorizer interface {
	Require(ctx context.Context, actorID, circleID, action, memberID string) error
}

type TokenIssuer interface {
	NewToken() (raw string, digest string, err error)
	Digest(raw string) (string, error)
}

type IDSource interface {
	NewID() string
}
