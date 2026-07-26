package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository { return &Repository{d.Collection("fire_runsheets")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("runsheet_command_unique")}, {Keys: bson.D{{Key: "fireKey", Value: 1}, {Key: "version", Value: 1}}, Options: options.Index().SetUnique(true).SetName("runsheet_fire_version")}})
	return e
}

type doc struct {
	ID        string           `bson:"_id"`
	FireKey   string           `bson:"fireKey"`
	HostKey   string           `bson:"hostKey"`
	Version   uint64           `bson:"version"`
	Segments  []domain.Segment `bson:"segments"`
	Status    domain.Status    `bson:"status"`
	Current   int              `bson:"current"`
	StartedAt time.Time        `bson:"startedAt,omitempty"`
	EndsAt    time.Time        `bson:"endsAt,omitempty"`
	Revision  uint64           `bson:"revision"`
	Events    []domain.Event   `bson:"events"`
	Commands  []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, v domain.RunSheet) error {
	_, e := r.c.InsertOne(ctx, toDoc(v))
	return r.dupe(ctx, v.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.RunSheet, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.RunSheet, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.RunSheet, error) {
	var d doc
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.RunSheet{}, application.ErrNotFound
		}
		return domain.RunSheet{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, FireKey: d.FireKey, HostKey: d.HostKey, Version: d.Version, Segments: d.Segments, Status: d.Status, Current: d.Current, StartedAt: d.StartedAt, EndsAt: d.EndsAt, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, v domain.RunSheet, expected uint64, id string) error {
	es, cs := v.Events(), v.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": v.ID(), "revision": expected}, bson.M{"$set": bson.M{"status": v.Status(), "current": v.Current(), "startedAt": v.StartedAt(), "endsAt": v.EndsAt(), "revision": v.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
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
func toDoc(v domain.RunSheet) doc {
	return doc{ID: v.ID(), FireKey: v.FireKey(), HostKey: v.HostKey(), Version: v.Version(), Segments: v.Segments(), Status: v.Status(), Current: v.Current(), StartedAt: v.StartedAt(), EndsAt: v.EndsAt(), Revision: v.Revision(), Events: v.Events(), Commands: v.Commands()}
}
