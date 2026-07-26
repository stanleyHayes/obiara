package mongodb

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/safety/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/safety/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	buckets *mongo.Collection
	signals *mongo.Collection
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{buckets: database.Collection("seed_safety_buckets"), signals: database.Collection("seed_care_signals")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.signals.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "actorKey", Value: 1}, {Key: "windowRevision", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("care_signal_once"),
	})
	return err
}

func (repository *Repository) Find(ctx context.Context, actorKey string) (domain.Bucket, error) {
	var bucket domain.Bucket
	err := repository.buckets.FindOne(ctx, bson.M{"_id": actorKey}).Decode(&bucket)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Bucket{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Bucket{}, err
	}
	return domain.Rehydrate(bucket)
}

func (repository *Repository) Create(ctx context.Context, bucket domain.Bucket) error {
	_, err := repository.buckets.InsertOne(ctx, bson.M{
		"_id": bucket.ActorKey, "actorKey": bucket.ActorKey, "windowStarted": bucket.WindowStarted,
		"sows": bucket.Sows, "candidates": bucket.Candidates, "denials": bucket.Denials, "revision": bucket.Revision,
	})
	return err
}

func (repository *Repository) Save(ctx context.Context, bucket domain.Bucket, expected uint64) error {
	result, err := repository.buckets.UpdateOne(ctx, bson.M{"_id": bucket.ActorKey, "revision": expected}, bson.M{"$set": bson.M{
		"windowStarted": bucket.WindowStarted, "sows": bucket.Sows, "candidates": bucket.Candidates,
		"denials": bucket.Denials, "revision": bucket.Revision,
	}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrStaleRevision
	}
	return nil
}

func (repository *Repository) AppendCareSignal(ctx context.Context, signal application.CareSignal) error {
	_, err := repository.signals.InsertOne(ctx, bson.M{
		"actorKey": signal.ActorKey, "code": signal.Code, "windowRevision": signal.WindowRevision,
	})
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}
