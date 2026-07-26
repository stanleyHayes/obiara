package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/garden/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("seed_garden")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "ownerKey", Value: 1}, {Key: "seedKey", Value: 1}}, Options: options.Index().SetUnique(true).SetName("garden_owner_seed")},
		{Keys: bson.D{{Key: "ownerKey", Value: 1}, {Key: "expiresAt", Value: 1}, {Key: "state", Value: 1}}, Options: options.Index().SetName("garden_owner_expiry")},
	})
	return err
}

type document struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	SeedKey   string        `bson:"seedKey"`
	OwnerKey  string        `bson:"ownerKey"`
	State     domain.State  `bson:"state"`
	ExpiresAt time.Time     `bson:"expiresAt"`
	UpdatedAt time.Time     `bson:"updatedAt"`
	Revision  uint64        `bson:"revision"`
}

func (repository *Repository) Create(ctx context.Context, item domain.Item) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(item))
	return err
}

func (repository *Repository) Find(ctx context.Context, ownerKey, seedKey string) (domain.Item, error) {
	var found document
	err := repository.collection.FindOne(ctx, bson.M{"ownerKey": ownerKey, "seedKey": seedKey}).Decode(&found)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Item{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Item{}, err
	}
	return domain.Rehydrate(toDomain(found))
}

func (repository *Repository) Save(ctx context.Context, item domain.Item, expected uint64) error {
	result, err := repository.collection.UpdateOne(ctx,
		bson.M{"ownerKey": item.OwnerKey, "seedKey": item.SeedKey, "revision": expected},
		bson.M{"$set": bson.M{"state": item.State, "updatedAt": item.UpdatedAt, "revision": item.Revision}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrStaleRevision
	}
	return nil
}

func (repository *Repository) ListOwner(ctx context.Context, ownerKey string) ([]domain.Item, error) {
	cursor, err := repository.collection.Find(ctx, bson.M{"ownerKey": ownerKey},
		options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetLimit(100))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []document
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	items := make([]domain.Item, 0, len(documents))
	for _, stored := range documents {
		item, itemErr := domain.Rehydrate(toDomain(stored))
		if itemErr != nil {
			return nil, itemErr
		}
		items = append(items, item)
	}
	return items, nil
}

func (repository *Repository) ExpireDue(ctx context.Context, ownerKey string, at time.Time, limit int) (int64, error) {
	if limit < 1 || limit > 100 {
		return 0, domain.ErrInvalidProjection
	}
	cursor, err := repository.collection.Find(ctx, bson.M{
		"ownerKey": ownerKey, "expiresAt": bson.M{"$lte": at},
		"state": bson.M{"$in": []domain.State{domain.StateQueued, domain.StateDelivered, domain.StateHeard}},
	}, options.Find().SetProjection(bson.M{"_id": 1, "revision": 1}).SetLimit(int64(limit)))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	var due []struct {
		ID       bson.ObjectID `bson:"_id"`
		Revision uint64        `bson:"revision"`
	}
	if err = cursor.All(ctx, &due); err != nil {
		return 0, err
	}
	var changed int64
	for _, candidate := range due {
		result, updateErr := repository.collection.UpdateOne(ctx,
			bson.M{"_id": candidate.ID, "revision": candidate.Revision},
			bson.M{"$set": bson.M{"state": domain.StateExpired, "updatedAt": at.UTC()}, "$inc": bson.M{"revision": 1}},
		)
		if updateErr != nil {
			return changed, updateErr
		}
		changed += result.ModifiedCount
	}
	return changed, nil
}

func toDocument(item domain.Item) document {
	return document{SeedKey: item.SeedKey, OwnerKey: item.OwnerKey, State: item.State, ExpiresAt: item.ExpiresAt, UpdatedAt: item.UpdatedAt, Revision: item.Revision}
}

func toDomain(stored document) domain.Item {
	return domain.Item{SeedKey: stored.SeedKey, OwnerKey: stored.OwnerKey, State: stored.State, ExpiresAt: stored.ExpiresAt, UpdatedAt: stored.UpdatedAt, Revision: stored.Revision}
}
