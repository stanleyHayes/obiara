// Package mongodb logs email deliveries and applies webhook status updates.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
)

type DeliveryLog struct {
	database *mongo.Database
}

func NewDeliveryLog(database *mongo.Database) *DeliveryLog {
	return &DeliveryLog{database: database}
}

func (log *DeliveryLog) collection() *mongo.Collection {
	return log.database.Collection("email_deliveries")
}

type entryDocument struct {
	To          string    `bson:"to"`
	Template    string    `bson:"template"`
	ProviderRef string    `bson:"providerRef"`
	Status      string    `bson:"status"`
	At          time.Time `bson:"at"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}

func (log *DeliveryLog) EnsureIndexes(ctx context.Context) error {
	_, err := log.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "providerRef", Value: 1}},
			Options: options.Index().SetName("email_deliveries_provider"),
		},
		{
			Keys:    bson.D{{Key: "at", Value: 1}},
			Options: options.Index().SetName("email_deliveries_ttl").SetExpireAfterSeconds(90 * 24 * 3600),
		},
	})
	return err
}

func (log *DeliveryLog) Record(ctx context.Context, entry application.DeliveryEntry) error {
	_, err := log.collection().InsertOne(ctx, entryDocument{
		To:          entry.To,
		Template:    string(entry.Template),
		ProviderRef: entry.ProviderRef,
		Status:      entry.Status,
		At:          entry.At.UTC(),
		UpdatedAt:   entry.At.UTC(),
	})
	return err
}

func (log *DeliveryLog) UpdateStatus(ctx context.Context, providerRef, status string, at time.Time) error {
	result, err := log.collection().UpdateOne(ctx,
		bson.M{"providerRef": providerRef},
		bson.M{"$set": bson.M{"status": status, "updatedAt": at.UTC()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrDeliveryNotFound
	}
	return nil
}

// StatusOf reads back a delivery status for tests and metrics.
func (log *DeliveryLog) StatusOf(ctx context.Context, providerRef string) (string, error) {
	var document entryDocument
	if err := log.collection().FindOne(ctx, bson.M{"providerRef": providerRef}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", application.ErrDeliveryNotFound
		}
		return "", err
	}
	return document.Status, nil
}
