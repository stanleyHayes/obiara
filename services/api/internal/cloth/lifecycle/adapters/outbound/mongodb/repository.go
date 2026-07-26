package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/lifecycle/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("cloth_lifecycles")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true)})
	return e
}

type document struct {
	ID         string            `bson:"_id"`
	Members    []string          `bson:"members"`
	Status     domain.Status     `bson:"status"`
	Provenance domain.Provenance `bson:"provenance"`
	ArchiveRef string            `bson:"archiveRef,omitempty"`
	ArchivedAt time.Time         `bson:"archivedAt,omitempty"`
	Tombstone  *domain.Tombstone `bson:"tombstone,omitempty"`
	Revision   uint64            `bson:"revision"`
	Commands   []domain.Applied  `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, v domain.Lifecycle) error {
	_, e := r.c.InsertOne(ctx, toDoc(v))
	return e
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Lifecycle, error) {
	var d document
	if e := r.c.FindOne(ctx, bson.M{"_id": id}).Decode(&d); e != nil {
		return domain.Lifecycle{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, Members: d.Members, Status: d.Status, Provenance: d.Provenance, ArchiveRef: d.ArchiveRef, ArchivedAt: d.ArchivedAt, Tombstone: d.Tombstone, Revision: d.Revision, Commands: d.Commands})
}
func (r *Repository) Save(ctx context.Context, v domain.Lifecycle, expected uint64, commandID string) error {
	cs := v.Commands()
	res, e := r.c.UpdateOne(ctx, bson.M{"_id": v.ID(), "revision": expected}, bson.M{"$set": bson.M{"status": v.Status(), "archiveRef": v.ArchiveRef(), "archivedAt": v.ArchivedAt(), "tombstone": v.Tombstone(), "revision": v.Revision()}, "$push": bson.M{"commands": cs[len(cs)-1]}})
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
func toDoc(v domain.Lifecycle) document {
	return document{ID: v.ID(), Members: v.Members(), Status: v.Status(), Provenance: v.Provenance(), ArchiveRef: v.ArchiveRef(), ArchivedAt: v.ArchivedAt(), Tombstone: v.Tombstone(), Revision: v.Revision(), Commands: v.Commands()}
}
