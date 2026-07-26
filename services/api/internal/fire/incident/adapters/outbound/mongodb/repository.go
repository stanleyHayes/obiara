package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/incident/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/incident/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository { return &Repository{d.Collection("fire_incidents")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("incident_command_unique")}})
	return e
}

type doc struct {
	CaseID      string           `bson:"_id"`
	FireKey     string           `bson:"fireKey"`
	ActorKey    string           `bson:"actorKey"`
	Category    domain.Category  `bson:"category"`
	EvidenceRef string           `bson:"evidenceRef,omitempty"`
	OccurredAt  time.Time        `bson:"occurredAt"`
	Status      domain.Status    `bson:"status"`
	RoutedAt    time.Time        `bson:"routedAt,omitempty"`
	Revision    uint64           `bson:"revision"`
	Events      []domain.Event   `bson:"events"`
	Commands    []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, i domain.Incident) error {
	_, e := r.c.InsertOne(ctx, toDoc(i))
	return r.dupe(ctx, i.Commands()[0].ID, e)
}
func (r *Repository) FindByCase(ctx context.Context, id string) (domain.Incident, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Incident, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Incident, error) {
	var d doc
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Incident{}, application.ErrNotFound
		}
		return domain.Incident{}, e
	}
	return domain.Rehydrate(domain.State{CaseID: d.CaseID, FireKey: d.FireKey, ActorKey: d.ActorKey, Category: d.Category, EvidenceRef: d.EvidenceRef, OccurredAt: d.OccurredAt, Status: d.Status, RoutedAt: d.RoutedAt, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, i domain.Incident, expected uint64, id string) error {
	es, cs := i.Events(), i.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": i.CaseID(), "revision": expected}, bson.M{"$set": bson.M{"status": i.Status(), "routedAt": i.RoutedAt(), "revision": i.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
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
func toDoc(i domain.Incident) doc {
	return doc{CaseID: i.CaseID(), FireKey: i.FireKey(), ActorKey: i.ActorKey(), Category: i.Category(), EvidenceRef: i.EvidenceRef(), OccurredAt: i.OccurredAt(), Status: i.Status(), RoutedAt: i.RoutedAt(), Revision: i.Revision(), Events: i.Events(), Commands: i.Commands()}
}
