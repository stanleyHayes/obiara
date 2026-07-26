package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("commerce_catalog_skus")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "skuKey", Value: 1}, {Key: "version", Value: 1}}, Options: options.Index().SetUnique(true).SetName("catalog_sku_version_unique")}, {Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("catalog_command_unique")}, {Keys: bson.D{{Key: "skuKey", Value: 1}, {Key: "status", Value: 1}, {Key: "version", Value: -1}}, Options: options.Index().SetName("catalog_published_lookup")}})
	return e
}

type doc struct {
	ID          string           `bson:"_id"`
	SKUKey      string           `bson:"skuKey"`
	TitleRef    string           `bson:"titleRef"`
	Version     uint64           `bson:"version"`
	Kind        domain.Kind      `bson:"kind"`
	Price       domain.Price     `bson:"price"`
	Status      domain.Status    `bson:"status"`
	PublishedAt time.Time        `bson:"publishedAt,omitempty"`
	RetiredAt   time.Time        `bson:"retiredAt,omitempty"`
	Revision    uint64           `bson:"revision"`
	Events      []domain.Event   `bson:"events"`
	Commands    []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, s domain.SKU) error {
	_, e := r.c.InsertOne(ctx, toDoc(s))
	return r.dupe(ctx, s.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, sku string, version uint64) (domain.SKU, error) {
	return r.find(ctx, bson.M{"skuKey": sku, "version": version}, nil)
}
func (r *Repository) FindLatest(ctx context.Context, sku string) (domain.SKU, error) {
	return r.find(ctx, bson.M{"skuKey": sku}, options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}}))
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.SKU, error) {
	return r.find(ctx, bson.M{"commands.id": id}, nil)
}
func (r *Repository) find(ctx context.Context, f bson.M, o *options.FindOneOptionsBuilder) (domain.SKU, error) {
	var d doc
	var e error
	if o == nil {
		e = r.c.FindOne(ctx, f).Decode(&d)
	} else {
		e = r.c.FindOne(ctx, f, o).Decode(&d)
	}
	if e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.SKU{}, application.ErrNotFound
		}
		return domain.SKU{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, SKUKey: d.SKUKey, TitleRef: d.TitleRef, Version: d.Version, Kind: d.Kind, Price: d.Price, Status: d.Status, PublishedAt: d.PublishedAt, RetiredAt: d.RetiredAt, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, s domain.SKU, expected uint64, command string) error {
	es, cs := s.Events(), s.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	res, e := r.c.UpdateOne(ctx, bson.M{"_id": s.ID(), "revision": expected, "price": s.Price()}, bson.M{"$set": bson.M{"status": s.Status(), "publishedAt": s.PublishedAt(), "retiredAt": s.RetiredAt(), "revision": s.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
	if e != nil {
		return r.dupe(ctx, command, e)
	}
	if res.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) dupe(ctx context.Context, command string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func toDoc(s domain.SKU) doc {
	return doc{ID: s.ID(), SKUKey: s.SKUKey(), TitleRef: s.TitleRef(), Version: s.Version(), Kind: s.Kind(), Price: s.Price(), Status: s.Status(), PublishedAt: s.PublishedAt(), RetiredAt: s.RetiredAt(), Revision: s.Revision(), Events: s.Events(), Commands: s.Commands()}
}
