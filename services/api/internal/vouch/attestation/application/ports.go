package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Repository interface {
	Create(context.Context, domain.Attestation) error
	Find(context.Context, string) (domain.Attestation, error)
	FindByCommand(context.Context, string) (domain.Attestation, error)
	Append(context.Context, domain.Attestation, uint64, string) error
}
type Authorizer interface {
	Require(context.Context, string, string, string) error
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type StakePolicy interface {
	Validate(context.Context, string, uint8) (string, error)
}
type IDSource interface{ NewID() string }
