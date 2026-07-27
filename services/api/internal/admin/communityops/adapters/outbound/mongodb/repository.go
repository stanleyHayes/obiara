package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/application"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func New(db *mongo.Database) *Repository {
	return &Repository{db.Collection("community_operation_proposals")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "appliedids", Value: 1}}, Options: options.Index().SetUnique(true).SetName("community_operation_command_unique")})
	return e
}
func (r *Repository) Create(ctx context.Context, p domain.Proposal) error {
	s := p.State()
	_, e := r.c.InsertOne(ctx, s)
	return r.dupe(ctx, s.AppliedIDs[0], e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Proposal, error) {
	var s domain.State
	e := r.c.FindOne(ctx, bson.M{"id": id}).Decode(&s)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return domain.Proposal{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Proposal{}, e
	}
	return domain.Rehydrate(s)
}
func (r *Repository) Save(ctx context.Context, p domain.Proposal, expected uint64, command string) error {
	s := p.State()
	result, e := r.c.UpdateOne(ctx, bson.M{"id": s.ID, "revision": expected}, bson.M{"$set": bson.M{"noticeacknowledged": s.NoticeAcknowledged, "revision": s.Revision}, "$push": bson.M{"audit": s.Audit[len(s.Audit)-1], "appliedids": command}})
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
