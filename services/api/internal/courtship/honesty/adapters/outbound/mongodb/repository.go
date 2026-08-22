package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("courtship_honesty_ribbons")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("honesty_command")})
	return e
}

type document struct {
	ID              string `bson:"_id"`
	Members, Grants []string
	Revision        uint64
	Events          []domain.Event
	Commands        []domain.Applied
}

func (r *Repository) Create(ctx context.Context, x domain.Ribbon) error {
	_, e := r.collection.InsertOne(ctx, toDoc(x))
	return e
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Ribbon, error) {
	var d document
	e := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if e != nil {
		return domain.Ribbon{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, Members: d.Members, Grants: d.Grants, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Save(ctx context.Context, x domain.Ribbon, expected uint64, commandID string) error {
	events, commands := x.Events(), x.Commands()
	// An aggregate with no events has never been acted on, so there is
	// nothing to append and the caller wanted Create. Returning an error
	// beats indexing past the end of an empty slice, which in an HTTP
	// handler is a panic rather than a rejected request.
	if len(events) == 0 || len(commands) == 0 {
		return domain.ErrInvalid
	}
	result, e := r.collection.UpdateOne(ctx, bson.M{"_id": x.ID(), "revision": expected}, bson.M{"$set": bson.M{"grants": x.Grants(), "revision": x.Revision()}, "$push": bson.M{"events": events[len(events)-1], "commands": commands[len(commands)-1]}})
	if mongo.IsDuplicateKeyError(e) {
		return nil
	}
	if e != nil {
		return e
	}
	if result.MatchedCount == 0 {
		return domain.ErrStaleRevision
	}
	return nil
}
func toDoc(x domain.Ribbon) document {
	// A freshly opened ribbon has no events or commands yet. Nil slices
	// marshal to null, and $push cannot append to null, so the first action
	// on the room would fail with a write exception. Empty arrays keep the
	// document shape stable from the moment it is created.
	events, commands := x.Events(), x.Commands()
	if events == nil {
		events = []domain.Event{}
	}
	if commands == nil {
		commands = []domain.Applied{}
	}
	return document{x.ID(), x.Members(), x.Grants(), x.Revision(), events, commands}
}
