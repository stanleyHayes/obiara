package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/application"
	"github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func New(db *mongo.Database) *Repository {
	return &Repository{db.Collection("launch_readiness_snapshots")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "appliedids", Value: 1}}, Options: options.Index().SetUnique(true).SetName("launch_readiness_command_unique")})
	return e
}
func (r *Repository) Create(ctx context.Context, s domain.Snapshot) error {
	state := s.State()
	_, e := r.c.InsertOne(ctx, state)
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"appliedids": state.AppliedIDs[0]}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Snapshot, error) {
	var s domain.State
	e := r.c.FindOne(ctx, bson.M{"id": id}).Decode(&s)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return domain.Snapshot{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Snapshot{}, e
	}
	return domain.Rehydrate(s)
}
