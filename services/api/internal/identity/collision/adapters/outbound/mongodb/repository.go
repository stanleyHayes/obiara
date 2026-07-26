package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.database.Collection("identity_collision_cases").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: 1}},
		Options: options.Index().SetName("collision_review_queue"),
	})
	return err
}

type signalDocument struct {
	ID          string   `bson:"_id"`
	Kind        string   `bson:"kind"`
	SubjectKeys []string `bson:"subjectKeys"`
}

type auditDocument struct {
	Sequence   int64     `bson:"sequence"`
	From       string    `bson:"from,omitempty"`
	To         string    `bson:"to"`
	ReasonCode string    `bson:"reasonCode"`
	ActorKey   string    `bson:"actorKey"`
	OccurredAt time.Time `bson:"occurredAt"`
}

type caseDocument struct {
	ID         string          `bson:"_id"`
	Kind       string          `bson:"kind"`
	SignalKey  string          `bson:"signalKey"`
	SubjectKey string          `bson:"subjectKey"`
	Status     string          `bson:"status"`
	ReasonCode string          `bson:"reasonCode,omitempty"`
	Version    int64           `bson:"version"`
	CreatedAt  time.Time       `bson:"createdAt"`
	ResolvedAt *time.Time      `bson:"resolvedAt,omitempty"`
	Audit      []auditDocument `bson:"audit"`
}

func signalID(kind domain.Kind, signalKey string) string {
	return string(kind) + ":" + signalKey
}

func (repository *Repository) RegisterSignal(ctx context.Context, kind domain.Kind, signalKey, subjectKey string) (bool, error) {
	var document signalDocument
	collection := repository.database.Collection("identity_collision_signals")
	update := bson.M{
		"$setOnInsert": bson.M{"kind": string(kind)},
		"$addToSet":    bson.M{"subjectKeys": subjectKey},
	}
	err := collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": signalID(kind, signalKey)},
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&document)
	// Concurrent first observations can race on the upsert. The losing writer
	// retries against the now-existing document instead of dropping a signal.
	if mongo.IsDuplicateKeyError(err) {
		err = collection.FindOneAndUpdate(
			ctx,
			bson.M{"_id": signalID(kind, signalKey)},
			update,
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&document)
	}
	if err != nil {
		return false, err
	}
	return len(document.SubjectKeys) > 1, nil
}

func (repository *Repository) Create(ctx context.Context, reviewCase domain.Case, audit domain.AuditEvent) (domain.Case, bool, error) {
	document := toDocument(reviewCase, []domain.AuditEvent{audit})
	_, err := repository.database.Collection("identity_collision_cases").InsertOne(ctx, document)
	if err == nil {
		return reviewCase, true, nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return domain.Case{}, false, err
	}
	existing, findErr := repository.FindByID(ctx, reviewCase.ID())
	return existing, false, findErr
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Case, error) {
	var document caseDocument
	if err := repository.database.Collection("identity_collision_cases").FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Case{}, application.ErrCaseNotFound
		}
		return domain.Case{}, err
	}
	return toDomain(document), nil
}

func (repository *Repository) Resolve(ctx context.Context, reviewCase domain.Case, audit domain.AuditEvent, previousVersion int64) error {
	result, err := repository.database.Collection("identity_collision_cases").UpdateOne(
		ctx,
		bson.M{"_id": reviewCase.ID(), "version": previousVersion, "status": string(domain.StatusPending)},
		bson.M{
			"$set": bson.M{
				"status": reviewCase.Status(), "reasonCode": reviewCase.ReasonCode(),
				"version": reviewCase.Version(), "resolvedAt": reviewCase.ResolvedAt(),
			},
			"$push": bson.M{"audit": toAuditDocument(audit)},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrStaleCase
	}
	return nil
}

func toDocument(reviewCase domain.Case, audit []domain.AuditEvent) caseDocument {
	events := make([]auditDocument, 0, len(audit))
	for _, event := range audit {
		events = append(events, toAuditDocument(event))
	}
	return caseDocument{
		ID: reviewCase.ID(), Kind: string(reviewCase.Kind()), SignalKey: reviewCase.SignalKey(),
		SubjectKey: reviewCase.SubjectKey(), Status: string(reviewCase.Status()),
		ReasonCode: reviewCase.ReasonCode(), Version: reviewCase.Version(),
		CreatedAt: reviewCase.CreatedAt(), ResolvedAt: reviewCase.ResolvedAt(), Audit: events,
	}
}

func toAuditDocument(event domain.AuditEvent) auditDocument {
	return auditDocument{
		Sequence: event.Sequence, From: string(event.From), To: string(event.To),
		ReasonCode: event.ReasonCode, ActorKey: event.ActorKey, OccurredAt: event.OccurredAt,
	}
}

func toDomain(document caseDocument) domain.Case {
	return domain.ReconstituteCase(
		document.ID, domain.Kind(document.Kind), document.SignalKey, document.SubjectKey,
		domain.Status(document.Status), document.ReasonCode, document.Version,
		document.CreatedAt, document.ResolvedAt,
	)
}
