package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/application"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

type Repository struct {
	collection *mongo.Collection
}

type fieldDocument struct {
	Value      string            `bson:"value"`
	Visibility domain.Visibility `bson:"visibility"`
	ConsentRef string            `bson:"consentRef,omitempty"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type profileDocument struct {
	MemberID     string            `bson:"_id"`
	DisplayName  fieldDocument     `bson:"displayName"`
	Introduction fieldDocument     `bson:"introduction"`
	Revision     uint64            `bson:"revision"`
	UpdatedAt    time.Time         `bson:"updatedAt"`
	Commands     []commandDocument `bson:"commands"`
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("profiles")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "commands.id", Value: 1}},
		Options: options.Index().
			SetName("profiles_command_id_unique").
			SetUnique(true).
			SetPartialFilterExpression(bson.M{"commands.id": bson.M{"$exists": true}}),
	})
	return err
}

func (repository *Repository) Find(ctx context.Context, memberID string) (domain.Profile, error) {
	var document profileDocument
	if err := repository.collection.FindOne(ctx, bson.M{"_id": memberID}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Profile{}, application.ErrNotFound
		}
		return domain.Profile{}, err
	}
	return toDomain(document)
}

func (repository *Repository) Save(ctx context.Context, profile domain.Profile, expectedRevision uint64, commandID string) error {
	if !profile.HasCommand(commandID) || profile.Revision() != expectedRevision+1 {
		return domain.ErrInvalidProfile
	}
	document := fromDomain(profile)
	if expectedRevision == 0 {
		if _, err := repository.collection.InsertOne(ctx, document); err != nil {
			return repository.translateWriteError(ctx, profile.MemberID(), commandID, err)
		}
		return nil
	}
	result, err := repository.collection.ReplaceOne(ctx, bson.M{
		"_id": profile.MemberID(), "revision": expectedRevision,
	}, document)
	if err != nil {
		return repository.translateWriteError(ctx, profile.MemberID(), commandID, err)
	}
	if result.MatchedCount == 0 {
		return repository.classifyConflict(ctx, profile.MemberID(), commandID)
	}
	return nil
}

func (repository *Repository) translateWriteError(ctx context.Context, memberID, commandID string, err error) error {
	if !apimongo.IsDuplicateKey(err) {
		return err
	}
	return repository.classifyConflict(ctx, memberID, commandID)
}

func (repository *Repository) classifyConflict(ctx context.Context, memberID, commandID string) error {
	var found struct {
		Commands []commandDocument `bson:"commands"`
	}
	err := repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Decode(&found)
	if err == nil {
		return application.ErrCommandAlreadyApplied
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	err = repository.collection.FindOne(ctx, bson.M{"_id": memberID}).Err()
	if err == nil || errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrOptimisticConflict
	}
	return err
}

func fromDomain(profile domain.Profile) profileDocument {
	commands := make([]commandDocument, 0, len(profile.Commands()))
	for _, command := range profile.Commands() {
		commands = append(commands, commandDocument{
			ID: command.ID(), Fingerprint: command.Fingerprint(), Revision: command.Revision(),
		})
	}
	return profileDocument{
		MemberID: profile.MemberID(),
		DisplayName: fieldDocument{
			Value: profile.DisplayName().Value(), Visibility: profile.DisplayName().Visibility(),
			ConsentRef: profile.DisplayName().ConsentRef(),
		},
		Introduction: fieldDocument{
			Value: profile.Introduction().Value(), Visibility: profile.Introduction().Visibility(),
			ConsentRef: profile.Introduction().ConsentRef(),
		},
		Revision: profile.Revision(), UpdatedAt: profile.UpdatedAt(), Commands: commands,
	}
}

func toDomain(document profileDocument) (domain.Profile, error) {
	displayName, err := domain.NewField(document.DisplayName.Value, document.DisplayName.Visibility, document.DisplayName.ConsentRef, 80, true)
	if err != nil {
		return domain.Profile{}, err
	}
	introduction, err := domain.NewField(document.Introduction.Value, document.Introduction.Visibility, document.Introduction.ConsentRef, 280, false)
	if err != nil {
		return domain.Profile{}, err
	}
	commands := make([]domain.AppliedCommand, 0, len(document.Commands))
	for _, persisted := range document.Commands {
		command, err := domain.NewAppliedCommand(persisted.ID, persisted.Fingerprint, persisted.Revision)
		if err != nil {
			return domain.Profile{}, err
		}
		commands = append(commands, command)
	}
	return domain.Rehydrate(domain.State{
		MemberID: document.MemberID, DisplayName: displayName, Introduction: introduction,
		Revision: document.Revision, UpdatedAt: document.UpdatedAt, Commands: commands,
	})
}
