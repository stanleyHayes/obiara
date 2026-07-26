package mongodb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/privacy/domain"
)

const exportLifetime = 7 * 24 * time.Hour

// subjectSource is an explicit cross-context allowlist. It prevents a newly
// introduced collection from silently entering exports or erasure.
type subjectSource struct {
	collection string
	field      string
	erase      bool
}

var subjectSources = []subjectSource{
	{collection: "members", field: "_id", erase: true},
	{collection: "profiles", field: "_id", erase: true},
	{collection: "accounts", field: "_id", erase: true},
	{collection: "doorway_questions", field: "_id", erase: true},
	{collection: "photo_vault", field: "memberId", erase: true},
	{collection: "identity_verifications", field: "accountId", erase: true},
	{collection: "host_applications", field: "accountId", erase: true},
	{collection: "media_assets", field: "ownerId", erase: true},
	{collection: "biometric_templates", field: "accountId", erase: true},
	{collection: "voice_blobs", field: "ownerId", erase: true},
}

// ArchiveAssembler builds a bounded, machine-readable Extended JSON archive.
// The archive collection is private infrastructure and expires automatically.
type ArchiveAssembler struct {
	database *mongo.Database
	now      func() time.Time
}

func NewArchiveAssembler(database *mongo.Database, now func() time.Time) *ArchiveAssembler {
	if now == nil {
		now = time.Now
	}
	return &ArchiveAssembler{database: database, now: now}
}

func (assembler *ArchiveAssembler) EnsureIndexes(ctx context.Context) error {
	_, err := assembler.database.Collection("privacy_export_archives").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "accountId", Value: 1}}, Options: options.Index().SetName("privacy_export_account")},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetName("privacy_export_expiry").SetExpireAfterSeconds(0)},
	})
	return err
}

func (assembler *ArchiveAssembler) Assemble(ctx context.Context, requestID, accountID string) (string, error) {
	if requestID == "" || accountID == "" {
		return "", domain.ErrAccountIDRequired
	}
	collection := assembler.database.Collection("privacy_export_archives")
	var existing struct {
		ArchiveRef string `bson:"archiveRef"`
	}
	if err := collection.FindOne(ctx, bson.M{"_id": requestID}).Decode(&existing); err == nil {
		return existing.ArchiveRef, nil
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return "", err
	}

	sections := bson.M{}
	for _, source := range subjectSources {
		cursor, err := assembler.database.Collection(source.collection).Find(ctx, bson.M{source.field: accountID})
		if err != nil {
			return "", fmt.Errorf("read export source %s: %w", source.collection, err)
		}
		var documents []bson.M
		if err := cursor.All(ctx, &documents); err != nil {
			return "", fmt.Errorf("decode export source %s: %w", source.collection, err)
		}
		if len(documents) > 0 {
			sections[source.collection] = documents
		}
	}
	payload, err := bson.MarshalExtJSON(bson.M{
		"schemaVersion": 1,
		"subject":       accountID,
		"sections":      sections,
	}, false, false)
	if err != nil {
		return "", err
	}
	now := assembler.now().UTC()
	archiveRef := "privacy-export:" + requestID
	_, err = collection.InsertOne(ctx, bson.M{
		"_id": requestID, "accountId": accountID, "archiveRef": archiveRef,
		"mediaType": "application/json", "payload": payload,
		"createdAt": now, "expiresAt": now.Add(exportLifetime),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			if findErr := collection.FindOne(ctx, bson.M{"_id": requestID}).Decode(&existing); findErr == nil {
				return existing.ArchiveRef, nil
			}
		}
		return "", err
	}
	return archiveRef, nil
}

// ErasureRunner removes allowlisted subject-owned records and writes a
// pseudonymous, request-scoped audit receipt. Replays return the same receipt.
type ErasureRunner struct {
	database *mongo.Database
	now      func() time.Time
}

func NewErasureRunner(database *mongo.Database, now func() time.Time) *ErasureRunner {
	if now == nil {
		now = time.Now
	}
	return &ErasureRunner{database: database, now: now}
}

func (runner *ErasureRunner) Erase(ctx context.Context, requestID, accountID string) error {
	if requestID == "" || accountID == "" {
		return domain.ErrAccountIDRequired
	}
	audit := runner.database.Collection("privacy_erasure_audit")
	if err := audit.FindOne(ctx, bson.M{"_id": requestID, "status": "completed"}).Err(); err == nil {
		return nil
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if err := runner.database.Collection("legal_holds").FindOne(ctx, bson.M{"_id": accountID, "liftedAt": nil}).Err(); err == nil {
		return domain.ErrLegalHoldActive
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	counts := bson.M{}
	for _, source := range subjectSources {
		if !source.erase {
			continue
		}
		result, err := runner.database.Collection(source.collection).DeleteMany(ctx, bson.M{source.field: accountID})
		if err != nil {
			return fmt.Errorf("erase source %s: %w", source.collection, err)
		}
		counts[source.collection] = result.DeletedCount
	}
	sum := sha256.Sum256([]byte(accountID))
	_, err := audit.UpdateOne(ctx, bson.M{"_id": requestID}, bson.M{"$setOnInsert": bson.M{
		"subjectDigest": hex.EncodeToString(sum[:]), "status": "completed",
		"counts": counts, "completedAt": runner.now().UTC(),
	}}, options.UpdateOne().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		if findErr := audit.FindOne(ctx, bson.M{"_id": requestID, "status": "completed"}).Err(); findErr == nil {
			return nil
		}
	}
	return err
}
