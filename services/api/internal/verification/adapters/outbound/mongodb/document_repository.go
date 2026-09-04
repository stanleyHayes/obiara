// Documents hold encrypted Ghana Card photographs awaiting human review.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
)

type DocumentRepository struct {
	database *mongo.Database
}

func NewDocumentRepository(database *mongo.Database) *DocumentRepository {
	return &DocumentRepository{database: database}
}

func (repository *DocumentRepository) collection() *mongo.Collection {
	return repository.database.Collection("identity_documents")
}

type documentRecord struct {
	ID         string    `bson:"_id"`
	CaseID     string    `bson:"caseId"`
	SubjectKey string    `bson:"subjectKey"`
	Side       string    `bson:"side"`
	MediaType  string    `bson:"mediaType"`
	Ciphertext []byte    `bson:"ciphertext"`
	Nonce      []byte    `bson:"nonce"`
	CreatedAt  time.Time `bson:"createdAt"`
}

func (repository *DocumentRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		// Reviewers open a case and want both of its sides. No TTL here on
		// purpose: a queue can take days, and an image that expired mid-review
		// would leave a member unverifiable with nothing explaining why.
		Keys:    bson.D{{Key: "caseId", Value: 1}, {Key: "side", Value: 1}},
		Options: options.Index().SetName("identity_documents_case"),
	})
	return err
}

func (repository *DocumentRepository) SaveDocument(ctx context.Context, document application.Document) error {
	_, err := repository.collection().InsertOne(ctx, documentRecord{
		ID: document.ID, CaseID: document.CaseID, SubjectKey: document.SubjectKey,
		Side: document.Side, MediaType: document.MediaType,
		Ciphertext: document.Ciphertext, Nonce: document.Nonce, CreatedAt: document.CreatedAt,
	})
	return err
}

func (repository *DocumentRepository) DocumentsForCase(ctx context.Context, caseID string) ([]application.Document, error) {
	cursor, err := repository.collection().Find(ctx,
		bson.M{"caseId": caseID},
		options.Find().SetSort(bson.D{{Key: "side", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []application.Document
	for cursor.Next(ctx) {
		var record documentRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, err
		}
		documents = append(documents, application.Document{
			ID: record.ID, CaseID: record.CaseID, SubjectKey: record.SubjectKey,
			Side: record.Side, MediaType: record.MediaType,
			Ciphertext: record.Ciphertext, Nonce: record.Nonce, CreatedAt: record.CreatedAt,
		})
	}
	return documents, cursor.Err()
}
