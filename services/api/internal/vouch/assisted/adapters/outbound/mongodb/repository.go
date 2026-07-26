package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/application"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/domain"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("assisted_vouch_requests")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "commands.id", Value: 1}},
			Options: options.Index().SetName("assisted_vouch_command").SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "subjectKey", Value: 1}, {Key: "voucherKey", Value: 1}},
			Options: options.Index().SetName("assisted_vouch_open_pair").SetUnique(true).
				SetPartialFilterExpression(bson.M{"status": bson.M{"$in": bson.A{
					domain.StatusAwaitingConsent, domain.StatusConsented,
				}}}),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("assisted_vouch_expiry"),
		},
	})
	return err
}

type eventDocument struct {
	Sequence   uint64        `bson:"sequence"`
	CommandID  string        `bson:"commandId"`
	ActorKey   string        `bson:"actorKey"`
	Action     domain.Action `bson:"action"`
	ReasonCode string        `bson:"reasonCode"`
	At         time.Time     `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type outcomeDocument struct {
	Decision    domain.Decision `bson:"decision"`
	ReasonCode  string          `bson:"reasonCode"`
	OperatorKey string          `bson:"operatorKey"`
	Provenance  string          `bson:"provenance"`
	DecidedAt   time.Time       `bson:"decidedAt"`
}

type requestDocument struct {
	ID           string            `bson:"_id"`
	SubjectKey   string            `bson:"subjectKey"`
	RequesterKey string            `bson:"requesterKey"`
	VoucherKey   string            `bson:"voucherKey"`
	Status       domain.Status     `bson:"status"`
	ExpiresAt    time.Time         `bson:"expiresAt"`
	ConsentedAt  *time.Time        `bson:"consentedAt,omitempty"`
	Outcome      *outcomeDocument  `bson:"outcome,omitempty"`
	Revision     uint64            `bson:"revision"`
	Events       []eventDocument   `bson:"events"`
	Commands     []commandDocument `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, request domain.Request) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(request))
	return repository.translateDuplicate(ctx, request.Commands()[0].ID, err)
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Request, error) {
	return repository.find(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindByCommand(ctx context.Context, commandID string) (domain.Request, error) {
	return repository.find(ctx, bson.M{"commands.id": commandID})
}

func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Request, error) {
	var document requestDocument
	if err := repository.collection.FindOne(ctx, filter).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Request{}, application.ErrNotFound
		}
		return domain.Request{}, err
	}
	return toDomain(document)
}

func (repository *Repository) Save(ctx context.Context, request domain.Request, expected uint64, commandID string) error {
	result, err := repository.collection.ReplaceOne(
		ctx, bson.M{"_id": request.ID(), "revision": expected}, toDocument(request),
	)
	if err != nil {
		return repository.translateDuplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		return repository.classifyConflict(ctx, request.ID(), commandID)
	}
	return nil
}

func (repository *Repository) translateDuplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if findErr := repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err(); findErr == nil {
		return application.ErrCommandApplied
	}
	return application.ErrOptimisticConflict
}

func (repository *Repository) classifyConflict(ctx context.Context, id, commandID string) error {
	if err := repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err(); err == nil {
		return application.ErrCommandApplied
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Err(); err == nil {
		return application.ErrOptimisticConflict
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrNotFound
	} else {
		return err
	}
}

func toDocument(request domain.Request) requestDocument {
	document := requestDocument{
		ID: request.ID(), SubjectKey: request.SubjectKey(), RequesterKey: request.RequesterKey(),
		VoucherKey: request.VoucherKey(), Status: request.Status(), ExpiresAt: request.ExpiresAt(),
		ConsentedAt: request.ConsentedAt(), Revision: request.Revision(),
	}
	if outcome := request.Outcome(); outcome != nil {
		document.Outcome = &outcomeDocument{
			Decision: outcome.Decision, ReasonCode: outcome.ReasonCode,
			OperatorKey: outcome.OperatorKey, Provenance: outcome.Provenance,
			DecidedAt: outcome.DecidedAt,
		}
	}
	for _, event := range request.Events() {
		document.Events = append(document.Events, eventDocument{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
			Action: event.Action, ReasonCode: event.ReasonCode, At: event.At,
		})
	}
	for _, command := range request.Commands() {
		document.Commands = append(document.Commands, commandDocument{
			ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
		})
	}
	return document
}

func toDomain(document requestDocument) (domain.Request, error) {
	state := domain.State{
		ID: document.ID, SubjectKey: document.SubjectKey, RequesterKey: document.RequesterKey,
		VoucherKey: document.VoucherKey, Status: document.Status, ExpiresAt: document.ExpiresAt,
		ConsentedAt: document.ConsentedAt, Revision: document.Revision,
	}
	if document.Outcome != nil {
		state.Outcome = &domain.Outcome{
			Decision: document.Outcome.Decision, ReasonCode: document.Outcome.ReasonCode,
			OperatorKey: document.Outcome.OperatorKey, Provenance: document.Outcome.Provenance,
			DecidedAt: document.Outcome.DecidedAt,
		}
	}
	for _, event := range document.Events {
		state.Events = append(state.Events, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
			Action: event.Action, ReasonCode: event.ReasonCode, At: event.At,
		})
	}
	for _, command := range document.Commands {
		state.Commands = append(state.Commands, domain.AppliedCommand{
			ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
		})
	}
	return domain.Rehydrate(state)
}
