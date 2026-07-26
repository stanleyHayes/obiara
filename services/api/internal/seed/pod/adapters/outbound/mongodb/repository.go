package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository { return &Repository{d.Collection("seed_pods")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("seed_pod_command")}, {Keys: bson.D{{Key: "expiresAt", Value: 1}, {Key: "status", Value: 1}}, Options: options.Index().SetName("seed_pod_expiry")}})
	return e
}

type eventDoc struct {
	Sequence   uint64        `bson:"sequence"`
	CommandID  string        `bson:"commandId"`
	ActorKey   string        `bson:"actorKey"`
	ReasonCode string        `bson:"reasonCode"`
	Action     domain.Action `bson:"action"`
	At         time.Time     `bson:"at"`
}
type commandDoc struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type document struct {
	ID            string        `bson:"_id"`
	OwnerKey      string        `bson:"ownerKey"`
	MediaKey      string        `bson:"mediaKey"`
	RecipientKeys []string      `bson:"recipientKeys"`
	Status        domain.Status `bson:"status"`
	ExpiresAt     time.Time     `bson:"expiresAt"`
	EndedAt       *time.Time    `bson:"endedAt,omitempty"`
	Revision      uint64        `bson:"revision"`
	Events        []eventDoc    `bson:"events"`
	Commands      []commandDoc  `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, p domain.Pod) error {
	_, e := r.c.InsertOne(ctx, toDoc(p))
	return r.duplicate(ctx, p.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Pod, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Pod, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Pod, error) {
	var d document
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Pod{}, application.ErrNotFound
		}
		return domain.Pod{}, e
	}
	return toDomain(d)
}
func (r *Repository) Append(ctx context.Context, p domain.Pod, expected uint64, id string) error {
	es, cs := p.Events(), p.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalidPod
	}
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": p.ID(), "revision": expected}, bson.M{"$set": bson.M{"status": p.Status(), "revision": p.Revision(), "endedAt": p.EndedAt()}, "$push": bson.M{"events": event(es[len(es)-1]), "commands": command(cs[len(cs)-1])}})
	if e != nil {
		return r.duplicate(ctx, id, e)
	}
	if x.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
			return application.ErrCommandApplied
		}
		return application.ErrOptimisticConflict
	}
	return nil
}
func (r *Repository) duplicate(ctx context.Context, id string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
		return application.ErrCommandApplied
	}
	return application.ErrOptimisticConflict
}
func toDoc(p domain.Pod) document {
	d := document{ID: p.ID(), OwnerKey: p.OwnerKey(), MediaKey: p.MediaKey(), RecipientKeys: p.RecipientKeys(), Status: p.Status(), ExpiresAt: p.ExpiresAt(), EndedAt: p.EndedAt(), Revision: p.Revision()}
	for _, x := range p.Events() {
		d.Events = append(d.Events, event(x))
	}
	for _, x := range p.Commands() {
		d.Commands = append(d.Commands, command(x))
	}
	return d
}
func toDomain(d document) (domain.Pod, error) {
	s := domain.State{ID: d.ID, OwnerKey: d.OwnerKey, MediaKey: d.MediaKey, RecipientKeys: d.RecipientKeys, Status: d.Status, ExpiresAt: d.ExpiresAt, EndedAt: d.EndedAt, Revision: d.Revision}
	for _, x := range d.Events {
		s.Events = append(s.Events, domain.Event{Sequence: x.Sequence, CommandID: x.CommandID, ActorKey: x.ActorKey, ReasonCode: x.ReasonCode, Action: x.Action, At: x.At})
	}
	for _, x := range d.Commands {
		s.Commands = append(s.Commands, domain.AppliedCommand{ID: x.ID, Fingerprint: x.Fingerprint, Revision: x.Revision})
	}
	return domain.Rehydrate(s)
}
func event(x domain.Event) eventDoc {
	return eventDoc{x.Sequence, x.CommandID, x.ActorKey, x.ReasonCode, x.Action, x.At}
}
func command(x domain.AppliedCommand) commandDoc { return commandDoc{x.ID, x.Fingerprint, x.Revision} }
