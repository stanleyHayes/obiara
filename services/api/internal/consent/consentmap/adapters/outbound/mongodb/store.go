// Package mongodb persists consent states and receipts.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
)

type Store struct {
	database *mongo.Database
}

func NewStore(database *mongo.Database) *Store {
	return &Store{database: database}
}

func (store *Store) states() *mongo.Collection {
	return store.database.Collection("consent_states")
}

func (store *Store) receipts() *mongo.Collection {
	return store.database.Collection("consent_receipts")
}

func stateKey(memberID string, purpose domain.Purpose) string {
	return memberID + "|" + string(purpose)
}

func (store *Store) Get(ctx context.Context, memberID string, purpose domain.Purpose) (*bool, error) {
	var document struct {
		Enabled bool `bson:"enabled"`
	}
	err := store.states().FindOne(ctx, bson.M{"_id": stateKey(memberID, purpose)}).Decode(&document)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &document.Enabled, nil
}

func (store *Store) Set(ctx context.Context, memberID string, purpose domain.Purpose, enabled bool) error {
	_, err := store.states().UpdateOne(ctx,
		bson.M{"_id": stateKey(memberID, purpose)},
		bson.M{"$set": bson.M{"enabled": enabled, "updatedAt": time.Now().UTC()}},
		options.UpdateOne().SetUpsert(true))
	return err
}

func (store *Store) AllForMember(ctx context.Context, memberID string) (map[domain.Purpose]bool, error) {
	cursor, err := store.states().Find(ctx, bson.M{"_id": bson.M{"$regex": "^" + memberID + "\\|"}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	states := map[domain.Purpose]bool{}
	for cursor.Next(ctx) {
		var document struct {
			ID      string `bson:"_id"`
			Enabled bool   `bson:"enabled"`
		}
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		states[domain.Purpose(document.ID[len(memberID)+1:])] = document.Enabled
	}
	return states, cursor.Err()
}

func (store *Store) Append(ctx context.Context, receipt domain.Receipt) error {
	_, err := store.receipts().InsertOne(ctx, bson.M{
		"_id": receipt.ID, "memberId": receipt.MemberID, "purpose": string(receipt.Purpose),
		"enabled": receipt.Enabled, "createdAt": receipt.CreatedAt.UTC(),
	})
	return err
}

// EnsureIndexes declares the consent-map indexes.
func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.receipts().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "memberId", Value: 1}, {Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("receipts_member_time"),
	})
	return err
}
