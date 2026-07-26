package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

// AccessAuditStore persists immutable evidence-access records.
type AccessAuditStore struct {
	database *mongo.Database
}

func NewAccessAuditStore(database *mongo.Database) *AccessAuditStore {
	return &AccessAuditStore{database: database}
}

func (store *AccessAuditStore) collection() *mongo.Collection {
	return store.database.Collection("evidence_access_log")
}

type accessDocument struct {
	ID         string    `bson:"_id"`
	CaseID     string    `bson:"caseId"`
	AgentID    string    `bson:"agentId"`
	Purpose    string    `bson:"purpose"`
	AccessedAt time.Time `bson:"accessedAt"`
}

func (store *AccessAuditStore) Append(ctx context.Context, record domain.AccessRecord) error {
	_, err := store.collection().InsertOne(ctx, accessDocument{
		ID:         record.ID,
		CaseID:     record.CaseID,
		AgentID:    record.AgentID,
		Purpose:    string(record.Purpose),
		AccessedAt: record.AccessedAt.UTC(),
	})
	return err
}

func (store *AccessAuditStore) CountForCase(ctx context.Context, caseID string) (int, error) {
	count, err := store.collection().CountDocuments(ctx, bson.M{"caseId": caseID})
	return int(count), err
}

// CountForAgent supports insider-access reviews (plan §15).
func (store *AccessAuditStore) CountForAgent(ctx context.Context, agentID string, since time.Time) (int, error) {
	count, err := store.collection().CountDocuments(ctx, bson.M{"agentId": agentID, "accessedAt": bson.M{"$gte": since.UTC()}})
	return int(count), err
}
