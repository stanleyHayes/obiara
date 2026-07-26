// Package reactivation lifts expired account suspensions on a schedule
// (E12-S04): Tier-B suspensions end automatically at their timestamp on
// server time. The adapter mirrors the identity account document; if a
// third worker consumer of accounts appears, the account repository
// should move to internal/platform.
package reactivation

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	jobsapplication "github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

type Store struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewStore(database *mongo.Database, clock func() time.Time) *Store {
	return &Store{database: database, clock: clock}
}

func (store *Store) collection() *mongo.Collection {
	return store.database.Collection("accounts")
}

// ReactivateExpired lifts every suspension past its end time and returns
// the count.
func (store *Store) ReactivateExpired(ctx context.Context) (int, error) {
	cursor, err := store.collection().Find(ctx,
		bson.M{"status": "suspended", "suspendedUntil": bson.M{"$lte": store.clock().UTC()}},
		options.Find().SetProjection(bson.M{"_id": 1, "version": 1}).SetLimit(500))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var document struct {
			ID      string `bson:"_id"`
			Version int64  `bson:"version"`
		}
		if err := cursor.Decode(&document); err != nil {
			return count, err
		}
		result, err := store.collection().UpdateOne(ctx,
			bson.M{"_id": document.ID, "version": document.Version},
			bson.M{
				"$set":   bson.M{"status": "active", "version": document.Version + 1},
				"$unset": bson.M{"suspendedUntil": ""},
			})
		if err != nil {
			return count, err
		}
		if result.ModifiedCount == 1 {
			count++
		}
	}
	return count, cursor.Err()
}

// NewJob builds the scheduled suspension reactivation job.
func NewJob(store *Store, interval time.Duration) jobsapplication.Job {
	return jobsapplication.Job{
		Name:        "safety.reactivation",
		Interval:    interval,
		Timeout:     time.Minute,
		MaxAttempts: 5,
		Run: func(ctx context.Context) error {
			_, err := store.ReactivateExpired(ctx)
			return err
		},
	}
}
