// Package analytics is the composition root of the Analytics and Audit
// bounded context slice for the producer-enforced event pipeline
// (E15-S01/S02). Dashboards and retention jobs land with E15-S03/S08.
package analytics

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/analytics/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/application"
)

type Module struct {
	Analytics application.AnalyticsService
}

// NewModule builds the pipeline. consent may be nil to run without the
// consent gate until the registry bridge is composed.
func NewModule(ctx context.Context, database *mongo.Database, consent application.ConsentGate) (Module, error) {
	sink := mongodb.NewSink(database)
	if err := sink.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Analytics: application.NewAnalyticsService(sink, consent, time.Now),
	}, nil
}
