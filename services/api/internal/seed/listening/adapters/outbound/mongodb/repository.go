// Package mongodb persists playback records. Eligibility is derived from
// merged intervals; the cached total is never authoritative on its own.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) collection() *mongo.Collection {
	return repository.database.Collection("playback_records")
}

type document struct {
	ID         string            `bson:"_id"`
	ListenerID string            `bson:"listenerId"`
	AssetID    string            `bson:"assetId"`
	Duration   float64           `bson:"duration"`
	Intervals  []domain.Interval `bson:"intervals"`
	Version    int64             `bson:"version"`
	UpdatedAt  time.Time         `bson:"updatedAt"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "listenerId", Value: 1}, {Key: "assetId", Value: 1}},
		Options: options.Index().SetName("playback_listener_asset_unique").SetUnique(true),
	})
	return err
}

func key(listenerID, assetID string) string {
	return listenerID + "|" + assetID
}

func (repository *Repository) Find(ctx context.Context, listenerID, assetID string) (domain.Playback, error) {
	var doc document
	if err := repository.collection().FindOne(ctx, bson.M{"_id": key(listenerID, assetID)}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Playback{}, domain.ErrPlaybackNotFound
		}
		return domain.Playback{}, err
	}
	return domain.ReconstitutePlayback(doc.ListenerID, doc.AssetID, doc.Duration, doc.Intervals, doc.Version, doc.UpdatedAt), nil
}

// Save inserts new records and updates existing ones pinned to the read
// version (the Record call increments version once per batch).
func (repository *Repository) Save(ctx context.Context, playback domain.Playback) error {
	doc := document{
		ID:         key(playback.ListenerID(), playback.AssetID()),
		ListenerID: playback.ListenerID(),
		AssetID:    playback.AssetID(),
		Duration:   playback.Duration(),
		Intervals:  playback.Intervals(),
		Version:    playback.Version(),
		UpdatedAt:  playback.UpdatedAt(),
	}
	if playback.Version() == 1 {
		_, err := repository.collection().InsertOne(ctx, doc)
		if apimongo.IsDuplicateKey(err) {
			return domain.ErrStalePlayback
		}
		return err
	}
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{"_id": doc.ID, "version": playback.Version() - 1},
		bson.M{"$set": bson.M{"intervals": doc.Intervals, "version": doc.Version, "updatedAt": doc.UpdatedAt}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrStalePlayback
	}
	return nil
}
