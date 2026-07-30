// Package mongodb persists privacy-minimal liveness attempts and temporary
// manual-review artifact references.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	platformmongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
)

type Store struct {
	attempts  *mongo.Collection
	reviews   *mongo.Collection
	artifacts *mongo.Collection
}

type eventDocument struct {
	Status     string    `bson:"status"`
	Reason     string    `bson:"reason"`
	ActorKey   string    `bson:"actorKey"`
	OccurredAt time.Time `bson:"occurredAt"`
	Version    uint64    `bson:"version"`
}

type attemptDocument struct {
	ID          string          `bson:"_id"`
	CommandID   string          `bson:"commandId"`
	SubjectKey  string          `bson:"subjectKey"`
	InputKey    string          `bson:"inputKey"`
	Status      string          `bson:"status"`
	Reason      string          `bson:"reason,omitempty"`
	ProviderRef string          `bson:"providerRef,omitempty"`
	CreatedAt   time.Time       `bson:"createdAt"`
	DecidedAt   time.Time       `bson:"decidedAt,omitempty"`
	Version     uint64          `bson:"version"`
	Events      []eventDocument `bson:"events"`
}

type reviewDocument struct {
	ID               string    `bson:"_id"`
	VoiceArtifactRef string    `bson:"voiceArtifactRef"`
	FaceArtifactRef  string    `bson:"faceArtifactRef"`
	Reason           string    `bson:"reason"`
	CreatedAt        time.Time `bson:"createdAt"`
	ExpiresAt        time.Time `bson:"expiresAt"`
}

func NewStore(database *mongo.Database) *Store {
	return &Store{
		attempts:  database.Collection("liveness_attempts"),
		reviews:   database.Collection("liveness_review_tasks"),
		artifacts: database.Collection("liveness_artifacts"),
	}
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	if _, err := store.attempts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "commandId", Value: 1}},
		Options: options.Index().SetName("liveness_command_unique").SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := store.reviews.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetName("liveness_review_ttl").
			SetExpireAfterSeconds(0),
	}); err != nil {
		return err
	}
	_, err := store.artifacts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetName("liveness_artifact_ttl").
			SetExpireAfterSeconds(0),
	})
	return err
}

func (store *Store) SaveArtifact(ctx context.Context, artifact application.Artifact) error {
	_, err := store.artifacts.InsertOne(ctx, bson.M{
		"_id": artifact.ID, "subjectKey": artifact.SubjectKey, "kind": artifact.Kind,
		"mediaType": artifact.MediaType, "ciphertext": artifact.Ciphertext,
		"nonce": artifact.Nonce, "createdAt": artifact.CreatedAt, "expiresAt": artifact.ExpiresAt,
	})
	if err != nil {
		return application.ErrArtifactStore
	}
	return nil
}

func toDocument(attempt domain.Attempt) attemptDocument {
	events := make([]eventDocument, 0, len(attempt.Events()))
	for _, event := range attempt.Events() {
		events = append(events, eventDocument{
			Status: string(event.Status()), Reason: string(event.Reason()),
			ActorKey: event.ActorKey(), OccurredAt: event.OccurredAt(),
			Version: event.Version(),
		})
	}
	return attemptDocument{
		ID: attempt.ID(), CommandID: attempt.CommandID(),
		SubjectKey: attempt.SubjectKey(), InputKey: attempt.InputKey(),
		Status: string(attempt.Status()), Reason: string(attempt.Reason()),
		ProviderRef: attempt.ProviderRef(), CreatedAt: attempt.CreatedAt(),
		DecidedAt: attempt.DecidedAt(), Version: attempt.Version(), Events: events,
	}
}

func fromDocument(document attemptDocument) (domain.Attempt, error) {
	events := make([]domain.Event, 0, len(document.Events))
	for _, stored := range document.Events {
		event, err := domain.NewEvent(domain.EventParams{
			Status: domain.Status(stored.Status), Reason: domain.Reason(stored.Reason),
			ActorKey: stored.ActorKey, OccurredAt: stored.OccurredAt,
			Version: stored.Version,
		})
		if err != nil {
			return domain.Attempt{}, application.ErrServiceUnavailable
		}
		events = append(events, event)
	}
	attempt, err := domain.Reconstitute(
		document.ID, document.CommandID, document.SubjectKey, document.InputKey,
		domain.Status(document.Status), domain.Reason(document.Reason),
		document.ProviderRef, document.CreatedAt, document.DecidedAt,
		document.Version, events,
	)
	if err != nil {
		return domain.Attempt{}, application.ErrServiceUnavailable
	}
	return attempt, nil
}

func (store *Store) Create(ctx context.Context, attempt domain.Attempt) (domain.Attempt, bool, error) {
	_, err := store.attempts.InsertOne(ctx, toDocument(attempt))
	if err == nil {
		return attempt, false, nil
	}
	if !platformmongo.IsDuplicateKey(err) {
		return domain.Attempt{}, false, application.ErrServiceUnavailable
	}
	var existing attemptDocument
	if findErr := store.attempts.FindOne(ctx, bson.M{"commandId": attempt.CommandID()}).Decode(&existing); findErr != nil {
		return domain.Attempt{}, false, application.ErrServiceUnavailable
	}
	value, findErr := fromDocument(existing)
	return value, true, findErr
}

func (store *Store) FindByID(ctx context.Context, id string) (domain.Attempt, error) {
	var document attemptDocument
	if err := store.attempts.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Attempt{}, application.ErrAttemptNotFound
		}
		return domain.Attempt{}, application.ErrServiceUnavailable
	}
	return fromDocument(document)
}

func (store *Store) Update(ctx context.Context, attempt domain.Attempt, expectedVersion uint64) error {
	result, err := store.attempts.ReplaceOne(
		ctx,
		bson.M{"_id": attempt.ID(), "version": expectedVersion},
		toDocument(attempt),
	)
	if err != nil {
		return application.ErrServiceUnavailable
	}
	if result.MatchedCount != 1 {
		return application.ErrOptimisticConflict
	}
	return nil
}

func (store *Store) Enqueue(ctx context.Context, task application.ReviewTask) error {
	now := time.Now().UTC()
	_, err := store.reviews.UpdateOne(
		ctx,
		bson.M{"_id": task.AttemptID},
		bson.M{"$set": bson.M{
			"voiceArtifactRef": task.VoiceArtifactRef,
			"faceArtifactRef":  task.FaceArtifactRef, "reason": string(task.Reason),
			"createdAt": now, "expiresAt": now.Add(24 * time.Hour),
		}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return application.ErrReviewQueueUnavailable
	}
	return nil
}

func (store *Store) Complete(ctx context.Context, attemptID string) error {
	if _, err := store.reviews.DeleteOne(ctx, bson.M{"_id": attemptID}); err != nil {
		return application.ErrReviewQueueUnavailable
	}
	return nil
}
