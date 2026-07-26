// Package mongodb persists job dead letters for operator triage.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

type DeadLetterStore struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewDeadLetterStore(database *mongo.Database, clock func() time.Time) *DeadLetterStore {
	return &DeadLetterStore{database: database, clock: clock}
}

func (store *DeadLetterStore) collection() *mongo.Collection {
	return store.database.Collection("job_dead_letters")
}

type document struct {
	JobName    string    `bson:"jobName"`
	Reason     string    `bson:"reason"`
	Failures   int       `bson:"failures"`
	OccurredAt time.Time `bson:"occurredAt"`
}

func (store *DeadLetterStore) Record(ctx context.Context, letter application.DeadLetter) error {
	_, err := store.collection().InsertOne(ctx, document{
		JobName:    letter.JobName,
		Reason:     letter.Reason,
		Failures:   letter.Failures,
		OccurredAt: letter.OccurredAt.UTC(),
	})
	return err
}
