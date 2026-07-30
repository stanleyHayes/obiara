package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/domain"
)

type Repository struct {
	prompts *mongo.Collection
	duels   *mongo.Collection
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{
		prompts: database.Collection("ebe_approved_prompts"),
		duels:   database.Collection("ebe_duels"),
	}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.prompts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "promptId", Value: 1}, {Key: "version", Value: 1}}, Options: options.Index().SetUnique(true).SetName("ebe_prompt_version_unique")},
		{Keys: bson.D{{Key: "review.reviewedAt", Value: -1}}, Options: options.Index().SetName("ebe_prompt_reviewed")},
	}); err != nil {
		return err
	}
	_, err := repository.duels.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "createCommandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("ebe_create_command_unique")},
		{Keys: bson.D{{Key: "commandIds", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("ebe_answer_command_unique")},
		{Keys: bson.D{{Key: "roomKey", Value: 1}}, Options: options.Index().SetName("ebe_room")},
	})
	return err
}

type promptDocument struct {
	PromptID        string        `bson:"promptId"`
	Version         uint64        `bson:"version"`
	Language        string        `bson:"language"`
	Cue             string        `bson:"cue"`
	AcceptedAnswers []string      `bson:"acceptedAnswers"`
	Source          domain.Source `bson:"source"`
	Review          domain.Review `bson:"review"`
}

func promptDoc(prompt domain.Prompt) promptDocument {
	spec := prompt.Spec()
	return promptDocument{
		PromptID: spec.ID, Version: spec.Version, Language: spec.Language,
		Cue: spec.Cue, AcceptedAnswers: spec.AcceptedAnswers,
		Source: spec.Source, Review: spec.Review,
	}
}

func (document promptDocument) prompt() (domain.Prompt, error) {
	return domain.NewApprovedPrompt(domain.PromptSpec{
		ID: document.PromptID, Version: document.Version,
		Language: document.Language, Cue: document.Cue,
		AcceptedAnswers: document.AcceptedAnswers,
		Source:          document.Source, Review: document.Review,
	})
}

func (repository *Repository) SaveApproved(ctx context.Context, prompt domain.Prompt) error {
	document := promptDoc(prompt)
	_, err := repository.prompts.InsertOne(ctx, document)
	if mongo.IsDuplicateKeyError(err) {
		var existing promptDocument
		if findErr := repository.prompts.FindOne(ctx, bson.M{
			"promptId": document.PromptID, "version": document.Version,
		}).Decode(&existing); findErr == nil && promptDocEqual(existing, document) {
			return nil
		}
		return application.ErrConflict
	}
	return err
}

func (repository *Repository) ListApproved(ctx context.Context, limit int) ([]domain.Prompt, error) {
	cursor, err := repository.prompts.Find(
		ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "promptId", Value: 1}, {Key: "version", Value: -1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []promptDocument
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	prompts := make([]domain.Prompt, 0, len(documents))
	for _, document := range documents {
		prompt, promptErr := document.prompt()
		if promptErr != nil {
			return nil, promptErr
		}
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

type duelDocument struct {
	ID              string           `bson:"_id"`
	RoomKey         string           `bson:"roomKey"`
	PlayerKeys      [2]string        `bson:"playerKeys"`
	Prompts         []promptDocument `bson:"prompts"`
	Turns           []domain.Turn    `bson:"turns"`
	Revision        uint64           `bson:"revision"`
	CreateCommandID string           `bson:"createCommandId"`
	CommandIDs      []string         `bson:"commandIds"`
}

func (repository *Repository) Create(ctx context.Context, stored application.StoredDuel, commandID string) error {
	spec := stored.Duel.Specification()
	prompts := make([]promptDocument, 0, len(spec.Prompts))
	for _, prompt := range spec.Prompts {
		prompts = append(prompts, promptDoc(prompt))
	}
	_, err := repository.duels.InsertOne(ctx, duelDocument{
		ID: spec.ID, RoomKey: stored.RoomKey, PlayerKeys: spec.PlayerKeys,
		Prompts: prompts, Turns: stored.Duel.Turns(),
		Revision: stored.Duel.Revision(), CreateCommandID: commandID,
	})
	return repository.duplicate(ctx, commandID, err)
}

func (repository *Repository) Find(ctx context.Context, id string) (application.StoredDuel, error) {
	return repository.find(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindByCommand(ctx context.Context, commandID string) (application.StoredDuel, error) {
	return repository.find(ctx, bson.M{"$or": bson.A{
		bson.M{"createCommandId": commandID}, bson.M{"commandIds": commandID},
	}})
}

func (repository *Repository) find(ctx context.Context, filter bson.M) (application.StoredDuel, error) {
	var document duelDocument
	if err := repository.duels.FindOne(ctx, filter).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return application.StoredDuel{}, application.ErrNotFound
		}
		return application.StoredDuel{}, err
	}
	prompts := make([]domain.Prompt, 0, len(document.Prompts))
	for _, stored := range document.Prompts {
		prompt, err := stored.prompt()
		if err != nil {
			return application.StoredDuel{}, err
		}
		prompts = append(prompts, prompt)
	}
	duel, err := domain.Replay(domain.DuelSpec{
		ID: document.ID, PlayerKeys: document.PlayerKeys, Prompts: prompts,
	}, document.Turns)
	if err != nil {
		return application.StoredDuel{}, err
	}
	return application.StoredDuel{Duel: duel, RoomKey: document.RoomKey}, nil
}

func (repository *Repository) Append(ctx context.Context, stored application.StoredDuel, expected uint64, commandID string) error {
	result, err := repository.duels.UpdateOne(
		ctx,
		bson.M{"_id": stored.Duel.Specification().ID, "revision": expected},
		bson.M{
			"$set":  bson.M{"turns": stored.Duel.Turns(), "revision": stored.Duel.Revision()},
			"$push": bson.M{"commandIds": commandID},
		},
	)
	if err != nil {
		return repository.duplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		if repository.duels.FindOne(ctx, bson.M{"commandIds": commandID}).Err() == nil {
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
	if repository.duels.FindOne(ctx, bson.M{"$or": bson.A{
		bson.M{"createCommandId": commandID}, bson.M{"commandIds": commandID},
	}}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}

func promptDocEqual(left, right promptDocument) bool {
	leftEncoded, leftErr := bson.Marshal(left)
	rightEncoded, rightErr := bson.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftEncoded) == string(rightEncoded)
}
