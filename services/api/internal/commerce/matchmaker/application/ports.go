package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Engagement) error
	Find(context.Context, string) (domain.Engagement, error)
	FindByCommand(context.Context, string) (domain.Engagement, error)
	ListForMember(context.Context, string) ([]domain.Engagement, error)
	Save(context.Context, domain.Engagement, uint64, string) error
}
type LicenseCatalog interface {
	Current(context.Context, string) (domain.License, error)
	ListCurrent(context.Context, time.Time) ([]domain.LicensedProfile, error)
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
