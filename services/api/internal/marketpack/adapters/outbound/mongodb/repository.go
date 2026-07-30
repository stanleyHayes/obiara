// Package mongodb persists market packs and the configuration audit.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/application"
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) packs() *mongo.Collection {
	return repository.database.Collection("market_packs")
}

func (repository *Repository) changes() *mongo.Collection {
	return repository.database.Collection("configuration_changes")
}

type packDocument struct {
	ID             string          `bson:"_id"`
	Market         string          `bson:"market"`
	TerminologyRef string          `bson:"terminologyRef"`
	Features       map[string]bool `bson:"features"`
	Status         string          `bson:"status"`
	Version        int64           `bson:"version"`
	ProposedBy     string          `bson:"proposedBy"`
	ApprovedBy     string          `bson:"approvedBy,omitempty"`
	CreatedAt      time.Time       `bson:"createdAt"`
	PublishedAt    *time.Time      `bson:"publishedAt,omitempty"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.packs().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "market", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("packs_market_status"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("packs_status_recent"),
		},
	})
	return err
}

func (repository *Repository) CreateWithAudit(ctx context.Context, pack domain.MarketPack, actorID, action string, at time.Time) error {
	return repository.transaction(ctx, func(tx context.Context) error {
		if _, err := repository.packs().InsertOne(tx, toDocument(pack)); err != nil {
			return err
		}
		return repository.append(tx, actorID, action, pack.ID(), at)
	})
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.MarketPack, error) {
	var document packDocument
	if err := repository.packs().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.MarketPack{}, application.ErrPackNotFound
		}
		return domain.MarketPack{}, err
	}
	return toDomain(document), nil
}

func (repository *Repository) UpdateWithAudit(ctx context.Context, pack domain.MarketPack, actorID, action string, at time.Time) error {
	return repository.transaction(ctx, func(tx context.Context) error {
		document := toDocument(pack)
		result, err := repository.packs().UpdateOne(tx,
			bson.M{"_id": document.ID, "version": document.Version - 1},
			bson.M{"$set": bson.M{
				"status": document.Status, "approvedBy": document.ApprovedBy,
				"publishedAt": document.PublishedAt, "version": document.Version,
			}})
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return application.ErrPackNotFound
		}
		return repository.append(tx, actorID, action, pack.ID(), at)
	})
}

func (repository *Repository) ListPublished(ctx context.Context) ([]domain.MarketPack, error) {
	return repository.list(ctx, bson.M{"status": string(domain.StatusPublished)}, 200)
}

func (repository *Repository) ListAll(ctx context.Context, limit int) ([]domain.MarketPack, error) {
	if limit < 1 || limit > 200 {
		limit = 100
	}
	return repository.list(ctx, bson.M{}, limit)
}

func (repository *Repository) list(ctx context.Context, filter bson.M, limit int) ([]domain.MarketPack, error) {
	cursor, err := repository.packs().Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: 1}}).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var packs []domain.MarketPack
	for cursor.Next(ctx) {
		var document packDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		packs = append(packs, toDomain(document))
	}
	return packs, cursor.Err()
}

func (repository *Repository) append(ctx context.Context, actorID, action, packID string, at time.Time) error {
	_, err := repository.changes().InsertOne(ctx, bson.M{
		"actorId": actorID, "action": action, "packId": packID, "at": at.UTC(),
	})
	return err
}

func (repository *Repository) transaction(ctx context.Context, operation func(context.Context) error) error {
	session, err := repository.database.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx context.Context) (any, error) {
		return nil, operation(tx)
	})
	return err
}

// CountAudit supports governance reviews.
func (repository *Repository) CountAudit(ctx context.Context, packID string) (int, error) {
	count, err := repository.changes().CountDocuments(ctx, bson.M{"packId": packID})
	return int(count), err
}

func toDocument(pack domain.MarketPack) packDocument {
	return packDocument{
		ID: pack.ID(), Market: string(pack.Market()), TerminologyRef: pack.TerminologyRef(),
		Features: pack.Features(), Status: string(pack.Status()), Version: pack.Version(),
		ProposedBy: pack.ProposedBy(), ApprovedBy: pack.ApprovedBy(),
		CreatedAt: pack.CreatedAt(), PublishedAt: pack.PublishedAt(),
	}
}

func toDomain(document packDocument) domain.MarketPack {
	return domain.ReconstitutePack(
		document.ID, domain.Market(document.Market), document.TerminologyRef, document.Features,
		domain.PackStatus(document.Status), document.Version, document.ProposedBy, document.ApprovedBy,
		document.CreatedAt, document.PublishedAt,
	)
}
