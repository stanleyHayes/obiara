package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Escrow) error
	CreateAudited(context.Context, domain.Escrow, string) error
	Find(context.Context, string) (domain.Escrow, error)
	FindByCommand(context.Context, string) (domain.Escrow, error)
	ListForOwner(context.Context, string) ([]domain.Escrow, error)
	Save(context.Context, domain.Escrow, uint64, string) error
	SaveAudited(context.Context, domain.Escrow, uint64, string, string, string) error
	SettleAudited(context.Context, domain.Escrow, uint64, string, string, domain.Statement) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
