package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("courtship_pause_stones")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("pause_command")})
	return e
}

type document struct {
	ID           string           `bson:"_id"`
	Members      []string         `bson:"members"`
	Status       domain.Status    `bson:"status"`
	PausedBy     string           `bson:"pausedBy,omitempty"`
	Acknowledged []string         `bson:"acknowledged"`
	Revision     uint64           `bson:"revision"`
	Events       []domain.Event   `bson:"events"`
	Commands     []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, s domain.Stone) error {
	_, e := r.collection.InsertOne(ctx, toDoc(s))
	return e
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Stone, error) {
	var d document
	e := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if e != nil {
		return domain.Stone{}, e
	}
	return domain.Rehydrate(domain.State{
		ID: d.ID, Members: d.Members, Status: d.Status, PausedBy: d.PausedBy,
		Acknowledged: d.Acknowledged, Revision: d.Revision, Events: d.Events, Commands: d.Commands,
	})
}
func (r *Repository) Save(ctx context.Context, s domain.Stone, expected uint64, commandID string) error {
	events, commands := s.Events(), s.Commands()
	result, e := r.collection.UpdateOne(ctx, bson.M{"_id": s.ID(), "revision": expected}, bson.M{"$set": bson.M{"status": s.Status(), "pausedBy": s.PausedBy(), "acknowledged": s.Acknowledged(), "revision": s.Revision()}, "$push": bson.M{"events": events[len(events)-1], "commands": commands[len(commands)-1]}})
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
func toDoc(s domain.Stone) document {
	return document{s.ID(), s.Members(), s.Status(), s.PausedBy(), s.Acknowledged(), s.Revision(), s.Events(), s.Commands()}
}
