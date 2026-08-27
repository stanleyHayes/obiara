package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NotificationMarks records when each operator last acknowledged their
// inbox. One row per principal, holding one timestamp: the inbox itself is
// derived from live queues, so there is no per-notification state to keep.
type NotificationMarks struct {
	database *mongo.Database
}

func NewNotificationMarks(database *mongo.Database) *NotificationMarks {
	return &NotificationMarks{database: database}
}

func (marks *NotificationMarks) collection() *mongo.Collection {
	return marks.database.Collection("admin_notification_marks")
}

type notificationMarkDocument struct {
	PrincipalID string    `bson:"_id"`
	SeenAt      time.Time `bson:"seenAt"`
}

// SeenAt returns the zero time for an operator who has never acknowledged
// the inbox, which reads as "everything is unread" rather than as an error.
func (marks *NotificationMarks) SeenAt(ctx context.Context, principalID string) (time.Time, error) {
	var document notificationMarkDocument
	err := marks.collection().FindOne(ctx, bson.M{"_id": principalID}).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return document.SeenAt, nil
}

// MarkSeen moves the watermark forward. It never moves backwards: a stale
// request arriving late must not resurrect notifications the operator has
// already acknowledged.
func (marks *NotificationMarks) MarkSeen(ctx context.Context, principalID string, at time.Time) error {
	_, err := marks.collection().UpdateOne(ctx,
		bson.M{"_id": principalID},
		bson.M{"$max": bson.M{"seenAt": at.UTC()}},
		options.UpdateOne().SetUpsert(true))
	return err
}
