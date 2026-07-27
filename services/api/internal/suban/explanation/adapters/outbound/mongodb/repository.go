package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func New(db *mongo.Database) *Repository { return &Repository{db.Collection("suban_appeals")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "appliedids", Value: 1}}, Options: options.Index().SetUnique(true).SetName("suban_appeal_command_unique")}, {Keys: bson.D{{Key: "subjectkey", Value: 1}, {Key: "eventid", Value: 1}}, Options: options.Index().SetUnique(true).SetName("suban_event_appeal_unique")}})
	return e
}
func (r *Repository) Create(ctx context.Context, a domain.Appeal) error {
	s := a.State()
	_, e := r.c.InsertOne(ctx, s)
	return r.dupe(ctx, s.AppliedIDs[0], e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Appeal, error) {
	var s domain.State
	e := r.c.FindOne(ctx, bson.M{"id": id}).Decode(&s)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return domain.Appeal{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Appeal{}, e
	}
	return domain.Rehydrate(s)
}
func (r *Repository) Save(ctx context.Context, a domain.Appeal, expected uint64, command string) error {
	s := a.State()
	result, e := r.c.UpdateOne(ctx, bson.M{"id": s.ID, "revision": expected}, bson.M{"$set": bson.M{"status": s.Status, "reviewerkey": s.ReviewerKey, "reasoningref": s.ReasoningRef, "resolvedat": s.ResolvedAt, "revision": s.Revision}, "$push": bson.M{"audit": s.Audit[len(s.Audit)-1], "appliedids": command}})
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
func (r *Repository) dupe(ctx context.Context, id string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"appliedids": id}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
