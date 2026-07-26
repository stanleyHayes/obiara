// Package idempotency implements the write-command idempotency store
// (agent_plan.md §7.4: every write command carries an idempotency key).
// The first claim of a (scope, key) pair wins; retries of the same command
// observe the recorded outcome instead of re-executing.
package idempotency

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	apimongo "github.com/stanleyHayes/obiara/services/api/internal/platform/mongo"
)

var (
	ErrScopeRequired = errors.New("idempotency scope is required")
	ErrKeyRequired   = errors.New("idempotency key is required")
)

// Status is the lifecycle of a claimed command.
type Status string

const (
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
)

type document struct {
	ID         string    `bson:"_id"`
	Scope      string    `bson:"scope"`
	Key        string    `bson:"key"`
	Status     Status    `bson:"status"`
	ResultRef  string    `bson:"resultRef,omitempty"`
	ClaimedAt  time.Time `bson:"claimedAt"`
	FinishedAt time.Time `bson:"finishedAt,omitempty"`
}

type Store struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewStore(database *mongo.Database, clock func() time.Time) *Store {
	return &Store{database: database, clock: clock}
}

func (store *Store) collection() *mongo.Collection {
	return store.database.Collection("idempotency_keys")
}

// Claim reserves the key within a command scope (e.g. "member.register").
// claimed=false means the command is a replay; the caller must look up the
// recorded outcome instead of re-executing.
func (store *Store) Claim(ctx context.Context, scope, key string) (claimed bool, err error) {
	if scope == "" {
		return false, ErrScopeRequired
	}
	if key == "" {
		return false, ErrKeyRequired
	}
	_, err = store.collection().InsertOne(ctx, document{
		ID:        scope + "|" + key,
		Scope:     scope,
		Key:       key,
		Status:    StatusProcessing,
		ClaimedAt: store.clock().UTC(),
	})
	if apimongo.IsDuplicateKey(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Complete records the outcome reference (e.g. the created aggregate ID)
// for replay responses.
func (store *Store) Complete(ctx context.Context, scope, key, resultRef string) error {
	result, err := store.collection().UpdateOne(ctx,
		bson.M{"_id": scope + "|" + key},
		bson.M{"$set": bson.M{"status": StatusCompleted, "resultRef": resultRef, "finishedAt": store.clock().UTC()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("idempotency key %q in scope %q not claimed", key, scope)
	}
	return nil
}

// ResultRef returns the recorded outcome for a replayed command.
func (store *Store) ResultRef(ctx context.Context, scope, key string) (string, error) {
	var doc document
	if err := store.collection().FindOne(ctx, bson.M{"_id": scope + "|" + key}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", fmt.Errorf("idempotency key %q in scope %q not claimed", key, scope)
		}
		return "", err
	}
	return doc.ResultRef, nil
}
