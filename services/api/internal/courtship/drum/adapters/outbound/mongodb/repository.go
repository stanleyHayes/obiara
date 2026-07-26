package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/domain"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("courtship_drum_stages")}
}
func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetName("courtship_drum_command_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "members", Value: 1}}, Options: options.Index().SetName("courtship_drum_private_members")},
	})
	return err
}

type beatDocument struct {
	Sequence   uint64        `bson:"sequence"`
	CommandID  string        `bson:"commandId"`
	ActorKey   string        `bson:"actorKey"`
	Medium     domain.Medium `bson:"medium"`
	ContentKey string        `bson:"contentKey"`
	ReasonCode string        `bson:"reasonCode"`
	At         time.Time     `bson:"at"`
}
type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type document struct {
	ID           string            `bson:"_id"`
	Members      []string          `bson:"members"`
	Beats        []beatDocument    `bson:"beats"`
	Commands     []commandDocument `bson:"commands"`
	Revision     uint64            `bson:"revision"`
	NextActorKey string            `bson:"nextActorKey"`
}

func (repository *Repository) Create(ctx context.Context, stage domain.Stage) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(stage))
	return repository.translateDuplicate(ctx, stage.Commands()[0].ID, err)
}
func (repository *Repository) Find(ctx context.Context, id string) (domain.Stage, error) {
	return repository.find(ctx, bson.M{"_id": id})
}
func (repository *Repository) FindByCommand(ctx context.Context, commandID string) (domain.Stage, error) {
	return repository.find(ctx, bson.M{"commands.id": commandID})
}
func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Stage, error) {
	var stored document
	if err := repository.collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Stage{}, application.ErrNotFound
		}
		return domain.Stage{}, err
	}
	return toDomain(stored)
}
func (repository *Repository) Append(ctx context.Context, stage domain.Stage, expectedRevision uint64, commandID string) error {
	beats, commands := stage.Beats(), stage.Commands()
	if len(beats) != int(expectedRevision+1) || len(commands) != int(expectedRevision+1) {
		return domain.ErrInvalid
	}
	result, err := repository.collection.UpdateOne(ctx, bson.M{
		"_id": stage.ID(), "revision": expectedRevision,
	}, bson.M{
		"$set": bson.M{"revision": stage.Revision(), "nextActorKey": stage.NextActorKey()},
		"$push": bson.M{
			"beats":    fromBeat(beats[len(beats)-1]),
			"commands": fromCommand(commands[len(commands)-1]),
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
func toDocument(stage domain.Stage) document {
	stored := document{
		ID: stage.ID(), Members: stage.Members(), Revision: stage.Revision(),
		NextActorKey: stage.NextActorKey(),
	}
	for _, beat := range stage.Beats() {
		stored.Beats = append(stored.Beats, fromBeat(beat))
	}
	for _, command := range stage.Commands() {
		stored.Commands = append(stored.Commands, fromCommand(command))
	}
	return stored
}
func toDomain(stored document) (domain.Stage, error) {
	state := domain.State{ID: stored.ID, Members: stored.Members}
	for _, beat := range stored.Beats {
		state.Beats = append(state.Beats, domain.Beat{
			Sequence: beat.Sequence, CommandID: beat.CommandID, ActorKey: beat.ActorKey,
			Medium: beat.Medium, ContentKey: beat.ContentKey, ReasonCode: beat.ReasonCode, At: beat.At,
		})
	}
	for _, command := range stored.Commands {
		state.Commands = append(state.Commands, domain.AppliedCommand{
			ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
		})
	}
	stage, err := domain.Rehydrate(state)
	if err != nil {
		return domain.Stage{}, err
	}
	if stored.Revision != stage.Revision() || stored.NextActorKey != stage.NextActorKey() {
		return domain.Stage{}, domain.ErrInvalid
	}
	return stage, nil
}
func fromBeat(beat domain.Beat) beatDocument {
	return beatDocument{
		Sequence: beat.Sequence, CommandID: beat.CommandID, ActorKey: beat.ActorKey,
		Medium: beat.Medium, ContentKey: beat.ContentKey, ReasonCode: beat.ReasonCode, At: beat.At,
	}
}
func fromCommand(command domain.AppliedCommand) commandDocument {
	return commandDocument{ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision}
}
