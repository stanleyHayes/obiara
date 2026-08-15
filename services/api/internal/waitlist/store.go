package waitlist

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

var ErrInvalidEntry = errors.New("invalid waitlist entry")

type Entry struct {
	ID                bson.ObjectID `bson:"_id,omitempty" json:"-"`
	Name              string        `bson:"name" json:"name"`
	Email             string        `bson:"email" json:"email"`
	ConsentVersion    string        `bson:"consentVersion" json:"consentVersion"`
	SignedUpAt        time.Time     `bson:"signedUpAt" json:"signedUpAt"`
	NotificationState string        `bson:"notificationState" json:"notificationState"`
	NotifiedAt        *time.Time    `bson:"notifiedAt,omitempty" json:"notifiedAt,omitempty"`
}

type Store struct {
	collection *mongo.Collection
	now        func() time.Time
}

func NewStore(database *mongo.Database, now func() time.Time) *Store {
	return &Store{collection: database.Collection("launch_waitlist"), now: now}
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetName("launch_waitlist_email_unique").SetUnique(true),
	})
	return err
}

// Join is intentionally idempotent by normalized email. A repeat submission
// confirms the existing place without changing the original consent evidence.
func (store *Store) Join(ctx context.Context, name, email, consentVersion string) (Entry, bool, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	consentVersion = strings.TrimSpace(consentVersion)
	if name == "" || email == "" || consentVersion == "" {
		return Entry{}, false, ErrInvalidEntry
	}
	entry := Entry{
		ID: bson.NewObjectID(), Name: name, Email: email, ConsentVersion: consentVersion,
		SignedUpAt: store.now().UTC(), NotificationState: "pending",
	}
	result, err := store.collection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$setOnInsert": entry}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		if apimongo.IsDuplicateKey(err) {
			// A concurrent first-time submission of the same email won the
			// insert; join stays idempotent by returning that record.
			var existing Entry
			if findErr := store.collection.FindOne(ctx, bson.M{"email": email}).Decode(&existing); findErr != nil {
				return Entry{}, false, findErr
			}
			return existing, false, nil
		}
		return Entry{}, false, err
	}
	created := result.UpsertedCount == 1
	if !created {
		if err = store.collection.FindOne(ctx, bson.M{"email": email}).Decode(&entry); err != nil {
			return Entry{}, false, err
		}
	}
	return entry, created, nil
}

func (store *Store) List(ctx context.Context, limit int) ([]Entry, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}
	cursor, err := store.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "signedUpAt", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	entries := make([]Entry, 0)
	if err = cursor.All(ctx, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
