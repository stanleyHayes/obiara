package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/domain"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("ampe_rounds")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "createCommandId", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("ampe_create_command_unique"),
		},
		{
			Keys:    bson.D{{Key: "transcript.commands.id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true).SetName("ampe_command_unique"),
		},
		{
			Keys:    bson.D{{Key: "spec.roomKey", Value: 1}},
			Options: options.Index().SetName("ampe_room"),
		},
	})
	return err
}

type document struct {
	ID              string            `bson:"_id"`
	Spec            domain.Spec       `bson:"spec"`
	Transcript      domain.Transcript `bson:"transcript"`
	Sequence        uint64            `bson:"sequence"`
	CreateCommandID string            `bson:"createCommandId"`
}

func (repository *Repository) Create(ctx context.Context, round domain.Round, commandID string) error {
	_, err := repository.collection.InsertOne(ctx, document{
		ID: round.Specification().ID, Spec: round.Specification(),
		Transcript: round.PrivateTranscript(), Sequence: round.Sequence(),
		CreateCommandID: commandID,
	})
	return repository.duplicate(ctx, commandID, err)
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Round, error) {
	return repository.find(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindByCommand(ctx context.Context, commandID string) (domain.Round, error) {
	return repository.find(ctx, bson.M{"$or": bson.A{
		bson.M{"createCommandId": commandID},
		bson.M{"transcript.commands.id": commandID},
	}})
}

func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Round, error) {
	var stored document
	if err := repository.collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Round{}, application.ErrNotFound
		}
		return domain.Round{}, err
	}
	return domain.Replay(stored.Spec, stored.Transcript)
}

func (repository *Repository) Append(
	ctx context.Context,
	round domain.Round,
	expected uint64,
	commandID string,
) error {
	result, err := repository.collection.UpdateOne(
		ctx,
		bson.M{"_id": round.Specification().ID, "sequence": expected},
		bson.M{"$set": bson.M{
			"transcript": round.PrivateTranscript(),
			"sequence":   round.Sequence(),
		}},
	)
	if err != nil {
		return repository.duplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		if repository.collection.FindOne(ctx, bson.M{
			"transcript.commands.id": commandID,
		}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}

func (repository *Repository) duplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if repository.collection.FindOne(ctx, bson.M{"$or": bson.A{
		bson.M{"createCommandId": commandID},
		bson.M{"transcript.commands.id": commandID},
	}}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
