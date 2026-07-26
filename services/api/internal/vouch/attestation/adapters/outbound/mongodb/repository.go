package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/vouch/attestation/domain"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("vouch_attestations")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetName("vouch_attestation_command").SetUnique(true)},
		{
			Keys: bson.D{{Key: "subjectKey", Value: 1}, {Key: "voucherKey", Value: 1}, {Key: "scope.kind", Value: 1}, {Key: "scope.key", Value: 1}},
			Options: options.Index().SetName("vouch_attestation_active_scope").SetUnique(true).
				SetPartialFilterExpression(bson.M{"status": bson.M{"$in": bson.A{domain.StatusAwaitingConsent, domain.StatusActive}}}),
		},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}, {Key: "status", Value: 1}}, Options: options.Index().SetName("vouch_attestation_expiry")},
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
type provenanceDocument struct {
	VoucherKey    string    `bson:"voucherKey"`
	PolicyVersion string    `bson:"policyVersion"`
	ConsentedAt   time.Time `bson:"consentedAt"`
}
type document struct {
	ID            string              `bson:"_id"`
	SubjectKey    string              `bson:"subjectKey"`
	VoucherKey    string              `bson:"voucherKey"`
	Scope         domain.SubjectScope `bson:"scope"`
	StakeUnits    uint8               `bson:"stakeUnits"`
	PolicyVersion string              `bson:"policyVersion"`
	Status        domain.Status       `bson:"status"`
	ExpiresAt     time.Time           `bson:"expiresAt"`
	Provenance    *provenanceDocument `bson:"provenance,omitempty"`
	EndedAt       *time.Time          `bson:"endedAt,omitempty"`
	Revision      uint64              `bson:"revision"`
	Events        []eventDocument     `bson:"events"`
	Commands      []commandDocument   `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, a domain.Attestation) error {
	_, err := r.collection.InsertOne(ctx, toDocument(a))
	return r.translateDuplicate(ctx, a.Commands()[0].ID, err)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Attestation, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Attestation, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, filter bson.M) (domain.Attestation, error) {
	var stored document
	if err := r.collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Attestation{}, application.ErrNotFound
		}
		return domain.Attestation{}, err
	}
	return toDomain(stored)
}

// Append only pushes the new audit/command pair. Earlier history and immutable
// subject, scope, stake, policy, and expiry fields are never rewritten.
func (r *Repository) Append(ctx context.Context, a domain.Attestation, expected uint64, commandID string) error {
	events, commands := a.Events(), a.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalidAttestation
	}
	update := bson.M{
		"$set": bson.M{
			"status": a.Status(), "revision": a.Revision(),
			"provenance": provenanceToDocument(a.Provenance()), "endedAt": a.EndedAt(),
		},
		"$push": bson.M{
			"events":   eventToDocument(events[len(events)-1]),
			"commands": commandToDocument(commands[len(commands)-1]),
		},
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": a.ID(), "revision": expected}, update)
	if err != nil {
		return r.translateDuplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		return r.classifyConflict(ctx, a.ID(), commandID)
	}
	return nil
}
func (r *Repository) translateDuplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if findErr := r.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err(); findErr == nil {
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
func toDocument(a domain.Attestation) document {
	stored := document{
		ID: a.ID(), SubjectKey: a.SubjectKey(), VoucherKey: a.VoucherKey(), Scope: a.Scope(),
		StakeUnits: a.StakeUnits(), PolicyVersion: a.PolicyVersion(), Status: a.Status(),
		ExpiresAt: a.ExpiresAt(), Provenance: provenanceToDocument(a.Provenance()),
		EndedAt: a.EndedAt(), Revision: a.Revision(),
	}
	for _, event := range a.Events() {
		stored.Events = append(stored.Events, eventToDocument(event))
	}
	for _, command := range a.Commands() {
		stored.Commands = append(stored.Commands, commandToDocument(command))
	}
	return stored
}
func toDomain(stored document) (domain.Attestation, error) {
	state := domain.State{
		ID: stored.ID, SubjectKey: stored.SubjectKey, VoucherKey: stored.VoucherKey, Scope: stored.Scope,
		StakeUnits: stored.StakeUnits, PolicyVersion: stored.PolicyVersion, Status: stored.Status,
		ExpiresAt: stored.ExpiresAt, Provenance: provenanceToDomain(stored.Provenance),
		EndedAt: stored.EndedAt, Revision: stored.Revision,
	}
	for _, event := range stored.Events {
		state.Events = append(state.Events, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey,
			Action: event.Action, ReasonCode: event.ReasonCode, At: event.At,
		})
	}
	for _, command := range stored.Commands {
		state.Commands = append(state.Commands, domain.AppliedCommand{
			ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
		})
	}
	return domain.Rehydrate(state)
}
func eventToDocument(event domain.Event) eventDocument {
	return eventDocument{Sequence: event.Sequence, CommandID: event.CommandID, ActorKey: event.ActorKey, Action: event.Action, ReasonCode: event.ReasonCode, At: event.At}
}
func commandToDocument(command domain.AppliedCommand) commandDocument {
	return commandDocument{ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision}
}
func provenanceToDocument(value *domain.Provenance) *provenanceDocument {
	if value == nil {
		return nil
	}
	return &provenanceDocument{VoucherKey: value.VoucherKey, PolicyVersion: value.PolicyVersion, ConsentedAt: value.ConsentedAt}
}
func provenanceToDomain(value *provenanceDocument) *domain.Provenance {
	if value == nil {
		return nil
	}
	return &domain.Provenance{VoucherKey: value.VoucherKey, PolicyVersion: value.PolicyVersion, ConsentedAt: value.ConsentedAt}
}
