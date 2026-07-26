package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Export) error
	Find(context.Context, string) (domain.Export, error)
	FindByCommand(context.Context, string) (domain.Export, error)
	Append(context.Context, domain.Export, uint64, string) error
}
type Authority interface {
	RequireMember(context.Context, string, string) error
}
type Allowlist interface {
	Require(context.Context, domain.ReferenceKind, string, domain.Purpose) error
}
type Redactor interface {
	RedactReporterAndThirdParties(context.Context, domain.ReferenceKind, string, domain.Purpose) (string, error)
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
