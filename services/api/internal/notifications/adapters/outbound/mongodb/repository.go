// Package mongodb persists notification preferences and daily delivery
// counters. The cap claim is atomic (conditional increment, like fire
// seat claims) so concurrent sends cannot overshoot the six-per-day cap.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/notifications/application"
	"github.com/stanleyHayes/obiara/services/api/internal/notifications/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) preferences() *mongo.Collection {
	return repository.database.Collection("notification_preferences")
}

func (repository *Repository) deliveries() *mongo.Collection {
	return repository.database.Collection("notification_deliveries")
}

type preferencesDocument struct {
	MemberID   string          `bson:"_id"`
	Muted      map[string]bool `bson:"muted"`
	QuietStart int             `bson:"quietStart"`
	QuietEnd   int             `bson:"quietEnd"`
	Timezone   string          `bson:"timezone"`
	Version    int64           `bson:"version"`
	UpdatedAt  time.Time       `bson:"updatedAt"`
}

type counterDocument struct {
	ID        string    `bson:"_id"`
	MemberID  string    `bson:"memberId"`
	Date      string    `bson:"date"`
	Count     int       `bson:"count"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.deliveries().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "memberId", Value: 1}, {Key: "date", Value: 1}},
			Options: options.Index().SetName("deliveries_member_date_unique").SetUnique(true),
		},
		{
			// Daily counters are operational, not product truth; expire
			// them after 7 days.
			Keys:    bson.D{{Key: "updatedAt", Value: 1}},
			Options: options.Index().SetName("deliveries_ttl").SetExpireAfterSeconds(7 * 24 * 3600),
		},
	})
	return err
}

func (repository *Repository) Upsert(ctx context.Context, preferences domain.Preferences) error {
	muted := make(map[string]bool, len(preferences.Muted()))
	for category, value := range preferences.Muted() {
		muted[string(category)] = value
	}
	_, err := repository.preferences().ReplaceOne(ctx,
		bson.M{"_id": preferences.MemberID()},
		preferencesDocument{
			MemberID:   preferences.MemberID(),
			Muted:      muted,
			QuietStart: preferences.QuietStart(),
			QuietEnd:   preferences.QuietEnd(),
			Timezone:   preferences.Timezone(),
			Version:    preferences.Version(),
			UpdatedAt:  preferences.UpdatedAt(),
		},
		options.Replace().SetUpsert(true))
	return err
}

func (repository *Repository) FindByMember(ctx context.Context, memberID string) (domain.Preferences, error) {
	var document preferencesDocument
	if err := repository.preferences().FindOne(ctx, bson.M{"_id": memberID}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Preferences{}, application.ErrPreferencesNotFound
		}
		return domain.Preferences{}, err
	}
	muted := make(map[domain.Category]bool, len(document.Muted))
	for category, value := range document.Muted {
		muted[domain.Category(category)] = value
	}
	return domain.ReconstitutePreferences(document.MemberID, muted, document.QuietStart, document.QuietEnd, document.Timezone, document.Version, document.UpdatedAt), nil
}

// ClaimSlot atomically increments the day's counter when under the cap.
// The conditional update matches only documents still below the cap; the
// upsert path handles the first delivery of the day. A document already at
// the cap collides on the upsert insert, which is the cap-reached signal.
func (repository *Repository) ClaimSlot(ctx context.Context, memberID, localDate string, cap int) (bool, error) {
	key := memberID + "|" + localDate
	claimed, err := repository.deliveries().UpdateOne(ctx,
		bson.M{"_id": key, "count": bson.M{"$lt": cap}},
		bson.M{"$inc": bson.M{"count": 1}, "$set": bson.M{"updatedAt": time.Now().UTC()}},
		options.UpdateOne().SetUpsert(true))
	if apimongo.IsDuplicateKey(err) {
		return false, nil // counter exists at the cap
	}
	if err != nil {
		return false, err
	}
	if claimed.UpsertedCount == 1 || claimed.ModifiedCount == 1 {
		return true, nil
	}
	return false, nil
}
