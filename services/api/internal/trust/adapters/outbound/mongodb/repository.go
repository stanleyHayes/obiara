package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

type Repository struct {
	collection *mongo.Collection
}

type eventDocument struct {
	Revision  uint64        `bson:"revision"`
	CommandID string        `bson:"commandId"`
	Action    domain.Action `bson:"action"`
	ActorRef  string        `bson:"actorRef"`
	At        time.Time     `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type edgeDocument struct {
	ID            string            `bson:"_id"`
	SourceID      string            `bson:"sourceId"`
	TargetID      string            `bson:"targetId"`
	Type          domain.EdgeType   `bson:"type"`
	ProvenanceRef string            `bson:"provenanceRef"`
	ConsentRef    string            `bson:"consentRef"`
	Visibility    domain.Visibility `bson:"visibility"`
	CreatedAt     time.Time         `bson:"createdAt"`
	ExpiresAt     *time.Time        `bson:"expiresAt,omitempty"`
	RevokedAt     *time.Time        `bson:"revokedAt,omitempty"`
	Revision      uint64            `bson:"revision"`
	History       []eventDocument   `bson:"history"`
	Commands      []commandDocument `bson:"commands"`
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("trust_edges")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "commands.id", Value: 1}},
			Options: options.Index().SetName("trust_command_id_unique").SetUnique(true).
				SetPartialFilterExpression(bson.M{"commands.id": bson.M{"$exists": true}}),
		},
		{
			Keys:    bson.D{{Key: "sourceId", Value: 1}, {Key: "revokedAt", Value: 1}, {Key: "expiresAt", Value: 1}},
			Options: options.Index().SetName("trust_bounded_outgoing"),
		},
	})
	return err
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Edge, error) {
	var document edgeDocument
	if err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Edge{}, application.ErrNotFound
		}
		return domain.Edge{}, err
	}
	return toDomain(document)
}

// Outgoing is intentionally the only traversal read. There is no ListAll or
// reverse global-browser method on the repository port.
func (repository *Repository) Outgoing(ctx context.Context, sourceIDs []string) ([]domain.Edge, error) {
	if len(sourceIDs) == 0 || len(sourceIDs) > domain.MaxProjectionNodes {
		return nil, domain.ErrProjectionBounds
	}
	cursor, err := repository.collection.Find(
		ctx,
		bson.M{"sourceId": bson.M{"$in": sourceIDs}},
		options.Find().SetSort(bson.D{{Key: "sourceId", Value: 1}, {Key: "_id", Value: 1}}).
			SetLimit(int64(domain.MaxProjectionNodes*domain.MaxProjectionDepth)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []edgeDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	edges := make([]domain.Edge, 0, len(documents))
	for _, document := range documents {
		edge, err := toDomain(document)
		if err != nil {
			return nil, err
		}
		edges = append(edges, edge)
	}
	return edges, nil
}

func (repository *Repository) Save(ctx context.Context, edge domain.Edge, expectedRevision uint64, commandID string) error {
	if !edge.HasCommand(commandID) || edge.Revision() != expectedRevision+1 {
		return domain.ErrInvalidEdge
	}
	document := fromDomain(edge)
	if expectedRevision == 0 {
		if _, err := repository.collection.InsertOne(ctx, document); err != nil {
			return repository.translateWriteError(ctx, edge.ID(), commandID, err)
		}
		return nil
	}
	result, err := repository.collection.ReplaceOne(ctx, bson.M{
		"_id": edge.ID(), "revision": expectedRevision,
	}, document)
	if err != nil {
		return repository.translateWriteError(ctx, edge.ID(), commandID, err)
	}
	if result.MatchedCount == 0 {
		return repository.classifyConflict(ctx, edge.ID(), commandID)
	}
	return nil
}

func (repository *Repository) translateWriteError(ctx context.Context, edgeID, commandID string, err error) error {
	if !apimongo.IsDuplicateKey(err) {
		return err
	}
	return repository.classifyConflict(ctx, edgeID, commandID)
}

func (repository *Repository) classifyConflict(ctx context.Context, edgeID, commandID string) error {
	if err := repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err(); err == nil {
		return application.ErrCommandAlreadyApplied
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if err := repository.collection.FindOne(ctx, bson.M{"_id": edgeID}).Err(); err == nil || errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrOptimisticConflict
	} else {
		return err
	}
}

func fromDomain(edge domain.Edge) edgeDocument {
	document := edgeDocument{
		ID: edge.ID(), SourceID: edge.SourceID(), TargetID: edge.TargetID(), Type: edge.Type(),
		ProvenanceRef: edge.ProvenanceRef(), ConsentRef: edge.ConsentRef(), Visibility: edge.Visibility(),
		CreatedAt: edge.CreatedAt(), ExpiresAt: edge.ExpiresAt(), RevokedAt: edge.RevokedAt(),
		Revision: edge.Revision(),
	}
	for _, event := range edge.History() {
		document.History = append(document.History, eventDocument{
			Revision: event.Revision(), CommandID: event.CommandID(), Action: event.Action(),
			ActorRef: event.ActorRef(), At: event.At(),
		})
	}
	for _, command := range edge.Commands() {
		document.Commands = append(document.Commands, commandDocument{
			ID: command.ID(), Fingerprint: command.Fingerprint(), Revision: command.Revision(),
		})
	}
	return document
}

func toDomain(document edgeDocument) (domain.Edge, error) {
	state := domain.State{
		Params: domain.Params{
			ID: document.ID, SourceID: document.SourceID, TargetID: document.TargetID,
			Type: document.Type, ProvenanceRef: document.ProvenanceRef, ConsentRef: document.ConsentRef,
			Visibility: document.Visibility, CreatedAt: document.CreatedAt, ExpiresAt: document.ExpiresAt,
		},
		RevokedAt: document.RevokedAt, Revision: document.Revision,
	}
	for _, persisted := range document.History {
		event, err := domain.NewEvent(
			persisted.Revision, persisted.CommandID, persisted.Action, persisted.ActorRef, persisted.At,
		)
		if err != nil {
			return domain.Edge{}, err
		}
		state.History = append(state.History, event)
	}
	for _, persisted := range document.Commands {
		command, err := domain.NewAppliedCommand(persisted.ID, persisted.Fingerprint, persisted.Revision)
		if err != nil {
			return domain.Edge{}, err
		}
		state.Commands = append(state.Commands, command)
	}
	return domain.Rehydrate(state)
}
