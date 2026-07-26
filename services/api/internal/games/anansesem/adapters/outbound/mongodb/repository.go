package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("anansesem_stories")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("anansesem_command_unique")}, {Keys: bson.D{{Key: "roomKey", Value: 1}}, Options: options.Index().SetName("anansesem_private_room")}})
	return e
}

type doc struct {
	ID        string           `bson:"_id"`
	RoomKey   string           `bson:"roomKey"`
	TitleCode string           `bson:"titleCode"`
	Authors   []string         `bson:"authors"`
	Passages  []domain.Passage `bson:"passages"`
	Grants    []domain.Grant   `bson:"grants"`
	Editions  []domain.Edition `bson:"editions"`
	Revision  uint64           `bson:"revision"`
	Events    []domain.Event   `bson:"events"`
	Commands  []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, s domain.Story) error {
	_, e := r.c.InsertOne(ctx, toDoc(s))
	return r.dupe(ctx, s.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Story, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Story, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Story, error) {
	var d doc
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Story{}, application.ErrNotFound
		}
		return domain.Story{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, RoomKey: d.RoomKey, TitleCode: d.TitleCode, Authors: d.Authors, Passages: d.Passages, Grants: d.Grants, Editions: d.Editions, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, s domain.Story, expected uint64, id string) error {
	es, cs := s.Events(), s.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": s.ID(), "revision": expected}, bson.M{"$set": bson.M{"passages": s.Passages(), "grants": s.Grants(), "editions": s.Editions(), "revision": s.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
	if e != nil {
		return r.dupe(ctx, id, e)
	}
	if x.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
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
	if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func toDoc(s domain.Story) doc {
	return doc{ID: s.ID(), RoomKey: s.RoomKey(), TitleCode: s.TitleCode(), Authors: s.Authors(), Passages: s.Passages(), Grants: s.Grants(), Editions: s.Editions(), Revision: s.Revision(), Events: s.Events(), Commands: s.Commands()}
}
