// Package mongodb persists embers with the one-per-attendee-per-fire
// unique index (FR-402) and reads fire attendance for co-attendee checks.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) collection() *mongo.Collection {
	return repository.database.Collection("embers")
}

type emberDocument struct {
	ID         string     `bson:"_id"`
	FireID     string     `bson:"fireId"`
	FromID     string     `bson:"fromId"`
	ToID       string     `bson:"toId"`
	Status     string     `bson:"status"`
	ExpiresAt  time.Time  `bson:"expiresAt"`
	Version    int64      `bson:"version"`
	CreatedAt  time.Time  `bson:"createdAt"`
	RedeemedAt *time.Time `bson:"redeemedAt,omitempty"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// FR-402: one ember per attendee per fire.
			Keys:    bson.D{{Key: "fireId", Value: 1}, {Key: "fromId", Value: 1}},
			Options: options.Index().SetName("embers_fire_giver_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "fireId", Value: 1}, {Key: "toId", Value: 1}},
			Options: options.Index().SetName("embers_fire_recipient"),
		},
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, ember domain.Ember) error {
	_, err := repository.collection().InsertOne(ctx, toDocument(ember))
	if apimongo.IsDuplicateKey(err) {
		return application.ErrEmberAlreadyGiven
	}
	return err
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Ember, error) {
	var document emberDocument
	if err := repository.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Ember{}, application.ErrEmberNotFound
		}
		return domain.Ember{}, err
	}
	return toDomain(document), nil
}

func (repository *Repository) FindDirected(ctx context.Context, fireID, fromID, toID string) (domain.Ember, error) {
	var document emberDocument
	if err := repository.collection().FindOne(ctx, bson.M{"fireId": fireID, "fromId": fromID, "toId": toID}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Ember{}, application.ErrEmberNotFound
		}
		return domain.Ember{}, err
	}
	return toDomain(document), nil
}

func (repository *Repository) Update(ctx context.Context, ember domain.Ember) error {
	document := toDocument(ember)
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{"_id": document.ID, "version": document.Version - 1},
		bson.M{"$set": bson.M{"status": document.Status, "redeemedAt": document.RedeemedAt, "version": document.Version}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrEmberNotFound
	}
	return nil
}

// Attended reports whether the member holds a going RSVP for the fire.
func (repository *Repository) Attended(ctx context.Context, fireID, memberID string) (bool, error) {
	count, err := repository.database.Collection("fire_attendance").CountDocuments(ctx,
		bson.M{"_id": fireID + "|" + memberID, "status": "going"})
	return count > 0, err
}

func toDocument(ember domain.Ember) emberDocument {
	return emberDocument{
		ID:         ember.ID(),
		FireID:     ember.FireID(),
		FromID:     ember.FromID(),
		ToID:       ember.ToID(),
		Status:     string(ember.Status()),
		ExpiresAt:  ember.ExpiresAt(),
		Version:    ember.Version(),
		CreatedAt:  ember.CreatedAt(),
		RedeemedAt: ember.RedeemedAt(),
	}
}

func toDomain(document emberDocument) domain.Ember {
	return domain.ReconstituteEmber(
		document.ID, document.FireID, document.FromID, document.ToID,
		domain.EmberStatus(document.Status), document.ExpiresAt,
		document.Version, document.CreatedAt, document.RedeemedAt,
	)
}
