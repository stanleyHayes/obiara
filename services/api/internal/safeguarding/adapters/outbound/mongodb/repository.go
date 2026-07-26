package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/application"
	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) restrictions() *mongo.Collection {
	return repository.database.Collection("safeguarding_restrictions")
}

func (repository *Repository) jobs() *mongo.Collection {
	return repository.database.Collection("safeguarding_purge_jobs")
}

func (repository *Repository) events() *mongo.Collection {
	return repository.database.Collection("safeguarding_events")
}

type restrictionDocument struct {
	ID          string    `bson:"_id"`
	CommandID   string    `bson:"commandId"`
	SubjectKey  string    `bson:"subjectKey"`
	SourceKey   string    `bson:"sourceKey"`
	BlockedAt   time.Time `bson:"blockedAt"`
	PurgeDueAt  time.Time `bson:"purgeDueAt"`
	PurgeStatus string    `bson:"purgeStatus"`
	PurgedAt    time.Time `bson:"purgedAt,omitempty"`
	Version     uint64    `bson:"version"`
}

type purgeJobDocument struct {
	RestrictionID string    `bson:"_id"`
	SubjectID     string    `bson:"subjectId"`
	SourceRef     string    `bson:"sourceRef"`
	PurgeDueAt    time.Time `bson:"purgeDueAt"`
}

type eventDocument struct {
	RestrictionID string    `bson:"restrictionId"`
	CommandID     string    `bson:"commandId"`
	Kind          string    `bson:"kind"`
	OccurredAt    time.Time `bson:"occurredAt"`
	Version       uint64    `bson:"version"`
	WithinSLA     bool      `bson:"withinSla,omitempty"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.restrictions().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "commandId", Value: 1}},
		Options: options.Index().SetName("safeguarding_command_unique").SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := repository.restrictions().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "subjectKey", Value: 1}},
		Options: options.Index().SetName("safeguarding_subject_unique").SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := repository.jobs().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "purgeDueAt", Value: 1}},
		Options: options.Index().SetName("safeguarding_purge_due"),
	}); err != nil {
		return err
	}
	_, err := repository.events().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "restrictionId", Value: 1}, {Key: "version", Value: 1}},
		Options: options.Index().SetName("safeguarding_event_version").SetUnique(true),
	})
	return err
}

func (repository *Repository) CreateBlocked(ctx context.Context, restriction domain.Restriction, job application.PurgeJob) (domain.Restriction, bool, error) {
	err := apimongo.WithTransaction(ctx, repository.database.Client(), func(transaction context.Context) error {
		if _, err := repository.restrictions().InsertOne(transaction, toDocument(restriction)); err != nil {
			return err
		}
		if _, err := repository.jobs().InsertOne(transaction, purgeJobDocument{
			RestrictionID: job.RestrictionID, SubjectID: job.SubjectID,
			SourceRef: job.SourceRef, PurgeDueAt: job.PurgeDueAt,
		}); err != nil {
			return err
		}
		_, err := repository.events().InsertOne(transaction, eventDocument{
			RestrictionID: restriction.ID(), CommandID: restriction.CommandID(),
			Kind: "blocked", OccurredAt: restriction.BlockedAt(), Version: restriction.Version(),
		})
		return err
	})
	if err == nil {
		return restriction, false, nil
	}
	if !apimongo.IsDuplicateKey(err) {
		return domain.Restriction{}, false, err
	}
	existing, findErr := repository.findOne(ctx, bson.M{"commandId": restriction.CommandID()})
	if errors.Is(findErr, application.ErrRestrictionNotFound) {
		existing, findErr = repository.FindBySubjectKey(ctx, restriction.SubjectKey())
	}
	if findErr != nil {
		return domain.Restriction{}, false, findErr
	}
	if existing.SubjectKey() != restriction.SubjectKey() ||
		(existing.CommandID() == restriction.CommandID() && existing.SourceKey() != restriction.SourceKey()) {
		return domain.Restriction{}, false, application.ErrCommandConflict
	}
	return existing, true, nil
}

func (repository *Repository) FindPending(ctx context.Context, dueBefore time.Time, limit int) ([]application.PurgeJob, error) {
	cursor, err := repository.jobs().Find(ctx,
		bson.M{"purgeDueAt": bson.M{"$lte": dueBefore.UTC()}},
		options.Find().SetSort(bson.D{{Key: "purgeDueAt", Value: 1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []purgeJobDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	jobs := make([]application.PurgeJob, 0, len(documents))
	for _, document := range documents {
		jobs = append(jobs, application.PurgeJob{
			RestrictionID: document.RestrictionID, SubjectID: document.SubjectID,
			SourceRef: document.SourceRef, PurgeDueAt: document.PurgeDueAt,
		})
	}
	return jobs, nil
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Restriction, error) {
	return repository.findOne(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindBySubjectKey(ctx context.Context, subjectKey string) (domain.Restriction, error) {
	return repository.findOne(ctx, bson.M{"subjectKey": subjectKey})
}

func (repository *Repository) findOne(ctx context.Context, filter bson.M) (domain.Restriction, error) {
	var document restrictionDocument
	if err := repository.restrictions().FindOne(ctx, filter).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Restriction{}, application.ErrRestrictionNotFound
		}
		return domain.Restriction{}, err
	}
	return domain.Rehydrate(
		document.ID, document.CommandID, document.SubjectKey, document.SourceKey,
		document.BlockedAt, document.PurgeDueAt, domain.PurgeStatus(document.PurgeStatus),
		document.PurgedAt, document.Version,
	)
}

func (repository *Repository) CompletePurge(ctx context.Context, restriction domain.Restriction, expectedVersion uint64) error {
	return apimongo.WithTransaction(ctx, repository.database.Client(), func(transaction context.Context) error {
		result, err := repository.restrictions().UpdateOne(transaction,
			bson.M{"_id": restriction.ID(), "version": expectedVersion, "purgeStatus": string(domain.PurgePending)},
			bson.M{"$set": bson.M{
				"purgeStatus": string(restriction.PurgeStatus()), "purgedAt": restriction.PurgedAt(),
				"version": restriction.Version(),
			}},
		)
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return application.ErrOptimisticConflict
		}
		if _, err := repository.jobs().DeleteOne(transaction, bson.M{"_id": restriction.ID()}); err != nil {
			return err
		}
		_, err = repository.events().InsertOne(transaction, eventDocument{
			RestrictionID: restriction.ID(), CommandID: restriction.CommandID(),
			Kind: "purged", OccurredAt: restriction.PurgedAt(), Version: restriction.Version(),
			WithinSLA: restriction.PurgedWithinSLA(),
		})
		return err
	})
}

func toDocument(restriction domain.Restriction) restrictionDocument {
	return restrictionDocument{
		ID: restriction.ID(), CommandID: restriction.CommandID(),
		SubjectKey: restriction.SubjectKey(), SourceKey: restriction.SourceKey(),
		BlockedAt: restriction.BlockedAt(), PurgeDueAt: restriction.PurgeDueAt(),
		PurgeStatus: string(restriction.PurgeStatus()), PurgedAt: restriction.PurgedAt(),
		Version: restriction.Version(),
	}
}
