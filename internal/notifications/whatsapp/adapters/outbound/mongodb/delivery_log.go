// Package mongodb logs WhatsApp deliveries for observability (E13-S08
// groundwork: deliverability, fallback and opt-out metrics).
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/application"
)

type DeliveryLog struct {
	database *mongo.Database
}

func NewDeliveryLog(database *mongo.Database) *DeliveryLog {
	return &DeliveryLog{database: database}
}

func (log *DeliveryLog) collection() *mongo.Collection {
	return log.database.Collection("whatsapp_deliveries")
}

type entryDocument struct {
	To          string    `bson:"to"`
	Template    string    `bson:"template"`
	ProviderRef string    `bson:"providerRef,omitempty"`
	Status      string    `bson:"status"`
	At          time.Time `bson:"at"`
}

func (log *DeliveryLog) EnsureIndexes(ctx context.Context) error {
	_, err := log.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "at", Value: -1}},
			Options: options.Index().SetName("whatsapp_deliveries_status"),
		},
		{
			// Delivery logs are operational observability, retained 90 days
			// like pseudonymized analytics (Doc 09 retention posture).
			Keys:    bson.D{{Key: "at", Value: 1}},
			Options: options.Index().SetName("whatsapp_deliveries_ttl").SetExpireAfterSeconds(90 * 24 * 3600),
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
	})
	return err
}

// CountByStatus supports deliverability metrics.
func (log *DeliveryLog) CountByStatus(ctx context.Context, status string) (int, error) {
	count, err := log.collection().CountDocuments(ctx, bson.M{"status": status})
	return int(count), err
}
