// Package mongodb implements the privacy read-side adapters: the
// cross-context export assembler and the erasure runner. Export reads are
// read-only projections over other contexts' collections, used only for
// FR-106 archives; secrets and token hashes are always stripped.
package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// exportLifetime bounds archive retention; archives are delivery
// artifacts, not long-term stores (TTL follow-up to S4-005).
const exportLifetime = 7 * 24 * time.Hour

// ArchiveAssembler builds the machine-readable archive (FR-106: complete,
// within 72 h) and stores it in privacy_export_archives for delivery.
// Assembly is replay-safe: a request always yields the same archive
// reference.
type ArchiveAssembler struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewArchiveAssembler(database *mongo.Database, clock func() time.Time) *ArchiveAssembler {
	return &ArchiveAssembler{database: database, clock: clock}
}

func (assembler *ArchiveAssembler) collection() *mongo.Collection {
	return assembler.database.Collection("privacy_export_archives")
}

// EnsureIndexes declares the archive TTL index.
func (assembler *ArchiveAssembler) EnsureIndexes(ctx context.Context) error {
	_, err := assembler.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "generatedAt", Value: 1}},
		Options: options.Index().SetName("privacy_export_archives_ttl").SetExpireAfterSeconds(int32(exportLifetime.Seconds())),
	})
	return err
}

// exportSection maps one context collection to its archive section. Only
// the requesting account's documents are collected.
type exportSection struct {
	collection string
	section    string
	field      string
	// strip lists fields never exported (token hashes, provider secrets).
	strip []string
}

var exportSections = []exportSection{
	{collection: "members", section: "member", field: "_id"},
	{collection: "accounts", section: "account", field: "_id"},
	{collection: "profiles", section: "profile", field: "_id"},
	{collection: "doorway_questions", section: "doorwayQuestion", field: "_id"},
	{collection: "photo_vault", section: "photoVault", field: "memberId"},
	{collection: "media_assets", section: "mediaAssets", field: "ownerId"},
	{collection: "identity_verifications", section: "verification", field: "accountId"},
	{collection: "consent_records", section: "consent", field: "accountId"},
	{collection: "privacy_requests", section: "privacyRequests", field: "accountId"},
	{collection: "playback_records", section: "listening", field: "listenerId"},
	{collection: "sessions", section: "sessions", field: "memberId", strip: []string{"accessTokenHash", "refreshTokenHash", "replacedRefreshHash"}},
}

type archiveDocument struct {
	ID          string    `bson:"_id"`
	AccountID   string    `bson:"accountId"`
	ArchiveRef  string    `bson:"archiveRef"`
	Payload     []byte    `bson:"payload"`
	GeneratedAt time.Time `bson:"generatedAt"`
}

// Assemble collects the account's documents and stores the archive. When
// the request already has an archive, its reference is returned unchanged
// (replay safety).
func (assembler *ArchiveAssembler) Assemble(ctx context.Context, requestID, accountID string) (string, error) {
	var existing archiveDocument
	if err := assembler.collection().FindOne(ctx, bson.M{"_id": requestID}).Decode(&existing); err == nil {
		return existing.ArchiveRef, nil
	} else if err != mongo.ErrNoDocuments {
		return "", err
	}

	sections := make(map[string][]bson.M, len(exportSections))
	for _, spec := range exportSections {
		documents, err := assembler.collect(ctx, spec, accountID)
		if err != nil {
			return "", fmt.Errorf("collect %s: %w", spec.collection, err)
		}
		sections[spec.section] = documents
	}
	payload, err := json.Marshal(sections)
	if err != nil {
		return "", fmt.Errorf("encode archive payload: %w", err)
	}

	ref := "privacy-export:" + requestID
	_, err = assembler.collection().InsertOne(ctx, archiveDocument{
		ID:          requestID,
		AccountID:   accountID,
		ArchiveRef:  ref,
		Payload:     payload,
		GeneratedAt: assembler.clock().UTC(),
	})
	if err != nil {
		return "", err
	}
	return ref, nil
}

func (assembler *ArchiveAssembler) collect(ctx context.Context, spec exportSection, accountID string) ([]bson.M, error) {
	cursor, err := assembler.database.Collection(spec.collection).Find(ctx, bson.M{spec.field: accountID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []bson.M
	for cursor.Next(ctx) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		for _, field := range spec.strip {
			delete(document, field)
		}
		documents = append(documents, document)
	}
	return documents, cursor.Err()
}
