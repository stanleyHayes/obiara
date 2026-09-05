package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

// ActionLogStore persists the append-only action audit log and the
// device/biometric blocklist.
type ActionLogStore struct {
	database *mongo.Database
}

func NewActionLogStore(database *mongo.Database) *ActionLogStore {
	return &ActionLogStore{database: database}
}

func (store *ActionLogStore) actions() *mongo.Collection {
	return store.database.Collection("ts_action_log")
}

type actionDocument struct {
	ID        string    `bson:"_id"`
	CommandID string    `bson:"commandId"`
	CaseID    string    `bson:"caseId"`
	SubjectID string    `bson:"subjectId"`
	Action    string    `bson:"action"`
	ActorID   string    `bson:"actorId"`
	Priors    int       `bson:"priors"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (store *ActionLogStore) Append(ctx context.Context, record domain.ActionRecord) error {
	_, err := store.actions().InsertOne(ctx, actionDocument{
		ID:        record.ID,
		CommandID: record.CommandID,
		CaseID:    record.CaseID,
		SubjectID: record.SubjectID,
		Action:    string(record.Action),
		ActorID:   record.ActorID,
		Priors:    record.Priors,
		CreatedAt: record.CreatedAt.UTC(),
	})
	return err
}

// EnsureIndexes makes the command id unique, so two operators clicking at
// once cannot both write an action for the same request.
func (store *ActionLogStore) EnsureIndexes(ctx context.Context) error {
	_, err := store.actions().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "commandId", Value: 1}},
		Options: options.Index().SetName("ts_action_command_unique").SetUnique(true),
	})
	return err
}

// AppliedCommand reports whether this request already took effect.
func (store *ActionLogStore) AppliedCommand(ctx context.Context, commandID string) (bool, error) {
	err := store.actions().FindOne(ctx, bson.M{"commandId": commandID}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return err == nil, err
}

func (store *ActionLogStore) CountForSubject(ctx context.Context, subjectID string) (int, error) {
	count, err := store.actions().CountDocuments(ctx, bson.M{"subjectId": subjectID})
	return int(count), err
}

// Blocklist writes a device/biometric blocklist entry (Doc 09 §2: Tier-A
// bans propagate across alternate accounts via device/biometric signals).
func (store *ActionLogStore) Blocklist(ctx context.Context, subjectID, reason string, at time.Time) error {
	_, err := store.database.Collection("device_risk").UpdateOne(ctx,
		bson.M{"_id": subjectID},
		bson.M{"$set": bson.M{"blocked": true, "reason": reason, "blockedAt": at.UTC()}},
		options.UpdateOne().SetUpsert(true))
	return err
}
