// Package inbox implements consumer-side dedup records (agent_plan.md §7.4):
// every asynchronous consumer is at-least-once safe because the first
// processed delivery of a message wins and redeliveries are no-ops.
package inbox

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	apimongo "github.com/stanleyHayes/obiara/services/api/internal/platform/mongo"
)

var (
	ErrConsumerRequired  = errors.New("inbox consumer name is required")
	ErrMessageIDRequired = errors.New("inbox message id is required")
)

type document struct {
	ID          string    `bson:"_id"`
	Consumer    string    `bson:"consumer"`
	MessageID   string    `bson:"messageId"`
	ProcessedAt time.Time `bson:"processedAt"`
}

type Store struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewStore(database *mongo.Database, clock func() time.Time) *Store {
	return &Store{database: database, clock: clock}
}

func (store *Store) collection() *mongo.Collection {
	return store.database.Collection("inbox")
}

// AlreadyProcessed records the (consumer, messageID) pair and reports
// whether it had been processed before. The unique _id makes concurrent
// redeliveries safe: exactly one caller sees false.
func (store *Store) AlreadyProcessed(ctx context.Context, consumer, messageID string) (bool, error) {
	if consumer == "" {
		return false, ErrConsumerRequired
	}
	if messageID == "" {
		return false, ErrMessageIDRequired
	}
	_, err := store.collection().InsertOne(ctx, document{
		ID:          consumer + "|" + messageID,
		Consumer:    consumer,
		MessageID:   messageID,
		ProcessedAt: store.clock().UTC(),
	})
	if apimongo.IsDuplicateKey(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

// Forget removes a dedup record. Reserved for dead-letter replay tooling;
// normal consumers never call it.
func (store *Store) Forget(ctx context.Context, consumer, messageID string) error {
	_, err := store.collection().DeleteOne(ctx, bson.M{"_id": consumer + "|" + messageID})
	return err
}
