// Package mongodb persists device push registrations.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/internal/notifications/push/domain"
)

type Registry struct {
	database *mongo.Database
	now      func() time.Time
}

func NewRegistry(database *mongo.Database, now func() time.Time) *Registry {
	return &Registry{database: database, now: now}
}

func (registry *Registry) collection() *mongo.Collection {
	return registry.database.Collection("push_tokens")
}

type document struct {
	Token      string    `bson:"_id"`
	MemberID   string    `bson:"memberId"`
	Platform   string    `bson:"platform"`
	Registered time.Time `bson:"registeredAt"`
}

// EnsureIndexes declares the lookup index and an inactivity expiry. A handset
// that has not checked in for six months is almost certainly gone, and an
// unbounded token set eventually poisons whole provider batches.
func (registry *Registry) EnsureIndexes(ctx context.Context) error {
	_, err := registry.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "memberId", Value: 1}},
			Options: options.Index().SetName("push_tokens_member"),
		},
		{
			Keys: bson.D{{Key: "registeredAt", Value: 1}},
			Options: options.Index().SetName("push_tokens_stale").
				SetExpireAfterSeconds(int32((180 * 24 * time.Hour).Seconds())),
		},
	})
	return err
}

// Register upserts by token, so a handset that signs in as a different member
// moves rather than duplicating — which would otherwise deliver one member's
// notifications to another's device.
func (registry *Registry) Register(ctx context.Context, registration domain.Registration) error {
	_, err := registry.collection().UpdateOne(ctx,
		bson.M{"_id": registration.Token()},
		bson.M{"$set": bson.M{
			"memberId":     registration.MemberID(),
			"platform":     string(registration.Platform()),
			"registeredAt": registration.Registered(),
		}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (registry *Registry) TokensFor(ctx context.Context, memberID string) ([]string, error) {
	cursor, err := registry.collection().Find(ctx, bson.M{"memberId": memberID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []string
	for cursor.Next(ctx) {
		var stored document
		if err := cursor.Decode(&stored); err != nil {
			return nil, err
		}
		// A token that no longer parses cannot be delivered to and would
		// risk the batch it rides in.
		if domain.ValidToken(stored.Token) {
			tokens = append(tokens, stored.Token)
		}
	}
	return tokens, cursor.Err()
}

func (registry *Registry) Forget(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}
	_, err := registry.collection().DeleteMany(ctx, bson.M{"_id": bson.M{"$in": tokens}})
	return err
}

func (registry *Registry) ForgetMember(ctx context.Context, memberID string) error {
	_, err := registry.collection().DeleteMany(ctx, bson.M{"memberId": memberID})
	return err
}
