// Package mongodb stores sponsorship funds.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/domain"
)

type Repository struct{ funds *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{funds: database.Collection("sponsorship_funds")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.funds.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// One fund per organization. Two would split a balance and make
			// "how much is left" a question with two answers.
			Keys:    bson.D{{Key: "organizationId", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("sponsorship_fund_organization"),
		},
		{
			Keys:    bson.D{{Key: "commands.id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("sponsorship_fund_command"),
		},
	})
	return err
}

type eventDocument struct {
	Sequence      uint64        `bson:"sequence"`
	CommandID     string        `bson:"commandId"`
	ActorKey      string        `bson:"actorKey"`
	ReasonCode    string        `bson:"reasonCode"`
	Action        domain.Action `bson:"action"`
	AmountPesewas int64         `bson:"amountPesewas"`
	At            time.Time     `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type document struct {
	ID               string            `bson:"_id"`
	OrganizationID   string            `bson:"organizationId"`
	DepositedPesewas int64             `bson:"depositedPesewas"`
	DrawnPesewas     int64             `bson:"drawnPesewas"`
	DrawnSeats       map[string]int64  `bson:"drawnSeats"`
	Closed           bool              `bson:"closed"`
	Revision         uint64            `bson:"revision"`
	OpenedAt         time.Time         `bson:"openedAt"`
	Events           []eventDocument   `bson:"events"`
	Commands         []commandDocument `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, fund domain.Fund) error {
	_, err := repository.funds.InsertOne(ctx, toDocument(fund))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// Somebody opened it between the read and this write. The caller
			// re-reads rather than making a second fund.
			return application.ErrConflict
		}
		return application.ErrUnavailable
	}
	return nil
}

func (repository *Repository) FindByOrganization(
	ctx context.Context, organizationID string,
) (domain.Fund, error) {
	var stored document
	if err := repository.funds.FindOne(
		ctx, bson.M{"organizationId": organizationID},
	).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Fund{}, application.ErrNotFound
		}
		return domain.Fund{}, application.ErrUnavailable
	}
	return toDomain(stored)
}

// Append writes one movement onto the revision it was decided against.
//
// The revision guard is what keeps a balance honest under load: two members
// drawing the last seat both read revision N, and only one write lands.
func (repository *Repository) Append(
	ctx context.Context, fund domain.Fund, expected uint64, commandID string,
) error {
	events, commands := fund.Events(), fund.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalidFund
	}
	result, err := repository.funds.UpdateOne(ctx,
		bson.M{"_id": fund.ID(), "revision": expected},
		bson.M{
			"$set": bson.M{
				"depositedPesewas": fund.DepositedPesewas(),
				"drawnPesewas":     fund.DrawnPesewas(),
				"drawnSeats":       fund.DrawnSeats(),
				"closed":           fund.Closed(),
				"revision":         fund.Revision(),
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
		if repository.funds.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
			return nil
		}
		return application.ErrConflict
	}
	return nil
}

func (repository *Repository) List(ctx context.Context, limit int) ([]domain.Fund, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	cursor, err := repository.funds.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "organizationId", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, application.ErrUnavailable
	}
	defer cursor.Close(ctx)

	funds := make([]domain.Fund, 0, limit)
	for cursor.Next(ctx) {
		var stored document
		if err := cursor.Decode(&stored); err != nil {
			return nil, application.ErrUnavailable
		}
		fund, err := toDomain(stored)
		if err != nil {
			continue
		}
		funds = append(funds, fund)
	}
	return funds, cursor.Err()
}

func toDocument(fund domain.Fund) document {
	stored := document{
		ID: fund.ID(), OrganizationID: fund.OrganizationID(),
		DepositedPesewas: fund.DepositedPesewas(), DrawnPesewas: fund.DrawnPesewas(),
		DrawnSeats: fund.DrawnSeats(), Closed: fund.Closed(),
		Revision: fund.Revision(), OpenedAt: fund.OpenedAt(),
	}
	for _, event := range fund.Events() {
		stored.Events = append(stored.Events, toEvent(event))
	}
	for _, command := range fund.Commands() {
		stored.Commands = append(stored.Commands, toCommand(command))
	}
	return stored
}

func toDomain(stored document) (domain.Fund, error) {
	state := domain.State{
		ID: stored.ID, OrganizationID: stored.OrganizationID,
		DepositedPesewas: stored.DepositedPesewas, DrawnPesewas: stored.DrawnPesewas,
		DrawnSeats: stored.DrawnSeats, Closed: stored.Closed,
		Revision: stored.Revision, OpenedAt: stored.OpenedAt,
	}
	for _, event := range stored.Events {
		state.Events = append(state.Events, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
			ReasonCode: event.ReasonCode, Action: event.Action,
			AmountPesewas: event.AmountPesewas, At: event.At,
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
		Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
		ReasonCode: event.ReasonCode, Action: event.Action,
		AmountPesewas: event.AmountPesewas, At: event.At,
	}
}

func toCommand(command domain.AppliedCommand) commandDocument {
	return commandDocument{
		ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
	}
}
