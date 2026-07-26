package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/application"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("victim_export_authorizations")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("victim_export_command_unique")}, {Keys: bson.D{{Key: "authorization.tokenKey", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("victim_export_token_unique")}, {Keys: bson.D{{Key: "authorization.expiresAt", Value: 1}}, Options: options.Index().SetName("victim_export_expiry_audit_no_ttl")}})
	return e
}

type doc struct {
	ID            string               `bson:"_id"`
	MemberKey     string               `bson:"memberKey"`
	Purpose       domain.Purpose       `bson:"purpose"`
	References    []domain.Reference   `bson:"references"`
	Status        domain.Status        `bson:"status"`
	Authorization domain.Authorization `bson:"authorization"`
	Revision      uint64               `bson:"revision"`
	Events        []domain.Event       `bson:"events"`
	Commands      []domain.Applied     `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, e domain.Export) error {
	_, x := r.c.InsertOne(ctx, toDoc(e))
	return r.dupe(ctx, e.Commands()[0].ID, x)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Export, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Export, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Export, error) {
	var d doc
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Export{}, application.ErrNotFound
		}
		return domain.Export{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, MemberKey: d.MemberKey, Purpose: d.Purpose, References: d.References, Status: d.Status, Authorization: d.Authorization, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, e domain.Export, expected uint64, command string) error {
	es, cs := e.Events(), e.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	res, x := r.c.UpdateOne(ctx, bson.M{"_id": e.ID(), "revision": expected}, bson.M{"$set": bson.M{"references": e.References(), "status": e.Status(), "authorization": e.Authorization(), "revision": e.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
	if x != nil {
		return r.dupe(ctx, command, x)
	}
	if res.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
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
	if r.c.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func toDoc(e domain.Export) doc {
	return doc{ID: e.ID(), MemberKey: e.MemberKey(), Purpose: e.Purpose(), References: e.References(), Status: e.Status(), Authorization: e.Authorization(), Revision: e.Revision(), Events: e.Events(), Commands: e.Commands()}
}
