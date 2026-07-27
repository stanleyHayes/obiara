package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("courtship_closures")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("closure_command")})
	return err
}

type document struct {
	ID           string           `bson:"_id"`
	Members      []string         `bson:"members"`
	Status       domain.Status    `bson:"status"`
	LastActivity time.Time        `bson:"lastActivity"`
	ClosedAt     time.Time        `bson:"closedAt,omitempty"`
	Revision     uint64           `bson:"revision"`
	Event        *domain.Event    `bson:"event,omitempty"`
	Commands     []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, c domain.Closure) error {
	_, err := r.collection.InsertOne(ctx, toDoc(c))
	return err
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Closure, error) {
	var d document
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		return domain.Closure{}, err
	}
	return domain.Rehydrate(domain.State{ID: d.ID, Members: d.Members, Status: d.Status, LastActivity: d.LastActivity, ClosedAt: d.ClosedAt, Revision: d.Revision, Event: d.Event, Commands: d.Commands})
}
func (r *Repository) Save(ctx context.Context, c domain.Closure, expected uint64, commandID string) error {
	cmds := c.Commands()
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": c.ID(), "revision": expected}, bson.M{"$set": bson.M{"status": c.Status(), "closedAt": c.ClosedAt(), "revision": c.Revision(), "event": c.Event()}, "$push": bson.M{"commands": cmds[len(cmds)-1]}})
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrStaleRevision
	}
	return nil
}
func toDoc(c domain.Closure) document {
	cmds := c.Commands()
	if cmds == nil {
		cmds = []domain.Applied{}
	}
	return document{ID: c.ID(), Members: c.Members(), Status: c.Status(), LastActivity: c.LastActivity(), ClosedAt: c.ClosedAt(), Revision: c.Revision(), Event: c.Event(), Commands: cmds}
}
