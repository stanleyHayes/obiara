// Package mongodb is the append-only suban event store. Events are never
// updated or deleted through this adapter (Doc 08 §4: auditable,
// recomputable).
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
	explanationapp "github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/application"
)

type Store struct {
	database *mongo.Database
}

func NewStore(database *mongo.Database) *Store {
	return &Store{database: database}
}

func (store *Store) collection() *mongo.Collection {
	return store.database.Collection("suban_events")
}

type document struct {
	ID         string    `bson:"_id"`
	SubjectID  string    `bson:"subjectId"`
	Kind       string    `bson:"kind"`
	Source     string    `bson:"source"`
	Ref        string    `bson:"ref"`
	OccurredAt time.Time `bson:"occurredAt"`
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "subjectId", Value: 1}, {Key: "occurredAt", Value: -1}},
			Options: options.Index().SetName("suban_subject_time"),
		},
		{
			Keys:    bson.D{{Key: "subjectId", Value: 1}, {Key: "kind", Value: 1}, {Key: "occurredAt", Value: -1}},
			Options: options.Index().SetName("suban_subject_kind_time"),
		},
	})
	return err
}

func (store *Store) Append(ctx context.Context, event domain.Event) error {
	_, err := store.collection().InsertOne(ctx, document{
		ID:         event.ID,
		SubjectID:  event.SubjectID,
		Kind:       string(event.Kind),
		Source:     event.Provenance.Source,
		Ref:        event.Provenance.Ref,
		OccurredAt: event.OccurredAt.UTC(),
	})
	return err
}

func (store *Store) ListForSubject(ctx context.Context, subjectID string) ([]domain.Event, error) {
	cursor, err := store.collection().Find(ctx,
		bson.M{"subjectId": subjectID},
		options.Find().SetSort(bson.D{{Key: "occurredAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []domain.Event
	for cursor.Next(ctx) {
		var doc document
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		events = append(events, domain.Event{
			ID:         doc.ID,
			SubjectID:  doc.SubjectID,
			Kind:       domain.Kind(doc.Kind),
			Provenance: domain.Provenance{Source: doc.Source, Ref: doc.Ref},
			OccurredAt: doc.OccurredAt,
		})
	}
	return events, cursor.Err()
}

func (store *Store) FindForSubject(ctx context.Context, subjectID, eventID string) (domain.Event, error) {
	var doc document
	err := store.collection().FindOne(ctx, bson.M{"_id": eventID, "subjectId": subjectID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Event{}, explanationapp.ErrNotFound
	}
	if err != nil {
		return domain.Event{}, err
	}
	return domain.Event{
		ID: doc.ID, SubjectID: doc.SubjectID, Kind: domain.Kind(doc.Kind),
		Provenance: domain.Provenance{Source: doc.Source, Ref: doc.Ref},
		OccurredAt: doc.OccurredAt,
	}, nil
}

func (store *Store) CountForSubjectSince(ctx context.Context, subjectID string, kind domain.Kind, since time.Time) (int, error) {
	count, err := store.collection().CountDocuments(ctx, bson.M{
		"subjectId":  subjectID,
		"kind":       string(kind),
		"occurredAt": bson.M{"$gte": since.UTC()},
	})
	return int(count), err
}
