// Package outbox implements the durable outbox (agent_plan.md §7.4): event
// records are committed in the same transaction as the domain change and
// processed asynchronously by the worker relay (S1-009). Consumers are
// safe against at-least-once delivery via the inbox package.
package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Record is one domain event awaiting publication. Payload is the
// provider-neutral serialized event; it must never contain raw conversation
// content beyond the producing context's consent rules.
type Record struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	OccurredAt    time.Time
}

type document struct {
	ID            string     `bson:"_id"`
	AggregateType string     `bson:"aggregateType"`
	AggregateID   string     `bson:"aggregateId"`
	EventType     string     `bson:"eventType"`
	Payload       []byte     `bson:"payload"`
	OccurredAt    time.Time  `bson:"occurredAt"`
	PublishedAt   *time.Time `bson:"publishedAt"`
	Attempts      int        `bson:"attempts"`
}

var (
	ErrIDRequired            = errors.New("outbox record id is required")
	ErrAggregateTypeRequired = errors.New("outbox aggregate type is required")
	ErrAggregateIDRequired   = errors.New("outbox aggregate id is required")
	ErrEventTypeRequired     = errors.New("outbox event type is required")
)

type Store struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewStore(database *mongo.Database, clock func() time.Time) *Store {
	return &Store{database: database, clock: clock}
}

func (store *Store) collection() *mongo.Collection {
	return store.database.Collection("outbox")
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "publishedAt", Value: 1}, {Key: "occurredAt", Value: 1}},
		Options: options.Index().SetName("outbox_unpublished_fifo"),
	})
	return err
}

// Append validates and stores the record. Pass a transaction session
// context (platform/mongo WithTransaction) so the record commits atomically
// with the domain change it describes.
func (store *Store) Append(ctx context.Context, record Record) error {
	if err := validate(record); err != nil {
		return err
	}
	_, err := store.collection().InsertOne(ctx, document{
		ID:            record.ID,
		AggregateType: record.AggregateType,
		AggregateID:   record.AggregateID,
		EventType:     record.EventType,
		Payload:       record.Payload,
		OccurredAt:    record.OccurredAt.UTC(),
		Attempts:      0,
	})
	return err
}

// Pending returns unpublished records in occurrence order for the relay.
func (store *Store) Pending(ctx context.Context, limit int) ([]Record, error) {
	if limit < 1 {
		limit = 100
	}
	cursor, err := store.collection().Find(ctx,
		bson.M{"publishedAt": nil},
		options.Find().SetSort(bson.D{{Key: "occurredAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []Record
	for cursor.Next(ctx) {
		var doc document
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		records = append(records, Record{
			ID:            doc.ID,
			AggregateType: doc.AggregateType,
			AggregateID:   doc.AggregateID,
			EventType:     doc.EventType,
			Payload:       doc.Payload,
			OccurredAt:    doc.OccurredAt,
		})
	}
	return records, cursor.Err()
}

func (store *Store) MarkPublished(ctx context.Context, id string) error {
	now := store.clock().UTC()
	result, err := store.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"publishedAt": now}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("outbox record %q not found", id)
	}
	return nil
}

// MarkAttemptFailed increments the delivery-attempt counter for relay
// backoff and dead-letter decisions.
func (store *Store) MarkAttemptFailed(ctx context.Context, id string) error {
	_, err := store.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"attempts": 1}})
	return err
}

// FindByEventType returns records of one event type oldest-first,
// regardless of publication state. Consumers that process a specific
// event type (e.g. the safety case builder) dedupe via the inbox store
// and must not interfere with the relay's publication markers.
func (store *Store) FindByEventType(ctx context.Context, eventType string, limit int) ([]Record, error) {
	if limit < 1 {
		limit = 100
	}
	cursor, err := store.collection().Find(ctx,
		bson.M{"eventType": eventType},
		options.Find().SetSort(bson.D{{Key: "occurredAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []Record
	for cursor.Next(ctx) {
		var doc document
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		records = append(records, Record{
			ID:            doc.ID,
			AggregateType: doc.AggregateType,
			AggregateID:   doc.AggregateID,
			EventType:     doc.EventType,
			Payload:       doc.Payload,
			OccurredAt:    doc.OccurredAt,
		})
	}
	return records, cursor.Err()
}

func validate(record Record) error {
	if record.ID == "" {
		return ErrIDRequired
	}
	if record.AggregateType == "" {
		return ErrAggregateTypeRequired
	}
	if record.AggregateID == "" {
		return ErrAggregateIDRequired
	}
	if record.EventType == "" {
		return ErrEventTypeRequired
	}
	return nil
}
