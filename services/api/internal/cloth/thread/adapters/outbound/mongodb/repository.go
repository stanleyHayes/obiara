package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/thread/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("cloth_threads")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("thread_command")})
	return err
}

type document struct {
	ID         string             `bson:"_id"`
	Members    []string           `bson:"members"`
	Provenance *domain.Provenance `bson:"provenance,omitempty"`
	Revision   uint64             `bson:"revision"`
	Commands   []domain.Applied   `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, t domain.Thread) error {
	_, err := r.collection.InsertOne(ctx, toDoc(t))
	return err
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Thread, error) {
	var d document
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		return domain.Thread{}, err
	}
	return domain.Rehydrate(domain.State{ID: d.ID, Members: d.Members, Provenance: d.Provenance, Revision: d.Revision, Commands: d.Commands})
}
func (r *Repository) Save(ctx context.Context, t domain.Thread, expected uint64, commandID string) error {
	commands := t.Commands()
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": t.ID(), "revision": expected, "provenance": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"provenance": t.Provenance(), "revision": t.Revision()}, "$push": bson.M{"commands": commands[len(commands)-1]}})
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
func toDoc(t domain.Thread) document {
	commands := t.Commands()
	if commands == nil {
		commands = []domain.Applied{}
	}
	return document{ID: t.ID(), Members: t.Members(), Provenance: t.Provenance(), Revision: t.Revision(), Commands: commands}
}
