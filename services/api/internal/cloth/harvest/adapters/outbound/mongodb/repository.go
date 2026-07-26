package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository { return &Repository{d.Collection("cloth_harvests")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("harvest_command_unique")},
		{Keys: bson.D{{Key: "handoffId", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("harvest_handoff_unique")},
	})
	return e
}

type document struct {
	ID        string           `bson:"_id"`
	HandoffID string           `bson:"handoffId,omitempty"`
	Members   []string         `bson:"members"`
	Payload   domain.Payload   `bson:"payload"`
	Approvals []string         `bson:"approvals"`
	Status    domain.Status    `bson:"status"`
	ReadyAt   time.Time        `bson:"readyAt,omitempty"`
	ExpiresAt time.Time        `bson:"expiresAt,omitempty"`
	Revision  uint64           `bson:"revision"`
	Events    []domain.Event   `bson:"events"`
	Commands  []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, h domain.Harvest) error {
	_, e := r.c.InsertOne(ctx, toDoc(h))
	return r.dupe(ctx, h.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Harvest, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Harvest, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) FindByHandoff(ctx context.Context, id string) (domain.Harvest, error) {
	return r.find(ctx, bson.M{"handoffId": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Harvest, error) {
	var d struct {
		ID        string           `bson:"_id"`
		HandoffID string           `bson:"handoffId"`
		Members   []string         `bson:"members"`
		Payload   domain.Payload   `bson:"payload"`
		Approvals []string         `bson:"approvals"`
		Status    domain.Status    `bson:"status"`
		ReadyAt   time.Time        `bson:"readyAt"`
		ExpiresAt time.Time        `bson:"expiresAt"`
		Revision  uint64           `bson:"revision"`
		Events    []domain.Event   `bson:"events"`
		Commands  []domain.Applied `bson:"commands"`
	}
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Harvest{}, application.ErrNotFound
		}
		return domain.Harvest{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, HandoffID: d.HandoffID, Members: d.Members, Payload: d.Payload, Approvals: d.Approvals, Status: d.Status, ReadyAt: d.ReadyAt, ExpiresAt: d.ExpiresAt, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, h domain.Harvest, expected uint64, id string) error {
	es, cs := h.Events(), h.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": h.ID(), "revision": expected}, bson.M{"$set": bson.M{"handoffId": h.HandoffID(), "payload": h.Payload(), "approvals": h.Approvals(), "status": h.Status(), "readyAt": h.ReadyAt(), "expiresAt": h.ExpiresAt(), "revision": h.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
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
func toDoc(h domain.Harvest) bson.M {
	return bson.M{"_id": h.ID(), "members": h.Members(), "payload": h.Payload(), "approvals": h.Approvals(), "status": h.Status(), "revision": h.Revision(), "events": h.Events(), "commands": h.Commands()}
}
