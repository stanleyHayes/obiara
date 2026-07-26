package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/relay/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository { return &Repository{d.Collection("cloth_relays")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true)})
	return e
}

type document struct {
	ID        string            `bson:"_id"`
	Members   []string          `bson:"members"`
	Reviewer  string            `bson:"reviewer"`
	Questions []domain.Question `bson:"questions"`
	Audit     []domain.Audit    `bson:"audit"`
	Commands  []domain.Applied  `bson:"commands"`
	Revision  uint64            `bson:"revision"`
}

func (r *Repository) Create(ctx context.Context, v domain.Relay) error {
	_, e := r.c.InsertOne(ctx, toDoc(v))
	return e
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Relay, error) {
	var d document
	if e := r.c.FindOne(ctx, bson.M{"_id": id}).Decode(&d); e != nil {
		return domain.Relay{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, Members: d.Members, Reviewer: d.Reviewer, Questions: d.Questions, Audit: d.Audit, Commands: d.Commands, Revision: d.Revision})
}
func (r *Repository) Save(ctx context.Context, v domain.Relay, expected uint64, cmd string) error {
	res, e := r.c.ReplaceOne(ctx, bson.M{"_id": v.ID(), "revision": expected}, toDoc(v))
	if mongo.IsDuplicateKeyError(e) {
		return nil
	}
	if e != nil {
		return e
	}
	if res.MatchedCount == 0 {
		return domain.ErrStaleRevision
	}
	return nil
}
func toDoc(v domain.Relay) document {
	return document{v.ID(), v.Members(), v.Reviewer(), v.Questions(), v.Audit(), v.Commands(), v.Revision()}
}
