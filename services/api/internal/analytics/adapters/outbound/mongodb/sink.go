// Package mongodb persists pseudonymized analytics events (append-only;
// pseudonymized at 90 days, aggregated at 13 months per Doc 09 §7).
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/analytics/application"
)

type Sink struct {
	database *mongo.Database
}

func NewSink(database *mongo.Database) *Sink {
	return &Sink{database: database}
}

func (sink *Sink) collection() *mongo.Collection {
	return sink.database.Collection("analytics_events")
}

type document struct {
	Name       string         `bson:"name"`
	Props      map[string]any `bson:"props,omitempty"`
	SubjectRef string         `bson:"subjectRef"`
	OccurredAt time.Time      `bson:"occurredAt"`
}

func (sink *Sink) EnsureIndexes(ctx context.Context) error {
	_, err := sink.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}, {Key: "occurredAt", Value: -1}},
			Options: options.Index().SetName("analytics_name_time"),
		},
		{
			Keys:    bson.D{{Key: "subjectRef", Value: 1}, {Key: "occurredAt", Value: -1}},
			Options: options.Index().SetName("analytics_subject_time"),
		},
	})
	return err
}

func (sink *Sink) Append(ctx context.Context, event application.Event) error {
	_, err := sink.collection().InsertOne(ctx, document{
		Name:       event.Name,
		Props:      event.Props,
		SubjectRef: event.SubjectRef,
		OccurredAt: event.OccurredAt.UTC(),
	})
	return err
}

// CountByName supports pipeline assertions in tests and dashboards.
func (sink *Sink) CountByName(ctx context.Context, name string) (int, error) {
	count, err := sink.collection().CountDocuments(ctx, bson.M{"name": name})
	return int(count), err
}
