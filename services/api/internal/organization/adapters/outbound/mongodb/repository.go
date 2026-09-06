// Package mongodb stores organizations.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/organization/application"
	"github.com/stanleyHayes/obiara/services/api/internal/organization/domain"
)

type Repository struct{ organizations *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{organizations: database.Collection("organizations")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.organizations.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// Command idempotency. A retried registration must find the
			// organization it already made rather than making a second.
			Keys:    bson.D{{Key: "commands.id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("organization_command"),
		},
		{
			// Two bodies with one name make an audit trail ambiguous about
			// which of them a code was issued for.
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("organization_name"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetName("organization_roster"),
		},
	})
	return err
}

type eventDocument struct {
	Sequence   uint64        `bson:"sequence"`
	CommandID  string        `bson:"commandId"`
	ActorKey   string        `bson:"actorKey"`
	ReasonCode string        `bson:"reasonCode"`
	Action     domain.Action `bson:"action"`
	At         time.Time     `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type document struct {
	ID           string            `bson:"_id"`
	Name         string            `bson:"name"`
	BillingEmail string            `bson:"billingEmail"`
	Status       domain.Status     `bson:"status"`
	Revision     uint64            `bson:"revision"`
	RegisteredAt time.Time         `bson:"registeredAt"`
	Events       []eventDocument   `bson:"events"`
	Commands     []commandDocument `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, organization domain.Organization) error {
	_, err := repository.organizations.InsertOne(ctx, toDocument(organization))
	if err == nil {
		return nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return application.ErrUnavailable
	}
	// Two unique indexes can raise this: a retried command, or a name
	// somebody else already has. They need different answers, so which one it
	// was is established rather than guessed.
	if repository.organizations.FindOne(
		ctx, bson.M{"commands.id": organization.Commands()[0].ID},
	).Err() == nil {
		return application.ErrCommandApplied
	}
	return application.ErrNameTaken
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Organization, error) {
	return repository.find(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindByCommand(ctx context.Context, id string) (domain.Organization, error) {
	return repository.find(ctx, bson.M{"commands.id": id})
}

func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Organization, error) {
	var stored document
	if err := repository.organizations.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Organization{}, application.ErrNotFound
		}
		return domain.Organization{}, application.ErrUnavailable
	}
	return toDomain(stored)
}

// Append writes one transition, and only onto the revision it was decided
// against. Two concurrent changes would otherwise both succeed and the second
// would erase the first's audit entry.
func (repository *Repository) Append(
	ctx context.Context, organization domain.Organization, expected uint64, commandID string,
) error {
	events, commands := organization.Events(), organization.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalidOrganization
	}
	result, err := repository.organizations.UpdateOne(ctx,
		bson.M{"_id": organization.ID(), "revision": expected},
		bson.M{
			"$set": bson.M{
				"status":   organization.Status(),
				"name":     organization.Name(),
				"revision": organization.Revision(),
			},
			"$push": bson.M{
				"events":   toEvent(events[len(events)-1]),
				"commands": toCommand(commands[len(commands)-1]),
			},
		})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return application.ErrNameTaken
		}
		return application.ErrUnavailable
	}
	if result.MatchedCount == 0 {
		if repository.organizations.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
			return application.ErrCommandApplied
		}
		return application.ErrConflict
	}
	return nil
}

// List reads the roster the operator surface shows, active first.
func (repository *Repository) List(ctx context.Context, limit int) ([]domain.Organization, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	cursor, err := repository.organizations.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "status", Value: 1}, {Key: "name", Value: 1}}).
			SetLimit(int64(limit)))
	if err != nil {
		return nil, application.ErrUnavailable
	}
	defer cursor.Close(ctx)

	organizations := make([]domain.Organization, 0, limit)
	for cursor.Next(ctx) {
		var stored document
		if err := cursor.Decode(&stored); err != nil {
			return nil, application.ErrUnavailable
		}
		organization, err := toDomain(stored)
		if err != nil {
			// One unreadable row does not make the roster unreadable. Failing
			// the whole list on it would take the surface down for every
			// other organization.
			continue
		}
		organizations = append(organizations, organization)
	}
	if cursor.Err() != nil {
		return nil, application.ErrUnavailable
	}
	return organizations, nil
}

func toDocument(organization domain.Organization) document {
	stored := document{
		ID: organization.ID(), Name: organization.Name(),
		BillingEmail: organization.BillingEmail(), Status: organization.Status(),
		Revision: organization.Revision(), RegisteredAt: organization.RegisteredAt(),
	}
	for _, event := range organization.Events() {
		stored.Events = append(stored.Events, toEvent(event))
	}
	for _, command := range organization.Commands() {
		stored.Commands = append(stored.Commands, toCommand(command))
	}
	return stored
}

func toDomain(stored document) (domain.Organization, error) {
	state := domain.State{
		ID: stored.ID, Name: stored.Name, BillingEmail: stored.BillingEmail,
		Status: stored.Status, Revision: stored.Revision, RegisteredAt: stored.RegisteredAt,
	}
	for _, event := range stored.Events {
		state.Events = append(state.Events, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
			ReasonCode: event.ReasonCode, Action: event.Action, At: event.At,
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
		ReasonCode: event.ReasonCode, Action: event.Action, At: event.At,
	}
}

func toCommand(command domain.AppliedCommand) commandDocument {
	return commandDocument{
		ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
	}
}
