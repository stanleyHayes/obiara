// Package mongodb is the outbound session persistence adapter. Documents
// hold token hashes only; plaintext tokens never reach this package
// (data-classification C3/C4 handling).
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	apimongo "github.com/stanleyHayes/obiara/services/api/internal/platform/mongo"
)

var ErrSessionNotFound = errors.New("session not found")

type Repository struct {
	collection *mongo.Collection
}

type sessionDocument struct {
	ID                  string    `bson:"_id"`
	MemberID            string    `bson:"memberId"`
	DeviceID            string    `bson:"deviceId"`
	Status              string    `bson:"status"`
	AccessTokenHash     string    `bson:"accessTokenHash"`
	AccessExpiresAt     time.Time `bson:"accessExpiresAt"`
	RefreshTokenHash    string    `bson:"refreshTokenHash"`
	ReplacedRefreshHash string    `bson:"replacedRefreshHash,omitempty"`
	RefreshExpiresAt    time.Time `bson:"refreshExpiresAt"`
	Version             int64     `bson:"version"`
	CreatedAt           time.Time `bson:"createdAt"`
	UpdatedAt           time.Time `bson:"updatedAt"`
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("sessions")}
}

// EnsureIndexes declares the session indexes. Wrapped as a migration when
// the migration runner is wired into module startup (platform/mongo README).
func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "refreshTokenHash", Value: 1}},
			Options: options.Index().SetName("sessions_refresh_hash_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "memberId", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("sessions_member_status"),
		},
		{
			Keys:    bson.D{{Key: "deviceId", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("sessions_device_status"),
		},
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, session domain.Session) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(session))
	if apimongo.IsDuplicateKey(err) {
		return application.ErrStaleSession
	}
	return err
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Session, error) {
	var document sessionDocument
	if err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Session{}, ErrSessionNotFound
		}
		return domain.Session{}, err
	}
	return toDomain(document), nil
}

// Update applies optimistic concurrency: the filter pins the version the
// caller read, and a zero match means a racing writer won (agent_plan.md
// §7.4). The stored version increments on every successful update.
func (repository *Repository) Update(ctx context.Context, session domain.Session) error {
	document := toDocument(session)
	result, err := repository.collection.UpdateOne(ctx,
		bson.M{"_id": document.ID, "version": document.Version},
		bson.A{
			bson.M{"$set": bson.M{
				"status":              document.Status,
				"accessTokenHash":     document.AccessTokenHash,
				"accessExpiresAt":     document.AccessExpiresAt,
				"refreshTokenHash":    document.RefreshTokenHash,
				"replacedRefreshHash": document.ReplacedRefreshHash,
				"refreshExpiresAt":    document.RefreshExpiresAt,
				"updatedAt":           document.UpdatedAt,
			}},
			bson.M{"$set": bson.M{"version": document.Version + 1}},
		})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrStaleSession
	}
	return nil
}

func (repository *Repository) RevokeAllForMember(ctx context.Context, memberID string, now time.Time) error {
	return repository.revokeAll(ctx, bson.M{"memberId": memberID, "status": string(domain.StatusActive)}, now)
}

func (repository *Repository) RevokeAllForDevice(ctx context.Context, deviceID string, now time.Time) error {
	return repository.revokeAll(ctx, bson.M{"deviceId": deviceID, "status": string(domain.StatusActive)}, now)
}

func (repository *Repository) revokeAll(ctx context.Context, filter bson.M, now time.Time) error {
	_, err := repository.collection.UpdateMany(ctx, filter, bson.A{
		bson.M{"$set": bson.M{"status": string(domain.StatusRevoked), "updatedAt": now.UTC()}},
		bson.M{"$set": bson.M{"version": bson.M{"$add": bson.A{"$version", 1}}}},
	})
	return err
}

func toDocument(session domain.Session) sessionDocument {
	return sessionDocument{
		ID:                  session.ID(),
		MemberID:            session.MemberID(),
		DeviceID:            session.DeviceID(),
		Status:              string(session.Status()),
		AccessTokenHash:     session.AccessTokenHash(),
		AccessExpiresAt:     session.AccessExpiresAt(),
		RefreshTokenHash:    session.RefreshTokenHash(),
		ReplacedRefreshHash: session.ReplacedRefreshHash(),
		RefreshExpiresAt:    session.RefreshExpiresAt(),
		Version:             session.Version(),
		CreatedAt:           session.CreatedAt(),
		UpdatedAt:           session.UpdatedAt(),
	}
}

func toDomain(document sessionDocument) domain.Session {
	return domain.Reconstitute(
		document.ID,
		document.MemberID,
		document.DeviceID,
		domain.Status(document.Status),
		document.AccessTokenHash,
		document.AccessExpiresAt,
		document.RefreshTokenHash,
		document.ReplacedRefreshHash,
		document.RefreshExpiresAt,
		document.Version,
		document.CreatedAt,
		document.UpdatedAt,
	)
}
