// Package trust owns directed trust edges and bounded, owner-authorized path
// projection. It intentionally exposes no global graph enumeration.
package trust

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/trust/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/visibility"
)

type Module struct {
	Edges      *mongodb.Repository
	Projection application.Service
	Visibility visibility.Service
}

func NewModule(
	ctx context.Context,
	database *mongo.Database,
	authorizer application.ProjectionAuthorizer,
	consent application.ConsentEvaluator,
	endpoints visibility.EndpointAuthorizer,
) (Module, error) {
	edges := mongodb.NewRepository(database)
	if err := edges.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	projection := application.NewService(edges, authorizer, consent, nil)
	disclosure := visibility.NewDisclosurePolicy(edges, consent, endpoints, nil)
	return Module{
		Edges: edges, Projection: projection,
		Visibility: visibility.NewService(projection, disclosure),
	}, nil
}
