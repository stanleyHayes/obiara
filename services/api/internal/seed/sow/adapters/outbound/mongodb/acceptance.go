package mongodb

import (
	"context"
	"errors"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Acceptance struct{ database *mongo.Database }

func NewAcceptance(database *mongo.Database) *Acceptance { return &Acceptance{database} }
func (a *Acceptance) sows() *mongo.Collection            { return a.database.Collection("seed_sows") }
func (a *Acceptance) events() *mongo.Collection          { return a.database.Collection("seed_sow_events") }
func (a *Acceptance) heads() *mongo.Collection {
	return a.database.Collection("seed_allowance_ledgers")
}
func (a *Acceptance) allowanceEntries() *mongo.Collection {
	return a.database.Collection("seed_allowance_entries")
}

type mediaDocument struct {
	Key          string `bson:"key"`
	ScreeningKey string `bson:"screeningKey"`
}
type sowDocument struct {
	ID             string          `bson:"_id"`
	ActorKey       string          `bson:"actorKey"`
	Body           string          `bson:"body"`
	Media          []mediaDocument `bson:"media"`
	CommandID      string          `bson:"commandId"`
	Fingerprint    string          `bson:"fingerprint"`
	AllowanceUnits int64           `bson:"allowanceUnits"`
	Status         string          `bson:"status"`
	ScreeningRef   string          `bson:"screeningRef"`
	AcceptedAt     time.Time       `bson:"acceptedAt"`
	DecidedAt      *time.Time      `bson:"decidedAt,omitempty"`
}
type allowanceHead struct {
	ID        string    `bson:"_id"`
	Balance   int64     `bson:"balance"`
	WeekStart time.Time `bson:"weekStart"`
	Version   int64     `bson:"version"`
}
type eventDocument struct {
	SowID      string    `bson:"sowId"`
	Kind       string    `bson:"kind"`
	CommandID  string    `bson:"commandId"`
	ActorKey   string    `bson:"actorKey"`
	OccurredAt time.Time `bson:"occurredAt"`
}

func (a *Acceptance) EnsureIndexes(ctx context.Context) error {
	if _, err := a.sows().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("seed_sow_command_unique")}); err != nil {
		return err
	}
	_, err := a.events().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "sowId", Value: 1}, {Key: "kind", Value: 1}}, Options: options.Index().SetUnique(true).SetName("seed_sow_event_unique")})
	return err
}

func (a *Acceptance) Accept(ctx context.Context, sow domain.Sow) (domain.Sow, bool, error) {
	err := apimongo.WithTransaction(ctx, a.database.Client(), func(tx context.Context) error {
		if _, insertErr := a.sows().InsertOne(tx, toDocument(sow)); insertErr != nil {
			return insertErr
		}
		var head allowanceHead
		if findErr := a.heads().FindOne(tx, bson.M{"_id": sow.ActorKey}).Decode(&head); findErr != nil {
			if errors.Is(findErr, mongo.ErrNoDocuments) {
				return application.ErrInsufficientAllowance
			}
			return findErr
		}
		if head.Balance < sow.AllowanceUnits {
			return application.ErrInsufficientAllowance
		}
		result, updateErr := a.heads().UpdateOne(tx, bson.M{"_id": head.ID, "version": head.Version, "balance": bson.M{"$gte": sow.AllowanceUnits}},
			bson.M{"$inc": bson.M{"balance": -sow.AllowanceUnits, "version": 1}})
		if updateErr != nil {
			return updateErr
		}
		if result.MatchedCount == 0 {
			return application.ErrInsufficientAllowance
		}
		if _, entryErr := a.allowanceEntries().InsertOne(tx, bson.M{
			"ledgerId": head.ID, "sequence": head.Version + 1, "kind": "spend", "commandId": sow.CommandID,
			"fingerprint": sow.Fingerprint, "units": sow.AllowanceUnits, "balanceAfter": head.Balance - sow.AllowanceUnits,
			"weekStart": head.WeekStart, "occurredAt": sow.AcceptedAt, "actorKey": sow.ActorKey,
		}); entryErr != nil {
			return entryErr
		}
		_, eventErr := a.events().InsertOne(tx, eventDocument{sow.ID, "accepted", sow.CommandID, sow.ActorKey, sow.AcceptedAt})
		return eventErr
	})
	if err == nil {
		return sow, false, nil
	}
	if !apimongo.IsDuplicateKey(err) {
		return domain.Sow{}, false, err
	}
	existing, findErr := a.findByCommand(ctx, sow.CommandID)
	if findErr != nil {
		return domain.Sow{}, false, err
	}
	if existing.Fingerprint != sow.Fingerprint {
		return domain.Sow{}, false, domain.ErrCommandMismatch
	}
	return existing, true, nil
}

func (a *Acceptance) findByCommand(ctx context.Context, commandID string) (domain.Sow, error) {
	var document sowDocument
	if err := a.sows().FindOne(ctx, bson.M{"commandId": commandID}).Decode(&document); err != nil {
		return domain.Sow{}, err
	}
	return fromDocument(document)
}
func toDocument(s domain.Sow) sowDocument {
	media := make([]mediaDocument, 0, len(s.Media))
	for _, m := range s.Media {
		media = append(media, mediaDocument{m.Key, m.ScreeningKey})
	}
	return sowDocument{
		ID: s.ID, ActorKey: s.ActorKey, Body: s.Body, Media: media,
		CommandID: s.CommandID, Fingerprint: s.Fingerprint,
		AllowanceUnits: s.AllowanceUnits, Status: string(s.Status),
		ScreeningRef: s.ScreeningRef, AcceptedAt: s.AcceptedAt, DecidedAt: s.DecidedAt,
	}
}
func fromDocument(d sowDocument) (domain.Sow, error) {
	media := make([]domain.Media, 0, len(d.Media))
	for _, m := range d.Media {
		media = append(media, domain.Media{Key: m.Key, ScreeningKey: m.ScreeningKey})
	}
	return domain.Reconstitute(d.ID, d.ActorKey, d.Body, media, d.CommandID, d.Fingerprint,
		d.AllowanceUnits, domain.Status(d.Status), d.ScreeningRef, d.AcceptedAt, d.DecidedAt), nil
}
