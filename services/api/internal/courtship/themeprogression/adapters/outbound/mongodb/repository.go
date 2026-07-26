package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/domain"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("courtship_theme_progressions")}
}
func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetName("theme_progression_command_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "members", Value: 1}}, Options: options.Index().SetName("theme_progression_private_members")},
	})
	return err
}

type eventDocument struct {
	Sequence    uint64             `bson:"sequence"`
	CommandID   string             `bson:"commandId"`
	ActorKey    string             `bson:"actorKey"`
	Action      domain.Action      `bson:"action"`
	Theme       domain.ThemeNumber `bson:"theme"`
	ContentKey  string             `bson:"contentKey,omitempty"`
	EvidenceKey string             `bson:"evidenceKey,omitempty"`
	ReasonCode  string             `bson:"reasonCode"`
	At          time.Time          `bson:"at"`
}
type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type document struct {
	ID         string            `bson:"_id"`
	Members    []string          `bson:"members"`
	Events     []eventDocument   `bson:"events"`
	Commands   []commandDocument `bson:"commands"`
	Revision   uint64            `bson:"revision"`
	Projection domain.Projection `bson:"projection"`
}

func (repository *Repository) Create(ctx context.Context, progression domain.Progression) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(progression))
	return repository.translateDuplicate(ctx, progression.Commands()[0].ID, err)
}
func (repository *Repository) Find(ctx context.Context, id string) (domain.Progression, error) {
	return repository.find(ctx, bson.M{"_id": id})
}
func (repository *Repository) FindByCommand(ctx context.Context, commandID string) (domain.Progression, error) {
	return repository.find(ctx, bson.M{"commands.id": commandID})
}
func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Progression, error) {
	var stored document
	if err := repository.collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Progression{}, application.ErrNotFound
		}
		return domain.Progression{}, err
	}
	return toDomain(stored)
}
func (repository *Repository) Append(ctx context.Context, progression domain.Progression, expected uint64, commandID string) error {
	events, commands := progression.Events(), progression.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalid
	}
	result, err := repository.collection.UpdateOne(ctx, bson.M{
		"_id": progression.ID(), "revision": expected,
	}, bson.M{
		"$set": bson.M{"revision": progression.Revision(), "projection": progression.Projection()},
		"$push": bson.M{
			"events": fromEvent(events[len(events)-1]), "commands": fromCommand(commands[len(commands)-1]),
		},
	})
	if err != nil {
		return repository.translateDuplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		if repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
			return application.ErrCommandApplied
		}
		return application.ErrOptimisticConflict
	}
	return nil
}
func (repository *Repository) translateDuplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
		return application.ErrCommandApplied
	}
	return application.ErrOptimisticConflict
}
func toDocument(progression domain.Progression) document {
	stored := document{
		ID: progression.ID(), Members: progression.Members(),
		Revision: progression.Revision(), Projection: progression.Projection(),
	}
	for _, event := range progression.Events() {
		stored.Events = append(stored.Events, fromEvent(event))
	}
	for _, command := range progression.Commands() {
		stored.Commands = append(stored.Commands, fromCommand(command))
	}
	return stored
}
func toDomain(stored document) (domain.Progression, error) {
	state := domain.State{ID: stored.ID, Members: stored.Members}
	for _, event := range stored.Events {
		state.Events = append(state.Events, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
			Action: event.Action, Theme: event.Theme, ContentKey: event.ContentKey,
			EvidenceKey: event.EvidenceKey, ReasonCode: event.ReasonCode, At: event.At,
		})
	}
	for _, command := range stored.Commands {
		state.Commands = append(state.Commands, domain.AppliedCommand{
			ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
		})
	}
	progression, err := domain.Rehydrate(state)
	if err != nil {
		return domain.Progression{}, err
	}
	if stored.Revision != progression.Revision() ||
		!equalProjection(stored.Projection, progression.Projection()) {
		return domain.Progression{}, domain.ErrInvalid
	}
	return progression, nil
}
func equalProjection(left, right domain.Projection) bool {
	if len(left.Themes) != len(right.Themes) {
		return false
	}
	for index := range left.Themes {
		a, b := left.Themes[index], right.Themes[index]
		if a.Number != b.Number || a.PromptRef != b.PromptRef ||
			a.PromptVersion != b.PromptVersion || a.Unlocked != b.Unlocked ||
			a.SubmittedCount != b.SubmittedCount || a.Revealed != b.Revealed ||
			len(a.Submissions) != len(b.Submissions) {
			return false
		}
		for item := range a.Submissions {
			if a.Submissions[item] != b.Submissions[item] {
				return false
			}
		}
	}
	return true
}
func fromEvent(event domain.Event) eventDocument {
	return eventDocument{
		Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
		Action: event.Action, Theme: event.Theme, ContentKey: event.ContentKey,
		EvidenceKey: event.EvidenceKey, ReasonCode: event.ReasonCode, At: event.At,
	}
}
func fromCommand(command domain.AppliedCommand) commandDocument {
	return commandDocument{ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision}
}
