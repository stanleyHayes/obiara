// Package mongodb implements the analytics read ports.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// QueryStore answers funnel queries over analytics_events.
type QueryStore struct {
	database *mongo.Database
}

func NewQueryStore(database *mongo.Database) *QueryStore {
	return &QueryStore{database: database}
}

func (store *QueryStore) CountEvents(ctx context.Context, name string, since time.Time) (int, error) {
	count, err := store.database.Collection("analytics_events").CountDocuments(ctx,
		bson.M{"name": name, "occurredAt": bson.M{"$gte": since.UTC()}})
	return int(count), err
}

func (store *QueryStore) CountDistinctSubjects(ctx context.Context, name string, since time.Time) (int, error) {
	var subjects []string
	err := store.database.Collection("analytics_events").Distinct(ctx, "subjectRef",
		bson.M{"name": name, "occurredAt": bson.M{"$gte": since.UTC()}}).Decode(&subjects)
	if err != nil {
		return 0, err
	}
	return len(subjects), nil
}

// CohortStore reports the active cohort (identity read model).
type CohortStore struct {
	database *mongo.Database
}

func NewCohortStore(database *mongo.Database) *CohortStore {
	return &CohortStore{database: database}
}

func (store *CohortStore) CountActive(ctx context.Context) (int, error) {
	count, err := store.database.Collection("accounts").CountDocuments(ctx, bson.M{"status": "active"})
	return int(count), err
}
