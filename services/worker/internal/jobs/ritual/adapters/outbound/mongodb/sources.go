// Package mongodb provides the worker-side read models for ritual
// dispatch: active members, preference time zones, and fire herald
// windows. Read-only projections over the identity, notification and fire
// contexts' collections.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/internal/notifications/ritual/application"
)

type Sources struct {
	database *mongo.Database
}

func NewSources(database *mongo.Database) *Sources {
	return &Sources{database: database}
}

// ListActiveIDs returns active account IDs (identity context read model).
func (sources *Sources) ListActiveIDs(ctx context.Context, limit int) ([]string, error) {
	if limit < 1 {
		limit = 10000
	}
	cursor, err := sources.database.Collection("accounts").Find(ctx,
		bson.M{"status": "active"},
		options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []string
	for cursor.Next(ctx) {
		var document struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		ids = append(ids, document.ID)
	}
	return ids, cursor.Err()
}

// TimezoneFor reads the member's preference time zone, defaulting to
// Africa/Accra when no preferences exist yet (E13-S01 defaults).
func (sources *Sources) TimezoneFor(ctx context.Context, memberID string) (string, error) {
	var document struct {
		Timezone string `bson:"timezone"`
	}
	err := sources.database.Collection("notification_preferences").FindOne(ctx,
		bson.M{"_id": memberID},
		options.FindOne().SetProjection(bson.M{"timezone": 1}),
	).Decode(&document)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "Africa/Accra", nil
		}
		return "", err
	}
	return document.Timezone, nil
}

// StartingWithin lists scheduled fires starting in the window with their
// going attendees (fire context read model).
func (sources *Sources) StartingWithin(ctx context.Context, from, until time.Time) ([]application.FireWindow, error) {
	cursor, err := sources.database.Collection("fires").Find(ctx,
		bson.M{"status": "scheduled", "startsAt": bson.M{"$gte": from.UTC(), "$lte": until.UTC()}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var windows []application.FireWindow
	for cursor.Next(ctx) {
		var fire struct {
			ID       string    `bson:"_id"`
			StartsAt time.Time `bson:"startsAt"`
		}
		if err := cursor.Decode(&fire); err != nil {
			return nil, err
		}
		attendees, err := sources.goingAttendees(ctx, fire.ID)
		if err != nil {
			return nil, err
		}
		windows = append(windows, application.FireWindow{FireID: fire.ID, StartsAt: fire.StartsAt, Attendees: attendees})
	}
	return windows, cursor.Err()
}

func (sources *Sources) goingAttendees(ctx context.Context, fireID string) ([]string, error) {
	cursor, err := sources.database.Collection("fire_attendance").Find(ctx,
		bson.M{"fireId": fireID, "status": "going"},
		options.Find().SetProjection(bson.M{"memberId": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var attendees []string
	for cursor.Next(ctx) {
		var document struct {
			MemberID string `bson:"memberId"`
		}
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		attendees = append(attendees, document.MemberID)
	}
	return attendees, cursor.Err()
}
