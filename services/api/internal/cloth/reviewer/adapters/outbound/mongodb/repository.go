package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("cloth_reviewer_access")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("reviewer_command_unique")},
		{Keys: bson.D{{Key: "inviteDigest", Value: 1}}, Options: options.Index().SetUnique(true).SetName("reviewer_invite_unique")},
		{Keys: bson.D{{Key: "sessionDigest", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("reviewer_session_unique")},
	})
	return err
}

type document struct {
	ID              string                  `bson:"_id"`
	Members         []string                `bson:"members"`
	ReviewerKey     string                  `bson:"reviewerKey"`
	InviteDigest    string                  `bson:"inviteDigest"`
	OTPDigest       string                  `bson:"otpDigest"`
	SessionDigest   string                  `bson:"sessionDigest,omitempty"`
	WatermarkRef    string                  `bson:"watermarkRef"`
	QuestionRefs    []string                `bson:"questionRefs"`
	MaterialRefs    []string                `bson:"materialRefs"`
	Status          domain.Status           `bson:"status"`
	OTPExpiresAt    time.Time               `bson:"otpExpiresAt"`
	InviteExpiresAt time.Time               `bson:"inviteExpiresAt"`
	RedeemedAt      time.Time               `bson:"redeemedAt,omitempty"`
	RevokedAt       time.Time               `bson:"revokedAt,omitempty"`
	Revision        uint64                  `bson:"revision"`
	Commands        []domain.AppliedCommand `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, review domain.Review) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(review))
	return repository.classifyDuplicate(ctx, review.Commands()[0].ID, err)
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Review, error) {
	return repository.find(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindByCommand(ctx context.Context, id string) (domain.Review, error) {
	return repository.find(ctx, bson.M{"commands.id": id})
}

func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Review, error) {
	var stored document
	if err := repository.collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Review{}, application.ErrNotFound
		}
		return domain.Review{}, err
	}
	return domain.Rehydrate(toState(stored))
}

func (repository *Repository) Append(ctx context.Context, review domain.Review, expected uint64, commandID string) error {
	commands := review.Commands()
	if len(commands) != int(expected+1) {
		return domain.ErrInvalid
	}
	result, err := repository.collection.UpdateOne(ctx,
		bson.M{"_id": review.ID(), "revision": expected},
		bson.M{
			"$set": bson.M{
				"status": review.Status(), "sessionDigest": review.SessionDigest(),
				"redeemedAt": review.RedeemedAt(), "revokedAt": review.RevokedAt(),
				"revision": review.Revision(),
			},
			"$push": bson.M{"commands": commands[len(commands)-1]},
		},
	)
	if err != nil {
		return repository.classifyDuplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		if repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
			return application.ErrCommandApplied
		}
		return application.ErrOptimisticConflict
	}
	return nil
}

func (repository *Repository) classifyDuplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
		return application.ErrCommandApplied
	}
	return application.ErrOptimisticConflict
}

func toDocument(review domain.Review) document {
	return document{
		ID: review.ID(), Members: review.Members(), ReviewerKey: review.ReviewerKey(),
		InviteDigest: review.InviteDigest(), OTPDigest: review.OTPDigest(), SessionDigest: review.SessionDigest(),
		WatermarkRef: review.WatermarkRef(), QuestionRefs: review.QuestionRefs(), MaterialRefs: review.MaterialRefs(),
		Status: review.Status(), OTPExpiresAt: review.OTPExpiresAt(), InviteExpiresAt: review.InviteExpiresAt(),
		RedeemedAt: review.RedeemedAt(), RevokedAt: review.RevokedAt(), Revision: review.Revision(),
		Commands: review.Commands(),
	}
}

func toState(stored document) domain.State {
	return domain.State{
		ID: stored.ID, Members: stored.Members, ReviewerKey: stored.ReviewerKey,
		InviteDigest: stored.InviteDigest, OTPDigest: stored.OTPDigest, SessionDigest: stored.SessionDigest,
		WatermarkRef: stored.WatermarkRef, QuestionRefs: stored.QuestionRefs, MaterialRefs: stored.MaterialRefs,
		Status: stored.Status, OTPExpiresAt: stored.OTPExpiresAt, InviteExpiresAt: stored.InviteExpiresAt,
		RedeemedAt: stored.RedeemedAt, RevokedAt: stored.RevokedAt, Revision: stored.Revision,
		Commands: stored.Commands,
	}
}
