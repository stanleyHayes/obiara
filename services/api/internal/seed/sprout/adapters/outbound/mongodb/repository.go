package mongodb

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database} }
func (r *Repository) intents() *mongo.Collection         { return r.database.Collection("seed_sprout_intents") }
func (r *Repository) doorways() *mongo.Collection {
	return r.database.Collection("seed_sprout_doorways")
}
func (r *Repository) events() *mongo.Collection { return r.database.Collection("seed_sprout_events") }

type intentDocument struct {
	ID          string    `bson:"_id"`
	ActorKey    string    `bson:"actorKey"`
	TargetKey   string    `bson:"targetKey"`
	SeedKey     string    `bson:"seedKey"`
	PairKey     string    `bson:"pairKey"`
	CommandID   string    `bson:"commandId"`
	Fingerprint string    `bson:"fingerprint"`
	At          time.Time `bson:"at"`
}
type exchangeDocument struct {
	Number      int       `bson:"number"`
	ActorKey    string    `bson:"actorKey"`
	MessageKey  string    `bson:"messageKey"`
	CommandID   string    `bson:"commandId"`
	Fingerprint string    `bson:"fingerprint"`
	At          time.Time `bson:"at"`
}
type doorwayDocument struct {
	ID           string             `bson:"_id"`
	PairKey      string             `bson:"pairKey"`
	Participants [2]string          `bson:"participants"`
	NextActorKey string             `bson:"nextActorKey,omitempty"`
	Exchanges    []exchangeDocument `bson:"exchanges"`
	Revision     uint64             `bson:"revision"`
	OpenedAt     time.Time          `bson:"openedAt"`
	SealedAt     time.Time          `bson:"sealedAt,omitempty"`
}
type eventDocument struct {
	DoorwayID   string    `bson:"doorwayId"`
	Number      int       `bson:"number"`
	CommandID   string    `bson:"commandId"`
	Fingerprint string    `bson:"fingerprint"`
	ActorKey    string    `bson:"actorKey"`
	MessageKey  string    `bson:"messageKey"`
	At          time.Time `bson:"at"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("sprout_intent_command_unique")},
		{Keys: bson.D{{Key: "actorKey", Value: 1}, {Key: "targetKey", Value: 1}, {Key: "seedKey", Value: 1}}, Options: options.Index().SetUnique(true).SetName("sprout_directed_unique")},
	}
	if _, err := r.intents().Indexes().CreateMany(ctx, models); err != nil {
		return err
	}
	if _, err := r.doorways().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "pairKey", Value: 1}}, Options: options.Index().SetUnique(true).SetName("sprout_doorway_pair_unique")}); err != nil {
		return err
	}
	_, err := r.events().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "doorwayId", Value: 1}, {Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("sprout_exchange_command_unique")})
	return err
}

func (r *Repository) RecordIntent(ctx context.Context, intent domain.Intent) (*domain.Doorway, bool, error) {
	doc := intentDocument{intent.ID, intent.ActorKey, intent.TargetKey, intent.SeedKey, pairKey(intent.ActorKey, intent.TargetKey, intent.SeedKey), intent.CommandID, intent.Fingerprint, intent.At}
	replayed := false
	if _, err := r.intents().InsertOne(ctx, doc); err != nil {
		if !apimongo.IsDuplicateKey(err) {
			return nil, false, err
		}
		var existing intentDocument
		findErr := r.intents().FindOne(ctx, bson.M{"commandId": intent.CommandID}).Decode(&existing)
		if findErr == nil {
			if existing.Fingerprint != intent.Fingerprint {
				return nil, false, domain.ErrCommandMismatch
			}
			replayed = true
		} else {
			if findErr = r.intents().FindOne(ctx, bson.M{"actorKey": intent.ActorKey, "targetKey": intent.TargetKey, "seedKey": intent.SeedKey}).Decode(&existing); findErr != nil {
				return nil, false, err
			}
			replayed = true
		}
	}
	var reciprocal intentDocument
	if err := r.intents().FindOne(ctx, bson.M{"actorKey": intent.TargetKey, "targetKey": intent.ActorKey, "seedKey": intent.SeedKey}).Decode(&reciprocal); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, replayed, nil
		}
		return nil, false, err
	}
	first, second := orderedIntents(reciprocal, doc)
	doorway, err := domain.Open("doorway:"+first.ID, first.ActorKey, second.ActorKey, maxTime(first.At, second.At))
	if err != nil {
		return nil, false, err
	}
	document := toDoorway(doorway, doc.PairKey)
	if _, err = r.doorways().InsertOne(ctx, document); err == nil {
		return &doorway, replayed, nil
	} else if !apimongo.IsDuplicateKey(err) {
		return nil, false, err
	}
	existing, err := r.findByPair(ctx, doc.PairKey)
	if err != nil {
		return nil, false, err
	}
	return &existing, true, nil
}

func (r *Repository) FindDoorway(ctx context.Context, id string) (domain.Doorway, error) {
	var document doorwayDocument
	if err := r.doorways().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Doorway{}, application.ErrNotFound
		}
		return domain.Doorway{}, err
	}
	return fromDoorway(document)
}
func (r *Repository) findByPair(ctx context.Context, pair string) (domain.Doorway, error) {
	var d doorwayDocument
	if err := r.doorways().FindOne(ctx, bson.M{"pairKey": pair}).Decode(&d); err != nil {
		return domain.Doorway{}, err
	}
	return fromDoorway(d)
}

func (r *Repository) AppendExchange(ctx context.Context, next domain.Doorway, expected uint64) (domain.Doorway, bool, error) {
	exchanges := next.Exchanges()
	latest := exchanges[len(exchanges)-1]
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, updateErr := r.doorways().UpdateOne(tx, bson.M{"_id": next.ID(), "revision": expected},
			bson.M{"$set": bson.M{"nextActorKey": next.NextActorKey(), "revision": next.Revision(), "sealedAt": next.SealedAt()}, "$push": bson.M{"exchanges": toExchange(latest)}})
		if updateErr != nil {
			return updateErr
		}
		if result.MatchedCount == 0 {
			return application.ErrConcurrentChange
		}
		_, eventErr := r.events().InsertOne(tx, eventDocument{next.ID(), latest.Number, latest.CommandID, latest.Fingerprint, latest.ActorKey, latest.MessageKey, latest.At})
		return eventErr
	})
	if err == nil {
		return next, false, nil
	}
	if !errors.Is(err, application.ErrConcurrentChange) && !apimongo.IsDuplicateKey(err) {
		return domain.Doorway{}, false, err
	}
	var event eventDocument
	if findErr := r.events().FindOne(ctx, bson.M{"doorwayId": next.ID(), "commandId": latest.CommandID}).Decode(&event); findErr == nil {
		if event.Fingerprint != latest.Fingerprint {
			return domain.Doorway{}, false, domain.ErrCommandMismatch
		}
		stored, loadErr := r.FindDoorway(ctx, next.ID())
		return stored, true, loadErr
	}
	return domain.Doorway{}, false, application.ErrConcurrentChange
}

func pairKey(a, b, seed string) string {
	parts := []string{a, b}
	sort.Strings(parts)
	return strings.Join(parts, "|") + "|" + seed
}
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func orderedIntents(a, b intentDocument) (intentDocument, intentDocument) {
	if a.At.Before(b.At) || a.At.Equal(b.At) && a.ActorKey < b.ActorKey {
		return a, b
	}
	return b, a
}
func toExchange(e domain.Exchange) exchangeDocument {
	return exchangeDocument{e.Number, e.ActorKey, e.MessageKey, e.CommandID, e.Fingerprint, e.At}
}
func fromExchange(e exchangeDocument) domain.Exchange {
	return domain.Exchange{Number: e.Number, ActorKey: e.ActorKey, MessageKey: e.MessageKey, CommandID: e.CommandID, Fingerprint: e.Fingerprint, At: e.At}
}
func toDoorway(d domain.Doorway, pair string) doorwayDocument {
	return doorwayDocument{d.ID(), pair, d.Participants(), d.NextActorKey(), []exchangeDocument{}, d.Revision(), d.OpenedAt(), d.SealedAt()}
}
func fromDoorway(d doorwayDocument) (domain.Doorway, error) {
	exchanges := make([]domain.Exchange, 0, len(d.Exchanges))
	for _, e := range d.Exchanges {
		exchanges = append(exchanges, fromExchange(e))
	}
	return domain.Rehydrate(d.ID, d.Participants, d.NextActorKey, exchanges, d.Revision, d.OpenedAt, d.SealedAt)
}
