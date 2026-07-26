package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/room/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/room/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("courtship_rooms")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("courtship_room_command")}, {Keys: bson.D{{Key: "members", Value: 1}}, Options: options.Index().SetName("courtship_room_private_members")}})
	return e
}

type eventDoc struct {
	Sequence   uint64      `bson:"sequence"`
	CommandID  string      `bson:"commandId"`
	ActorKey   string      `bson:"actorKey"`
	ReasonCode string      `bson:"reasonCode"`
	Kind       domain.Kind `bson:"kind"`
	ContentKey string      `bson:"contentKey,omitempty"`
	At         time.Time   `bson:"at"`
}
type commandDoc struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type document struct {
	ID         string            `bson:"_id"`
	Members    []string          `bson:"members"`
	Events     []eventDoc        `bson:"events"`
	Commands   []commandDoc      `bson:"commands"`
	Revision   uint64            `bson:"revision"`
	Projection domain.Projection `bson:"projection"`
}

func (r *Repository) Create(ctx context.Context, x domain.Room) error {
	_, e := r.c.InsertOne(ctx, toDoc(x))
	return r.dupe(ctx, x.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Room, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Room, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Room, error) {
	var d document
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Room{}, application.ErrNotFound
		}
		return domain.Room{}, e
	}
	return toDomain(d)
}
func (r *Repository) Append(ctx context.Context, x domain.Room, expected uint64, id string) error {
	es, cs := x.Events(), x.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalidRoom
	}
	z, e := r.c.UpdateOne(ctx, bson.M{"_id": x.ID(), "revision": expected}, bson.M{"$set": bson.M{"revision": x.Revision(), "projection": x.Projection()}, "$push": bson.M{"events": event(es[len(es)-1]), "commands": command(cs[len(cs)-1])}})
	if e != nil {
		return r.dupe(ctx, id, e)
	}
	if z.MatchedCount == 0 {
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
func toDoc(x domain.Room) document {
	d := document{ID: x.ID(), Members: x.Members(), Revision: x.Revision(), Projection: x.Projection()}
	for _, v := range x.Events() {
		d.Events = append(d.Events, event(v))
	}
	for _, v := range x.Commands() {
		d.Commands = append(d.Commands, command(v))
	}
	return d
}
func toDomain(d document) (domain.Room, error) {
	s := domain.State{ID: d.ID, Members: d.Members}
	for _, x := range d.Events {
		s.Events = append(s.Events, domain.Event{Sequence: x.Sequence, CommandID: x.CommandID, ActorKey: x.ActorKey, ReasonCode: x.ReasonCode, Kind: x.Kind, ContentKey: x.ContentKey, At: x.At})
	}
	for _, x := range d.Commands {
		s.Commands = append(s.Commands, domain.AppliedCommand{ID: x.ID, Fingerprint: x.Fingerprint, Revision: x.Revision})
	}
	return domain.Rehydrate(s)
}
func event(x domain.Event) eventDoc {
	return eventDoc{x.Sequence, x.CommandID, x.ActorKey, x.ReasonCode, x.Kind, x.ContentKey, x.At}
}
func command(x domain.AppliedCommand) commandDoc { return commandDoc{x.ID, x.Fingerprint, x.Revision} }
