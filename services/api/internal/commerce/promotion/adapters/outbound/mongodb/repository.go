// Package mongodb stores discount codes.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/domain"
)

type Repository struct{ promotions *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{promotions: database.Collection("promotions")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.promotions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// One code, one promotion. Two rows for one string would make
			// which discount a member got depend on read order.
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("promotion_code"),
		},
		{
			Keys:    bson.D{{Key: "commands.id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("promotion_command"),
		},
		{
			Keys:    bson.D{{Key: "issuerId", Value: 1}, {Key: "endsAt", Value: -1}},
			Options: options.Index().SetName("promotion_issuer"),
		},
	})
	return err
}

type eventDocument struct {
	Sequence  uint64        `bson:"sequence"`
	CommandID string        `bson:"commandId"`
	Action    domain.Action `bson:"action"`
	At        time.Time     `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type document struct {
	ID       string       `bson:"_id"`
	Code     string       `bson:"code"`
	IssuerID string       `bson:"issuerId"`
	SKUID    string       `bson:"skuId"`
	Shape    domain.Shape `bson:"shape"`
	Amount   int64        `bson:"amount"`
	StartsAt time.Time    `bson:"startsAt"`
	EndsAt   time.Time    `bson:"endsAt"`
	Cap      uint32       `bson:"cap"`
	// RedeemedKeys are digests. The row says how many members used a code and
	// never which.
	RedeemedKeys []string          `bson:"redeemedKeys"`
	Withdrawn    bool              `bson:"withdrawn"`
	Revision     uint64            `bson:"revision"`
	Events       []eventDocument   `bson:"events"`
	Commands     []commandDocument `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, promotion domain.Promotion) error {
	_, err := repository.promotions.InsertOne(ctx, toDocument(promotion))
	if err == nil {
		return nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return application.ErrUnavailable
	}
	if repository.promotions.FindOne(
		ctx, bson.M{"commands.id": promotion.Commands()[0].ID},
	).Err() == nil {
		// A retried issue found its own code. Answering "taken" would send an
		// operator looking for a clash with somebody else.
		return application.ErrCodeTaken
	}
	return application.ErrCodeTaken
}

func (repository *Repository) FindByCode(ctx context.Context, code string) (domain.Promotion, error) {
	var stored document
	if err := repository.promotions.FindOne(ctx, bson.M{"code": code}).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Promotion{}, application.ErrNotFound
		}
		return domain.Promotion{}, application.ErrUnavailable
	}
	return toDomain(stored)
}

// Append writes one transition onto the revision it was decided against.
//
// The revision guard is what makes the cap hold under load: two members
// redeeming the last slot at once both read revision N, and only one write
// lands.
func (repository *Repository) Append(
	ctx context.Context, promotion domain.Promotion, expected uint64, commandID string,
) error {
	events, commands := promotion.Events(), promotion.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalidPromotion
	}
	result, err := repository.promotions.UpdateOne(ctx,
		bson.M{"_id": promotion.ID(), "revision": expected},
		bson.M{
			"$set": bson.M{
				"redeemedKeys": promotion.RedeemedKeys(),
				"withdrawn":    promotion.Withdrawn(),
				"revision":     promotion.Revision(),
			},
			"$push": bson.M{
				"events":   toEvent(events[len(events)-1]),
				"commands": toCommand(commands[len(commands)-1]),
			},
		})
	if err != nil {
		return application.ErrUnavailable
	}
	if result.MatchedCount == 0 {
		return application.ErrConflict
	}
	return nil
}

func (repository *Repository) ListByIssuer(
	ctx context.Context, issuerID string, limit int,
) ([]domain.Promotion, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	cursor, err := repository.promotions.Find(ctx, bson.M{"issuerId": issuerID},
		options.Find().SetSort(bson.D{{Key: "endsAt", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, application.ErrUnavailable
	}
	defer cursor.Close(ctx)

	promotions := make([]domain.Promotion, 0, limit)
	for cursor.Next(ctx) {
		var stored document
		if err := cursor.Decode(&stored); err != nil {
			return nil, application.ErrUnavailable
		}
		promotion, err := toDomain(stored)
		if err != nil {
			continue
		}
		promotions = append(promotions, promotion)
	}
	if cursor.Err() != nil {
		return nil, application.ErrUnavailable
	}
	return promotions, nil
}

func toDocument(promotion domain.Promotion) document {
	stored := document{
		ID: promotion.ID(), Code: promotion.Code(), IssuerID: promotion.IssuerID(),
		SKUID: promotion.SKUID(), Shape: promotion.Shape(), Amount: promotion.Amount(),
		StartsAt: promotion.StartsAt(), EndsAt: promotion.EndsAt(), Cap: promotion.Cap(),
		RedeemedKeys: promotion.RedeemedKeys(), Withdrawn: promotion.Withdrawn(),
		Revision: promotion.Revision(),
	}
	for _, event := range promotion.Events() {
		stored.Events = append(stored.Events, toEvent(event))
	}
	for _, command := range promotion.Commands() {
		stored.Commands = append(stored.Commands, toCommand(command))
	}
	return stored
}

func toDomain(stored document) (domain.Promotion, error) {
	state := domain.State{
		ID: stored.ID, Code: stored.Code, IssuerID: stored.IssuerID, SKUID: stored.SKUID,
		Shape: stored.Shape, Amount: stored.Amount, StartsAt: stored.StartsAt,
		EndsAt: stored.EndsAt, Cap: stored.Cap, RedeemedKeys: stored.RedeemedKeys,
		Withdrawn: stored.Withdrawn, Revision: stored.Revision,
	}
	for _, event := range stored.Events {
		state.Events = append(state.Events, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID,
			Action: event.Action, At: event.At,
		})
	}
	for _, command := range stored.Commands {
		state.Commands = append(state.Commands, domain.AppliedCommand{
			ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
		})
	}
	return domain.Rehydrate(state)
}

func toEvent(event domain.Event) eventDocument {
	return eventDocument{
		Sequence: event.Sequence, CommandID: event.CommandID,
		Action: event.Action, At: event.At,
	}
}

func toCommand(command domain.AppliedCommand) commandDocument {
	return commandDocument{
		ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
	}
}
