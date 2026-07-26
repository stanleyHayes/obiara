// Package mongodb persists the in-app inbox.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	notificationdomain "github.com/stanleyHayes/obiara/internal/notifications/domain"
	"github.com/stanleyHayes/obiara/internal/notifications/inapp/domain"
)

type Store struct {
	database *mongo.Database
}

func NewStore(database *mongo.Database) *Store {
	return &Store{database: database}
}

func (store *Store) collection() *mongo.Collection {
	return store.database.Collection("inapp_notifications")
}

type document struct {
	ID        string     `bson:"_id"`
	MemberID  string     `bson:"memberId"`
	Category  string     `bson:"category"`
	Reference string     `bson:"reference"`
	CreatedAt time.Time  `bson:"createdAt"`
	ReadAt    *time.Time `bson:"readAt,omitempty"`
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "memberId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("inapp_member_feed"),
		},
		{
			// The inbox shows the recent past; entries expire after 90 days.
			Keys:    bson.D{{Key: "createdAt", Value: 1}},
			Options: options.Index().SetName("inapp_ttl").SetExpireAfterSeconds(90 * 24 * 3600),
		},
	})
	return err
}

func (store *Store) Add(ctx context.Context, notification domain.Notification) error {
	_, err := store.collection().InsertOne(ctx, document{
		ID:        notification.ID(),
		MemberID:  notification.MemberID(),
		Category:  string(notification.Category()),
		Reference: notification.Reference(),
		CreatedAt: notification.CreatedAt(),
	})
	return err
}

func (store *Store) ListForMember(ctx context.Context, memberID string, limit int) ([]domain.Notification, error) {
	if limit < 1 {
		limit = 50
	}
	cursor, err := store.collection().Find(ctx,
		bson.M{"memberId": memberID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []domain.Notification
	for cursor.Next(ctx) {
		var doc document
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		notifications = append(notifications, domain.Reconstitute(
			doc.ID, doc.MemberID, notificationdomain.Category(doc.Category), doc.Reference, doc.CreatedAt, doc.ReadAt))
	}
	return notifications, cursor.Err()
}

func (store *Store) MarkRead(ctx context.Context, id string, now time.Time) error {
	read := now.UTC()
	_, err := store.collection().UpdateOne(ctx,
		bson.M{"_id": id, "readAt": nil},
		bson.M{"$set": bson.M{"readAt": read}})
	return err
}
