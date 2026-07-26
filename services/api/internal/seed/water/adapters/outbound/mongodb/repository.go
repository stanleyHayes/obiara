package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("seed_mutual_water")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("water_command")}, {Keys: bson.D{{Key: "roomKey", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("water_private_room")}})
	return e
}

type eventDoc struct {
	Sequence   uint64    `bson:"sequence"`
	CommandID  string    `bson:"commandId"`
	ActorKey   string    `bson:"actorKey"`
	ReasonCode string    `bson:"reasonCode"`
	Mutual     bool      `bson:"mutual"`
	RoomKey    string    `bson:"roomKey,omitempty"`
	At         time.Time `bson:"at"`
}
type commandDoc struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type document struct {
	ID       string        `bson:"_id"`
	Members  []string      `bson:"members"`
	Watered  []string      `bson:"watered"`
	RoomKey  string        `bson:"roomKey,omitempty"`
	Status   domain.Status `bson:"status"`
	Revision uint64        `bson:"revision"`
	Events   []eventDoc    `bson:"events"`
	Commands []commandDoc  `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, w domain.Water) error {
	_, e := r.c.InsertOne(ctx, toDoc(w))
	return r.dupe(ctx, w.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Water, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Water, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Water, error) {
	var d document
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Water{}, application.ErrNotFound
		}
		return domain.Water{}, e
	}
	return toDomain(d)
}
func (r *Repository) Append(ctx context.Context, w domain.Water, expected uint64, id string) error {
	es, cs := w.Events(), w.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalidWater
	}
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": w.ID(), "revision": expected, "roomKey": ""}, bson.M{"$set": bson.M{"watered": w.Watered(), "roomKey": w.RoomKey(), "status": w.Status(), "revision": w.Revision()}, "$push": bson.M{"events": event(es[len(es)-1]), "commands": command(cs[len(cs)-1])}})
	if e != nil {
		return r.dupe(ctx, id, e)
	}
	if x.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
			return application.ErrCommandApplied
		}
		return application.ErrOptimisticConflict
	}
	return nil
}
func (r *Repository) dupe(ctx context.Context, id string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
		return application.ErrCommandApplied
	}
	return application.ErrOptimisticConflict
}
func toDoc(w domain.Water) document {
	d := document{ID: w.ID(), Members: w.Members(), Watered: w.Watered(), RoomKey: w.RoomKey(), Status: w.Status(), Revision: w.Revision()}
	for _, x := range w.Events() {
		d.Events = append(d.Events, event(x))
	}
	for _, x := range w.Commands() {
		d.Commands = append(d.Commands, command(x))
	}
	return d
}
func toDomain(d document) (domain.Water, error) {
	s := domain.State{ID: d.ID, Members: d.Members, Watered: d.Watered, RoomKey: d.RoomKey, Status: d.Status, Revision: d.Revision}
	for _, x := range d.Events {
		s.Events = append(s.Events, domain.Event{Sequence: x.Sequence, CommandID: x.CommandID, ActorKey: x.ActorKey, ReasonCode: x.ReasonCode, Mutual: x.Mutual, RoomKey: x.RoomKey, At: x.At})
	}
	for _, x := range d.Commands {
		s.Commands = append(s.Commands, domain.AppliedCommand{ID: x.ID, Fingerprint: x.Fingerprint, Revision: x.Revision})
	}
	return domain.Rehydrate(s)
}
func event(x domain.Event) eventDoc {
	return eventDoc{x.Sequence, x.CommandID, x.ActorKey, x.ReasonCode, x.Mutual, x.RoomKey, x.At}
}
func command(x domain.AppliedCommand) commandDoc { return commandDoc{x.ID, x.Fingerprint, x.Revision} }
