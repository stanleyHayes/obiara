package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
)

type Repository struct {
	collection *mongo.Collection
}

type membershipDocument struct {
	MemberID  string                 `bson:"memberId"`
	State     domain.MembershipState `bson:"state"`
	UpdatedAt time.Time              `bson:"updatedAt"`
}

type transitionDocument struct {
	Revision  uint64                 `bson:"revision"`
	CommandID string                 `bson:"commandId"`
	ActorID   string                 `bson:"actorId"`
	MemberID  string                 `bson:"memberId"`
	From      string                 `bson:"from"`
	To        domain.MembershipState `bson:"to"`
	At        time.Time              `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type circleDocument struct {
	ID          string               `bson:"_id"`
	Type        domain.Type          `bson:"type"`
	Visibility  domain.Visibility    `bson:"visibility"`
	Memberships []membershipDocument `bson:"memberships"`
	History     []transitionDocument `bson:"history"`
	Commands    []commandDocument    `bson:"commands"`
	Revision    uint64               `bson:"revision"`
	UpdatedAt   time.Time            `bson:"updatedAt"`
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("circles")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "commands.id", Value: 1}},
			Options: options.Index().SetName("circles_command_id_unique").SetUnique(true).
				SetPartialFilterExpression(bson.M{"commands.id": bson.M{"$exists": true}}),
		},
		{
			Keys:    bson.D{{Key: "memberships.memberId", Value: 1}, {Key: "memberships.state", Value: 1}},
			Options: options.Index().SetName("circles_membership_lookup"),
		},
	})
	return err
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Circle, error) {
	var document circleDocument
	if err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Circle{}, application.ErrNotFound
		}
		return domain.Circle{}, err
	}
	return toDomain(document)
}

func (repository *Repository) Save(ctx context.Context, circle domain.Circle, expectedRevision uint64, commandID string) error {
	if !circle.HasCommand(commandID) || circle.Revision() != expectedRevision+1 {
		return domain.ErrInvalidCircle
	}
	document := fromDomain(circle)
	if expectedRevision == 0 {
		if _, err := repository.collection.InsertOne(ctx, document); err != nil {
			return repository.translateWriteError(ctx, circle.ID(), commandID, err)
		}
		return nil
	}
	result, err := repository.collection.ReplaceOne(ctx, bson.M{
		"_id": circle.ID(), "revision": expectedRevision,
	}, document)
	if err != nil {
		return repository.translateWriteError(ctx, circle.ID(), commandID, err)
	}
	if result.MatchedCount == 0 {
		return repository.classifyConflict(ctx, circle.ID(), commandID)
	}
	return nil
}

func (repository *Repository) translateWriteError(ctx context.Context, circleID, commandID string, err error) error {
	if !apimongo.IsDuplicateKey(err) {
		return err
	}
	return repository.classifyConflict(ctx, circleID, commandID)
}

func (repository *Repository) classifyConflict(ctx context.Context, circleID, commandID string) error {
	if err := repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err(); err == nil {
		return application.ErrCommandAlreadyApplied
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if err := repository.collection.FindOne(ctx, bson.M{"_id": circleID}).Err(); err == nil || errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrOptimisticConflict
	} else {
		return err
	}
}

func fromDomain(circle domain.Circle) circleDocument {
	document := circleDocument{
		ID: circle.ID(), Type: circle.Type(), Visibility: circle.Visibility(),
		Revision: circle.Revision(), UpdatedAt: circle.UpdatedAt(),
	}
	for _, membership := range circle.Memberships() {
		document.Memberships = append(document.Memberships, membershipDocument{
			MemberID: membership.MemberID(), State: membership.State(), UpdatedAt: membership.UpdatedAt(),
		})
	}
	for _, event := range circle.History() {
		document.History = append(document.History, transitionDocument{
			Revision: event.Revision(), CommandID: event.CommandID(), ActorID: event.ActorID(),
			MemberID: event.MemberID(), From: event.From(), To: event.To(), At: event.At(),
		})
	}
	for _, command := range circle.Commands() {
		document.Commands = append(document.Commands, commandDocument{
			ID: command.ID(), Fingerprint: command.Fingerprint(), Revision: command.Revision(),
		})
	}
	return document
}

func toDomain(document circleDocument) (domain.Circle, error) {
	state := domain.State{
		ID: document.ID, Type: document.Type, Visibility: document.Visibility,
		Revision: document.Revision, UpdatedAt: document.UpdatedAt,
	}
	for _, persisted := range document.Memberships {
		membership, err := domain.NewMembership(persisted.MemberID, persisted.State, persisted.UpdatedAt)
		if err != nil {
			return domain.Circle{}, err
		}
		state.Memberships = append(state.Memberships, membership)
	}
	for _, persisted := range document.History {
		event, err := domain.NewTransition(
			persisted.Revision, persisted.CommandID, persisted.ActorID,
			persisted.MemberID, persisted.From, persisted.To, persisted.At,
		)
		if err != nil {
			return domain.Circle{}, err
		}
		state.History = append(state.History, event)
	}
	for _, persisted := range document.Commands {
		command, err := domain.NewAppliedCommand(persisted.ID, persisted.Fingerprint, persisted.Revision)
		if err != nil {
			return domain.Circle{}, err
		}
		state.Commands = append(state.Commands, command)
	}
	return domain.Rehydrate(state)
}
