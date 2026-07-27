package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func New(db *mongo.Database) *Repository { return &Repository{db.Collection("diaspora_checkouts")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "appliedids", Value: 1}}, Options: options.Index().SetUnique(true).SetName("diaspora_command_unique")}, {Keys: bson.D{{Key: "requestref", Value: 1}}, Options: options.Index().SetUnique(true).SetName("diaspora_request_unique")}, {Keys: bson.D{{Key: "providerref", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("diaspora_provider_unique")}})
	return e
}
func (r *Repository) Create(ctx context.Context, c domain.Checkout) error {
	s := c.State()
	_, e := r.c.InsertOne(ctx, s)
	return r.dupe(ctx, s.AppliedIDs[0], e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Checkout, error) {
	var s domain.State
	e := r.c.FindOne(ctx, bson.M{"id": id}).Decode(&s)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return domain.Checkout{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Checkout{}, e
	}
	return domain.Rehydrate(s)
}
func (r *Repository) Save(ctx context.Context, c domain.Checkout, expected uint64, command string) error {
	s := c.State()
	result, e := r.c.ReplaceOne(ctx, bson.M{"id": s.ID, "revision": expected}, s)
	if e != nil {
		return r.dupe(ctx, command, e)
	}
	if result.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"appliedids": command}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) dupe(ctx context.Context, command string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"appliedids": command}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
