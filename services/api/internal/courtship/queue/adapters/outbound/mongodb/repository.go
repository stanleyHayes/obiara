package mongodb

import (
	"context"
	"errors"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database} }
func (r *Repository) heads() *mongo.Collection           { return r.database.Collection("courtship_queue_heads") }
func (r *Repository) events() *mongo.Collection {
	return r.database.Collection("courtship_queue_events")
}

type headDocument struct {
	ID       string `bson:"_id"`
	Sequence uint64 `bson:"sequence"`
	Revision uint64 `bson:"revision"`
}
type eventDocument struct {
	RoomKey     string    `bson:"roomKey"`
	Sequence    uint64    `bson:"sequence"`
	CommandID   string    `bson:"commandId"`
	DeviceKey   string    `bson:"deviceKey"`
	ActorKey    string    `bson:"actorKey"`
	PayloadKey  string    `bson:"payloadKey"`
	Fingerprint string    `bson:"fingerprint"`
	AcceptedAt  time.Time `bson:"acceptedAt"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "roomKey", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetUnique(true).SetName("courtship_queue_room_sequence_unique")},
		{Keys: bson.D{{Key: "roomKey", Value: 1}, {Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("courtship_queue_command_unique")},
	}
	_, err := r.events().Indexes().CreateMany(ctx, models)
	return err
}
func (r *Repository) State(ctx context.Context, roomKey string) (domain.State, error) {
	var document headDocument
	err := r.heads().FindOneAndUpdate(ctx, bson.M{"_id": roomKey}, bson.M{"$setOnInsert": bson.M{"sequence": uint64(0), "revision": uint64(0)}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&document)
	if err != nil {
		return domain.State{}, err
	}
	return domain.Rehydrate(document.ID, document.Sequence, document.Revision)
}
func (r *Repository) Append(ctx context.Context, next domain.State, event domain.Event, expected uint64) (domain.Event, bool, error) {
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, updateErr := r.heads().UpdateOne(tx, bson.M{"_id": next.RoomKey, "revision": expected, "sequence": event.Sequence - 1},
			bson.M{"$set": bson.M{"sequence": next.Sequence, "revision": next.Revision}})
		if updateErr != nil {
			return updateErr
		}
		if result.MatchedCount == 0 {
			return application.ErrConcurrentChange
		}
		_, insertErr := r.events().InsertOne(tx, toDocument(event))
		return insertErr
	})
	if err == nil {
		return event, false, nil
	}
	if !errors.Is(err, application.ErrConcurrentChange) && !apimongo.IsDuplicateKey(err) {
		return domain.Event{}, false, err
	}
	var existing eventDocument
	findErr := r.events().FindOne(ctx, bson.M{"roomKey": event.RoomKey, "commandId": event.CommandID}).Decode(&existing)
	if findErr == nil {
		if existing.Fingerprint != event.Fingerprint {
			return domain.Event{}, false, domain.ErrCommandMismatch
		}
		return fromDocument(existing), true, nil
	}
	if !errors.Is(findErr, mongo.ErrNoDocuments) {
		return domain.Event{}, false, findErr
	}
	return domain.Event{}, false, domain.ErrStaleDevice
}
func (r *Repository) EventByCommand(ctx context.Context, roomKey, commandID string) (domain.Event, error) {
	var document eventDocument
	if err := r.events().FindOne(ctx, bson.M{"roomKey": roomKey, "commandId": commandID}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Event{}, application.ErrNotFound
		}
		return domain.Event{}, err
	}
	return fromDocument(document), nil
}
func (r *Repository) Events(ctx context.Context, roomKey string, after uint64, limit int) ([]domain.Event, error) {
	if limit < 1 || limit > 200 {
		limit = 100
	}
	cursor, err := r.events().Find(ctx, bson.M{"roomKey": roomKey, "sequence": bson.M{"$gt": after}}, options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []eventDocument
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	events := make([]domain.Event, 0, len(documents))
	for _, document := range documents {
		events = append(events, fromDocument(document))
	}
	return events, nil
}
func toDocument(e domain.Event) eventDocument {
	return eventDocument{e.RoomKey, e.Sequence, e.CommandID, e.DeviceKey, e.ActorKey, e.PayloadKey, e.Fingerprint, e.AcceptedAt}
}
func fromDocument(e eventDocument) domain.Event {
	return domain.Event{RoomKey: e.RoomKey, Sequence: e.Sequence, CommandID: e.CommandID, DeviceKey: e.DeviceKey, ActorKey: e.ActorKey, PayloadKey: e.PayloadKey, Fingerprint: e.Fingerprint, AcceptedAt: e.AcceptedAt}
}
