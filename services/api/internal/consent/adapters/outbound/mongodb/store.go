// Package mongodb persists immutable, versioned consent histories.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	platformmongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/application"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

type Store struct {
	collection *mongo.Collection
}

type eventDocument struct {
	Revision       uint64    `bson:"revision"`
	CommandID      string    `bson:"commandId"`
	Action         string    `bson:"action"`
	PurposeVersion uint64    `bson:"purposeVersion"`
	ActorID        string    `bson:"actorId"`
	ActorKind      string    `bson:"actorKind"`
	Source         string    `bson:"source"`
	EvidenceKind   string    `bson:"evidenceKind"`
	EvidenceRef    string    `bson:"evidenceRef"`
	PolicyVersion  uint64    `bson:"policyVersion"`
	RecordedAt     time.Time `bson:"recordedAt"`
}

type recordDocument struct {
	ID        string          `bson:"_id"`
	SubjectID string          `bson:"subjectId"`
	PurposeID string          `bson:"purposeId"`
	Revision  uint64          `bson:"revision"`
	History   []eventDocument `bson:"history"`
}

func NewStore(database *mongo.Database) *Store {
	return &Store{collection: database.Collection("consent_records")}
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "subjectId", Value: 1}, {Key: "purposeId", Value: 1}},
			Options: options.Index().SetName("consent_subject_purpose_unique").
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "history.commandId", Value: 1}},
			Options: options.Index().SetName("consent_command_unique").
				SetUnique(true),
		},
	})
	return err
}

func recordID(key application.Key) string {
	return key.SubjectID + "|" + key.PurposeID
}

func (store *Store) Find(ctx context.Context, key application.Key) (domain.Record, error) {
	var document recordDocument
	if err := store.collection.FindOne(ctx, bson.M{"_id": recordID(key)}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Record{}, application.ErrNotFound
		}
		return domain.Record{}, application.ErrRepositoryUnavailable
	}
	events := make([]domain.Event, 0, len(document.History))
	for _, stored := range document.History {
		evidence, err := domain.NewEvidence(
			domain.EvidenceKind(stored.EvidenceKind),
			stored.PolicyVersion,
			stored.EvidenceRef,
		)
		if err != nil {
			return domain.Record{}, application.ErrRepositoryUnavailable
		}
		event, err := domain.NewEvent(domain.EventParams{
			Revision: stored.Revision, CommandID: stored.CommandID,
			Action: domain.Action(stored.Action), PurposeVersion: stored.PurposeVersion,
			ActorID: stored.ActorID, ActorKind: domain.ActorKind(stored.ActorKind),
			Source: domain.Source(stored.Source), Evidence: evidence,
			RecordedAt: stored.RecordedAt,
		})
		if err != nil {
			return domain.Record{}, application.ErrRepositoryUnavailable
		}
		events = append(events, event)
	}
	record, err := domain.Rehydrate(document.SubjectID, document.PurposeID, events)
	if err != nil || record.Revision() != document.Revision {
		return domain.Record{}, application.ErrRepositoryUnavailable
	}
	return record, nil
}

func (store *Store) Save(ctx context.Context, record domain.Record, expectedRevision uint64, commandID string) error {
	history := record.History()
	if len(history) == 0 {
		return application.ErrRepositoryUnavailable
	}
	event := history[len(history)-1]
	stored := eventDocument{
		Revision: event.Revision(), CommandID: event.CommandID(),
		Action: string(event.Action()), PurposeVersion: event.PurposeVersion(),
		ActorID: event.ActorID(), ActorKind: string(event.ActorKind()),
		Source: string(event.Source()), EvidenceKind: string(event.Evidence().Kind()),
		EvidenceRef:   event.Evidence().Reference(),
		PolicyVersion: event.Evidence().PolicyVersion(), RecordedAt: event.RecordedAt(),
	}
	result, err := store.collection.UpdateOne(
		ctx,
		bson.M{"_id": recordID(application.Key{SubjectID: record.SubjectID(), PurposeID: record.PurposeID()}), "revision": expectedRevision},
		bson.M{
			"$setOnInsert": bson.M{"subjectId": record.SubjectID(), "purposeId": record.PurposeID()},
			"$set":         bson.M{"revision": record.Revision()},
			"$push":        bson.M{"history": stored},
		},
		options.UpdateOne().SetUpsert(expectedRevision == 0),
	)
	if platformmongo.IsDuplicateKey(err) {
		current, findErr := store.Find(ctx, application.Key{SubjectID: record.SubjectID(), PurposeID: record.PurposeID()})
		if findErr == nil && current.HasCommand(commandID) {
			return application.ErrCommandAlreadyApplied
		}
		return application.ErrOptimisticConflict
	}
	if err != nil {
		return application.ErrRepositoryUnavailable
	}
	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		return application.ErrOptimisticConflict
	}
	return nil
}

type Catalog struct {
	purposes map[string]domain.Purpose
}

func NewCatalog(purposes ...domain.Purpose) Catalog {
	values := make(map[string]domain.Purpose, len(purposes))
	for _, purpose := range purposes {
		values[purpose.ID()] = purpose
	}
	return Catalog{purposes: values}
}

func (catalog Catalog) FindVersion(_ context.Context, id string, version uint64) (domain.Purpose, error) {
	purpose, ok := catalog.purposes[id]
	if !ok || purpose.Version() != version {
		return domain.Purpose{}, domain.ErrInvalidPurpose
	}
	return purpose, nil
}
