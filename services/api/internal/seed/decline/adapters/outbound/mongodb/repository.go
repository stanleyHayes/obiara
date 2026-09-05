package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

type declineDocument struct {
	ID            string               `bson:"_id"`
	DeclinerKey   string               `bson:"declinerKey"`
	SeedKey       string               `bson:"seedKey"`
	RecipientKey  string               `bson:"recipientKey"`
	CommandID     string               `bson:"commandId"`
	Fingerprint   string               `bson:"fingerprint"`
	DeclinedAt    time.Time            `bson:"declinedAt"`
	ExcludedUntil time.Time            `bson:"excludedUntil"`
	Audit         auditDocument        `bson:"audit"`
	Notification  notificationDocument `bson:"notification"`
}

type auditDocument struct {
	CommandID   string    `bson:"commandId"`
	DeclineID   string    `bson:"declineId"`
	DeclinerKey string    `bson:"declinerKey"`
	SeedKey     string    `bson:"seedKey"`
	Action      string    `bson:"action"`
	Fingerprint string    `bson:"fingerprint"`
	OccurredAt  time.Time `bson:"occurredAt"`
}

type notificationDocument struct {
	EventKey     string    `bson:"eventKey"`
	RecipientKey string    `bson:"recipientKey"`
	Kind         string    `bson:"kind"`
	OccurredAt   time.Time `bson:"occurredAt"`
}

func (repository *Repository) declines() *mongo.Collection {
	return repository.database.Collection("seed_declines")
}
func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.declines().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetName("seed_decline_command_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "declinerKey", Value: 1}, {Key: "seedKey", Value: 1}, {Key: "excludedUntil", Value: -1}}, Options: options.Index().SetName("seed_decline_exclusion")},
		{Keys: bson.D{{Key: "declinerKey", Value: 1}, {Key: "recipientKey", Value: 1}, {Key: "excludedUntil", Value: -1}}, Options: options.Index().SetName("seed_decline_pair_exclusion")},
		{Keys: bson.D{{Key: "notification.eventKey", Value: 1}}, Options: options.Index().SetName("seed_decline_notification_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "notification.recipientKey", Value: 1}, {Key: "notification.occurredAt", Value: 1}}, Options: options.Index().SetName("seed_decline_notification_delivery")},
	})
	return err
}

func (repository *Repository) Record(ctx context.Context, decline domain.Decline, notification domain.Notification) (domain.Decline, bool, error) {
	document := fromDomain(decline)
	document.Audit = auditDocument{
		CommandID: decline.CommandID(), DeclineID: decline.ID(), DeclinerKey: decline.DeclinerKey(),
		SeedKey: decline.SeedKey(), Action: "declined", Fingerprint: decline.Fingerprint(),
		OccurredAt: decline.DeclinedAt(),
	}
	document.Notification = notificationDocument{
		EventKey: notification.EventKey(), RecipientKey: notification.RecipientKey(),
		Kind: notification.Kind(), OccurredAt: notification.OccurredAt(),
	}
	_, err := repository.declines().InsertOne(ctx, document)
	if err == nil {
		return decline, false, nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return domain.Decline{}, false, err
	}
	var existing declineDocument
	if findErr := repository.declines().FindOne(ctx, bson.M{"commandId": decline.CommandID()}).Decode(&existing); findErr != nil {
		return domain.Decline{}, false, err
	}
	stored, hydrateErr := toDomain(existing)
	if hydrateErr != nil {
		return domain.Decline{}, false, hydrateErr
	}
	if stored.Fingerprint() != decline.Fingerprint() {
		return domain.Decline{}, false, domain.ErrCommandMismatch
	}
	return stored, true, nil
}

func (repository *Repository) IsExcluded(ctx context.Context, declinerKey, seedKey string, at time.Time) (bool, error) {
	err := repository.declines().FindOne(ctx, bson.M{
		"declinerKey":   declinerKey,
		"seedKey":       seedKey,
		"declinedAt":    bson.M{"$lte": at.UTC()},
		"excludedUntil": bson.M{"$gt": at.UTC()},
	}, options.FindOne().SetSort(bson.D{{Key: "excludedUntil", Value: -1}})).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return err == nil, err
}

// IsPairExcluded answers whether the decliner is still shielded from this
// sower. It reads recipientKey, which the decline already stores: that is the
// seed owner, which is to say whoever reached out and was told no.
func (repository *Repository) IsPairExcluded(ctx context.Context, declinerKey, sowerKey string, at time.Time) (bool, error) {
	err := repository.declines().FindOne(ctx, bson.M{
		"declinerKey":   declinerKey,
		"recipientKey":  sowerKey,
		"declinedAt":    bson.M{"$lte": at.UTC()},
		"excludedUntil": bson.M{"$gt": at.UTC()},
	}, options.FindOne().SetSort(bson.D{{Key: "excludedUntil", Value: -1}})).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return err == nil, err
}

func fromDomain(decline domain.Decline) declineDocument {
	return declineDocument{
		ID: decline.ID(), DeclinerKey: decline.DeclinerKey(), SeedKey: decline.SeedKey(),
		RecipientKey: decline.RecipientKey(), CommandID: decline.CommandID(),
		Fingerprint: decline.Fingerprint(), DeclinedAt: decline.DeclinedAt(),
		ExcludedUntil: decline.ExcludedUntil(),
	}
}

func toDomain(document declineDocument) (domain.Decline, error) {
	return domain.Rehydrate(
		document.ID, document.DeclinerKey, document.SeedKey, document.RecipientKey,
		document.CommandID, document.Fingerprint, document.DeclinedAt, document.ExcludedUntil,
	)
}
