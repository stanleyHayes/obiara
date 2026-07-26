package visibility

import (
	"context"
	"errors"

	trustapplication "github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

var (
	ErrNotVisible       = errors.New("trust paths not visible")
	ErrInvalidBounds    = errors.New("invalid trust path bounds")
	ErrRequesterMissing = errors.New("requester identity missing")
)

type Projector interface {
	Project(context.Context, trustapplication.ProjectionRequest) (trustapplication.Projection, error)
}

type EdgeReader interface {
	Find(context.Context, string) (domain.Edge, error)
}

type ConsentEvaluator interface {
	Allows(context.Context, string, string) (bool, error)
}

type EndpointAuthorizer interface {
	CanReveal(context.Context, string, string) (bool, error)
}

type RequesterResolver interface {
	RequesterID(context.Context) (string, error)
}
