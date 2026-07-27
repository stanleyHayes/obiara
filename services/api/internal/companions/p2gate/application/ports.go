package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/p2gate/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type ConsentSource interface {
	CurrentGateConsent(context.Context, string) (domain.GateConsent, error)
}

type CompanionSource interface {
	CurrentCompanionFacts(context.Context, string) (domain.CompanionFacts, error)
}

type SessionAuthenticator interface {
	Authenticate(context.Context, string, string) error
}

type Repository interface {
	Create(context.Context, domain.Proposal) error
}

type IDSource interface {
	NewID() string
	NewTokenRef() string
	NewWatermarkRef() string
}

type Clock interface{ Now() time.Time }
