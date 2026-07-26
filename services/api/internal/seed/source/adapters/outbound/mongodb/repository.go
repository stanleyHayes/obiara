package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("seed_source_requests")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetName("seed_source_command").SetUnique(true)},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}, {Key: "status", Value: 1}}, Options: options.Index().SetName("seed_source_expiry")},
	})
	return err
}

type eventDocument struct {
	Sequence                        uint64 `bson:"sequence"`
	CommandID, ActorKey, ReasonCode string
	Action                          domain.Action `bson:"action"`
	At                              time.Time     `bson:"at"`
}
type commandDocument struct {
	ID, Fingerprint string
	Revision        uint64 `bson:"revision"`
}
type document struct {
	ID           string            `bson:"_id"`
	RequesterKey string            `bson:"requesterKey"`
	Source       domain.Source     `bson:"source"`
	CandidateIDs []string          `bson:"candidateIds"`
	Status       domain.Status     `bson:"status"`
	ExpiresAt    time.Time         `bson:"expiresAt"`
	EndedAt      *time.Time        `bson:"endedAt,omitempty"`
	Revision     uint64            `bson:"revision"`
	Events       []eventDocument   `bson:"events"`
	Commands     []commandDocument `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, request domain.Request) error {
	_, err := r.collection.InsertOne(ctx, toDocument(request))
	return r.translateDuplicate(ctx, request.Commands()[0].ID, err)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Request, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Request, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, filter bson.M) (domain.Request, error) {
	var stored document
	if err := r.collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Request{}, application.ErrNotFound
		}
		return domain.Request{}, err
	}
	return toDomain(stored)
}
func (r *Repository) Append(ctx context.Context, request domain.Request, expected uint64, commandID string) error {
	events, commands := request.Events(), request.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalidRequest
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": request.ID(), "revision": expected}, bson.M{
		"$set":  bson.M{"status": request.Status(), "revision": request.Revision(), "endedAt": request.EndedAt()},
		"$push": bson.M{"events": eventToDocument(events[len(events)-1]), "commands": commandToDocument(commands[len(commands)-1])},
	})
	if err != nil {
		return r.translateDuplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		return r.classifyConflict(ctx, request.ID(), commandID)
	}
	return nil
}
func (r *Repository) translateDuplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if r.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
		return application.ErrCommandApplied
	}
	return application.ErrOptimisticConflict
}
func (r *Repository) classifyConflict(ctx context.Context, id, commandID string) error {
	if err := r.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err(); err == nil {
		return application.ErrCommandApplied
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Err(); err == nil {
		return application.ErrOptimisticConflict
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrNotFound
	} else {
		return err
	}
}
func toDocument(r domain.Request) document {
	d := document{ID: r.ID(), RequesterKey: r.RequesterKey(), Source: r.Source(), CandidateIDs: r.CandidateIDs(), Status: r.Status(), ExpiresAt: r.ExpiresAt(), EndedAt: r.EndedAt(), Revision: r.Revision()}
	for _, event := range r.Events() {
		d.Events = append(d.Events, eventToDocument(event))
	}
	for _, command := range r.Commands() {
		d.Commands = append(d.Commands, commandToDocument(command))
	}
	return d
}
func toDomain(d document) (domain.Request, error) {
	state := domain.State{ID: d.ID, RequesterKey: d.RequesterKey, Source: d.Source, CandidateIDs: d.CandidateIDs, Status: d.Status, ExpiresAt: d.ExpiresAt, EndedAt: d.EndedAt, Revision: d.Revision}
	for _, e := range d.Events {
		state.Events = append(state.Events, domain.Event{Sequence: e.Sequence, CommandID: e.CommandID, ActorKey: e.ActorKey, Action: e.Action, ReasonCode: e.ReasonCode, At: e.At})
	}
	for _, c := range d.Commands {
		state.Commands = append(state.Commands, domain.AppliedCommand{ID: c.ID, Fingerprint: c.Fingerprint, Revision: c.Revision})
	}
	return domain.Rehydrate(state)
}
func eventToDocument(e domain.Event) eventDocument {
	return eventDocument{Sequence: e.Sequence, CommandID: e.CommandID, ActorKey: e.ActorKey, Action: e.Action, ReasonCode: e.ReasonCode, At: e.At}
}
func commandToDocument(c domain.AppliedCommand) commandDocument {
	return commandDocument{ID: c.ID, Fingerprint: c.Fingerprint, Revision: c.Revision}
}
