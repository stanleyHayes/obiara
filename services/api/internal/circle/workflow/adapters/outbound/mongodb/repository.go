package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.database.Collection("circle_workflow_invites").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "tokenDigest", Value: 1}}, Options: options.Index().SetName("circle_invite_token").SetUnique(true)},
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetName("circle_invite_command").SetUnique(true)},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetName("circle_invite_expiry")},
	})
	if err != nil {
		return err
	}
	_, err = repository.database.Collection("circle_workflow_requests").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetName("circle_request_command").SetUnique(true)},
		{
			Keys: bson.D{{Key: "circleId", Value: 1}, {Key: "memberId", Value: 1}},
			Options: options.Index().SetName("circle_active_request").SetUnique(true).
				SetPartialFilterExpression(bson.M{"status": bson.M{"$in": bson.A{"pending", "approved"}}}),
		},
	})
	return err
}

type eventDocument struct {
	Sequence   uint64        `bson:"sequence"`
	CommandID  string        `bson:"commandId"`
	ActorID    string        `bson:"actorId"`
	Action     domain.Action `bson:"action"`
	ReasonCode string        `bson:"reasonCode"`
	At         time.Time     `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type inviteDocument struct {
	ID          string              `bson:"_id"`
	CircleID    string              `bson:"circleId"`
	TokenDigest string              `bson:"tokenDigest"`
	Status      domain.InviteStatus `bson:"status"`
	ExpiresAt   time.Time           `bson:"expiresAt"`
	Revision    uint64              `bson:"revision"`
	Events      []eventDocument     `bson:"events"`
	Commands    []commandDocument   `bson:"commands"`
}

type requestDocument struct {
	ID       string               `bson:"_id"`
	CircleID string               `bson:"circleId"`
	MemberID string               `bson:"memberId"`
	Source   string               `bson:"source"`
	Status   domain.RequestStatus `bson:"status"`
	Revision uint64               `bson:"revision"`
	Events   []eventDocument      `bson:"events"`
	Commands []commandDocument    `bson:"commands"`
}

func (repository *Repository) CreateInvite(ctx context.Context, invite domain.Invite) error {
	_, err := repository.database.Collection("circle_workflow_invites").InsertOne(ctx, inviteToDocument(invite))
	return repository.translateDuplicate(ctx, "circle_workflow_invites", invite.Commands()[0].ID, err)
}

func (repository *Repository) FindInviteByDigest(ctx context.Context, digest string) (domain.Invite, error) {
	return repository.findInvite(ctx, bson.M{"tokenDigest": digest})
}

func (repository *Repository) FindInviteByCommand(ctx context.Context, commandID string) (domain.Invite, error) {
	return repository.findInvite(ctx, bson.M{"commands.id": commandID})
}

func (repository *Repository) findInvite(ctx context.Context, filter bson.M) (domain.Invite, error) {
	var document inviteDocument
	if err := repository.database.Collection("circle_workflow_invites").FindOne(ctx, filter).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Invite{}, application.ErrNotFound
		}
		return domain.Invite{}, err
	}
	return inviteToDomain(document)
}

func (repository *Repository) CreateRequest(ctx context.Context, request domain.Request) error {
	_, err := repository.database.Collection("circle_workflow_requests").InsertOne(ctx, requestToDocument(request))
	return repository.translateDuplicate(ctx, "circle_workflow_requests", request.Commands()[0].ID, err)
}

func (repository *Repository) FindRequest(ctx context.Context, id string) (domain.Request, error) {
	return repository.findRequest(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindRequestByCommand(ctx context.Context, commandID string) (domain.Request, error) {
	return repository.findRequest(ctx, bson.M{"commands.id": commandID})
}

func (repository *Repository) findRequest(ctx context.Context, filter bson.M) (domain.Request, error) {
	var document requestDocument
	if err := repository.database.Collection("circle_workflow_requests").FindOne(ctx, filter).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Request{}, application.ErrNotFound
		}
		return domain.Request{}, err
	}
	return requestToDomain(document)
}

func (repository *Repository) SaveRequest(ctx context.Context, request domain.Request, expected uint64, commandID string) error {
	result, err := repository.database.Collection("circle_workflow_requests").ReplaceOne(
		ctx, bson.M{"_id": request.ID(), "revision": expected}, requestToDocument(request),
	)
	if err != nil {
		return repository.translateDuplicate(ctx, "circle_workflow_requests", commandID, err)
	}
	if result.MatchedCount == 0 {
		return repository.classifyConflict(ctx, "circle_workflow_requests", request.ID(), commandID)
	}
	return nil
}

func (repository *Repository) Redeem(ctx context.Context, invite domain.Invite, request domain.Request, expectedInvite uint64, commandID string) error {
	session, err := repository.database.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(transactionContext context.Context) (any, error) {
		result, err := repository.database.Collection("circle_workflow_invites").ReplaceOne(
			transactionContext,
			bson.M{"_id": invite.ID(), "revision": expectedInvite, "status": domain.InviteActive},
			inviteToDocument(invite),
		)
		if err != nil {
			return nil, err
		}
		if result.MatchedCount == 0 {
			return nil, application.ErrOptimisticConflict
		}
		_, err = repository.database.Collection("circle_workflow_requests").InsertOne(transactionContext, requestToDocument(request))
		return nil, err
	})
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		if _, findErr := repository.FindRequestByCommand(ctx, commandID+".request"); findErr == nil {
			return application.ErrCommandApplied
		}
	}
	if errors.Is(err, application.ErrOptimisticConflict) {
		if existing, findErr := repository.FindInviteByCommand(ctx, commandID); findErr == nil && existing.HasCommand(commandID) {
			return application.ErrCommandApplied
		}
	}
	return err
}

func (repository *Repository) classifyConflict(ctx context.Context, collection, id, commandID string) error {
	store := repository.database.Collection(collection)
	if err := store.FindOne(ctx, bson.M{"commands.id": commandID}).Err(); err == nil {
		return application.ErrCommandApplied
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if err := store.FindOne(ctx, bson.M{"_id": id}).Err(); err == nil {
		return application.ErrOptimisticConflict
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrNotFound
	} else {
		return err
	}
}

func (repository *Repository) translateDuplicate(ctx context.Context, collection, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if findErr := repository.database.Collection(collection).FindOne(ctx, bson.M{"commands.id": commandID}).Err(); findErr == nil {
		return application.ErrCommandApplied
	}
	return application.ErrOptimisticConflict
}

func inviteToDocument(invite domain.Invite) inviteDocument {
	return inviteDocument{
		ID: invite.ID(), CircleID: invite.CircleID(), TokenDigest: invite.TokenDigest(),
		Status: invite.Status(), ExpiresAt: invite.ExpiresAt(), Revision: invite.Revision(),
		Events: eventsToDocuments(invite.Events()), Commands: commandsToDocuments(invite.Commands()),
	}
}

func inviteToDomain(document inviteDocument) (domain.Invite, error) {
	return domain.RehydrateInvite(domain.InviteState{
		ID: document.ID, CircleID: document.CircleID, TokenDigest: document.TokenDigest,
		Status: document.Status, ExpiresAt: document.ExpiresAt, Revision: document.Revision,
		Events: eventsToDomain(document.Events), Commands: commandsToDomain(document.Commands),
	})
}

func requestToDocument(request domain.Request) requestDocument {
	return requestDocument{
		ID: request.ID(), CircleID: request.CircleID(), MemberID: request.MemberID(),
		Source: request.Source(), Status: request.Status(), Revision: request.Revision(),
		Events: eventsToDocuments(request.Events()), Commands: commandsToDocuments(request.Commands()),
	}
}

func requestToDomain(document requestDocument) (domain.Request, error) {
	return domain.RehydrateRequest(domain.RequestState{
		ID: document.ID, CircleID: document.CircleID, MemberID: document.MemberID,
		Source: document.Source, Status: document.Status, Revision: document.Revision,
		Events: eventsToDomain(document.Events), Commands: commandsToDomain(document.Commands),
	})
}

func eventsToDocuments(events []domain.Event) []eventDocument {
	result := make([]eventDocument, 0, len(events))
	for _, event := range events {
		result = append(result, eventDocument{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorID: event.ActorID,
			Action: event.Action, ReasonCode: event.ReasonCode, At: event.At,
		})
	}
	return result
}

func eventsToDomain(events []eventDocument) []domain.Event {
	result := make([]domain.Event, 0, len(events))
	for _, event := range events {
		result = append(result, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorID: event.ActorID,
			Action: event.Action, ReasonCode: event.ReasonCode, At: event.At,
		})
	}
	return result
}

func commandsToDocuments(commands []domain.AppliedCommand) []commandDocument {
	result := make([]commandDocument, 0, len(commands))
	for _, command := range commands {
		result = append(result, commandDocument{ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision})
	}
	return result
}

func commandsToDomain(commands []commandDocument) []domain.AppliedCommand {
	result := make([]domain.AppliedCommand, 0, len(commands))
	for _, command := range commands {
		result = append(result, domain.AppliedCommand{ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision})
	}
	return result
}
