package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/safety/womensreview/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type DefinitionCatalog interface {
	Current(context.Context) (domain.Definition, error)
}

type AggregateSource interface {
	Current(context.Context, string) (domain.Aggregate, error)
}

type ApprovalSource interface {
	Current(context.Context, string, string) (domain.Approval, error)
}

type ReviewerAuthority interface {
	AuthorizeCurrentWomenReviewer(context.Context, string) error
}

type AssessmentSink interface {
	Record(context.Context, domain.Assessment) error
}

type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
