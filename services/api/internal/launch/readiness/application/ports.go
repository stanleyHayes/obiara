package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Authority interface {
	RequireLaunchReviewer(context.Context, string) error
}
type FamilyProjection interface {
	CurrentFamilyDensity(context.Context, string) (domain.FamilyDensity, error)
}
type HostProjection interface {
	CurrentHostCoverage(context.Context, string) (domain.HostCoverage, error)
}
type LicenseProjection interface {
	CurrentLicenseCoverage(context.Context, string, string) (domain.LicenseCoverage, error)
}
type Repository interface {
	Create(context.Context, domain.Snapshot) error
	Find(context.Context, string) (domain.Snapshot, error)
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
