// Package mongodb persists immutable media asset metadata.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

type AssetRepository struct {
	assets *mongo.Collection
}

func NewAssetRepository(database *mongo.Database) *AssetRepository {
	return &AssetRepository{assets: database.Collection("media_assets")}
}

type assetDocument struct {
	ID                string    `bson:"_id"`
	ObjectKey         string    `bson:"objectKey"`
	OwnerID           string    `bson:"ownerId"`
	ContentType       string    `bson:"contentType"`
	Size              int64     `bson:"size"`
	ChecksumAlgorithm string    `bson:"checksumAlgorithm"`
	ChecksumValue     string    `bson:"checksumValue"`
	DurationNanos     int64     `bson:"durationNanos"`
	CreatedAt         time.Time `bson:"createdAt"`
	ExpiresAt         time.Time `bson:"expiresAt,omitempty"`
	RetentionUntil    time.Time `bson:"retentionUntil,omitempty"`
	RetentionLegal    bool      `bson:"retentionLegalHold"`
	DeletedAt         time.Time `bson:"deletedAt,omitempty"`
}

func (repository *AssetRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.assets.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "objectKey", Value: 1}},
			Options: options.Index().SetName("media_asset_object_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "ownerId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("media_asset_owner"),
		},
	})
	return err
}

// Register records an asset before its upload is authorized, because the
// access service authorizes against a known owner and object key. An asset
// whose bytes never arrive keeps a placeholder size and is swept by expiry.
func (repository *AssetRepository) Register(ctx context.Context, asset domain.Asset) error {
	_, err := repository.assets.InsertOne(ctx, toAssetDocument(asset))
	return err
}

func (repository *AssetRepository) FindByID(ctx context.Context, id string) (domain.Asset, error) {
	var document assetDocument
	if err := repository.assets.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Asset{}, domain.ErrAssetUnavailable
		}
		return domain.Asset{}, domain.ErrAssetUnavailable
	}
	return fromAssetDocument(document)
}

// Complete records what storage actually accepted. Size, checksum and
// duration are written here and nowhere else: they are the server's own
// account of the bytes, and every gate downstream counts against them.
func (repository *AssetRepository) Complete(
	ctx context.Context,
	id string,
	size int64,
	duration time.Duration,
	checksum domain.Checksum,
) error {
	result, err := repository.assets.UpdateOne(ctx,
		bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{
			"size":              size,
			"durationNanos":     int64(duration),
			"checksumAlgorithm": checksum.Algorithm(),
			"checksumValue":     checksum.Value(),
		}})
	if err != nil || result.MatchedCount != 1 {
		return domain.ErrAssetUnavailable
	}
	return nil
}

// Delete erases the row. Retention and legal hold are enforced by the domain
// before this is reached.
func (repository *AssetRepository) Delete(ctx context.Context, id string) error {
	_, err := repository.assets.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deletedAt": time.Now().UTC()}})
	return err
}

func toAssetDocument(asset domain.Asset) assetDocument {
	return assetDocument{
		ID: asset.ID(), ObjectKey: asset.ObjectKey(), OwnerID: asset.OwnerID(),
		ContentType: asset.ContentType(), Size: asset.Size(),
		ChecksumAlgorithm: asset.Checksum().Algorithm(),
		ChecksumValue:     asset.Checksum().Value(),
		DurationNanos:     int64(asset.Duration()),
		CreatedAt:         asset.CreatedAt(), ExpiresAt: asset.ExpiresAt(),
		RetentionUntil: asset.Retention().Until(),
		RetentionLegal: asset.Retention().LegalHold(),
		DeletedAt:      asset.DeletedAt(),
	}
}

func fromAssetDocument(document assetDocument) (domain.Asset, error) {
	checksum, err := domain.NewChecksum(document.ChecksumAlgorithm, document.ChecksumValue)
	if err != nil {
		return domain.Asset{}, domain.ErrAssetUnavailable
	}
	asset, err := domain.NewAsset(domain.NewAssetParams{
		ID: document.ID, ObjectKey: document.ObjectKey, OwnerID: document.OwnerID,
		ContentType: document.ContentType, Size: document.Size, Checksum: checksum,
		Duration:  time.Duration(document.DurationNanos),
		CreatedAt: document.CreatedAt, ExpiresAt: document.ExpiresAt,
		Retention: domain.NewRetention(document.RetentionUntil, document.RetentionLegal),
	})
	if err != nil {
		return domain.Asset{}, domain.ErrAssetUnavailable
	}
	if !document.DeletedAt.IsZero() {
		return asset.MarkDeleted(document.DeletedAt)
	}
	return asset, nil
}
