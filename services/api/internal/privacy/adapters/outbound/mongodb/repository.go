// Package mongodb persists privacy requests and legal holds.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/privacy/application"
	"github.com/stanleyHayes/obiara/services/api/internal/privacy/domain"
)

type RequestRepository struct {
	database *mongo.Database
}

func NewRequestRepository(database *mongo.Database) *RequestRepository {
	return &RequestRepository{database: database}
}

func (repository *RequestRepository) collection() *mongo.Collection {
	return repository.database.Collection("privacy_requests")
}

func (repository *RequestRepository) holds() *mongo.Collection {
	return repository.database.Collection("legal_holds")
}

type requestDocument struct {
	ID          string     `bson:"_id"`
	AccountID   string     `bson:"accountId"`
	Kind        string     `bson:"kind"`
	Status      string     `bson:"status"`
	DueAt       time.Time  `bson:"dueAt"`
	Version     int64      `bson:"version"`
	CreatedAt   time.Time  `bson:"createdAt"`
	CompletedAt *time.Time `bson:"completedAt,omitempty"`
}

type holdDocument struct {
	AccountID string     `bson:"_id"`
	Reason    string     `bson:"reason"`
	PlacedBy  string     `bson:"placedBy"`
	PlacedAt  time.Time  `bson:"placedAt"`
	LiftedAt  *time.Time `bson:"liftedAt,omitempty"`
}

func (repository *RequestRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "accountId", Value: 1}, {Key: "kind", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("privacy_open_requests"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "dueAt", Value: 1}},
			Options: options.Index().SetName("privacy_due"),
		},
	})
	return err
}

func (repository *RequestRepository) Create(ctx context.Context, request domain.PrivacyRequest) error {
	_, err := repository.collection().InsertOne(ctx, toDocument(request))
	return err
}

func (repository *RequestRepository) FindByID(ctx context.Context, id string) (domain.PrivacyRequest, error) {
	var document requestDocument
	if err := repository.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.PrivacyRequest{}, application.ErrRequestNotFound
		}
		return domain.PrivacyRequest{}, err
	}
	return toDomain(document), nil
}

func (repository *RequestRepository) FindOpenByAccountAndKind(ctx context.Context, accountID string, kind domain.Kind) (domain.PrivacyRequest, error) {
	var document requestDocument
	err := repository.collection().FindOne(ctx, bson.M{
		"accountId": accountID,
		"kind":      string(kind),
		"status":    bson.M{"$in": bson.A{string(domain.StatusRequested), string(domain.StatusProcessing), string(domain.StatusBlocked)}},
	}).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.PrivacyRequest{}, application.ErrRequestNotFound
		}
		return domain.PrivacyRequest{}, err
	}
	return toDomain(document), nil
}

func (repository *RequestRepository) Update(ctx context.Context, request domain.PrivacyRequest) error {
	document := toDocument(request)
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{"_id": document.ID, "version": document.Version - 1},
		bson.M{"$set": bson.M{"status": document.Status, "completedAt": document.CompletedAt, "version": document.Version}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrRequestNotFound
	}
	return nil
}

func (repository *RequestRepository) NextExecutable(ctx context.Context, limit int) ([]domain.PrivacyRequest, error) {
	if limit < 1 {
		limit = 25
	}
	cursor, err := repository.collection().Find(ctx,
		bson.M{"status": string(domain.StatusRequested)},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []domain.PrivacyRequest
	for cursor.Next(ctx) {
		var document requestDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		requests = append(requests, toDomain(document))
	}
	return requests, cursor.Err()
}

func (repository *RequestRepository) Place(ctx context.Context, hold domain.LegalHold) error {
	_, err := repository.holds().InsertOne(ctx, holdDocument{
		AccountID: hold.AccountID,
		Reason:    hold.Reason,
		PlacedBy:  hold.PlacedBy,
		PlacedAt:  hold.PlacedAt,
	})
	return err
}

func (repository *RequestRepository) ActiveFor(ctx context.Context, accountID string) (domain.LegalHold, error) {
	var document holdDocument
	if err := repository.holds().FindOne(ctx, bson.M{"_id": accountID, "liftedAt": nil}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.LegalHold{}, application.ErrHoldNotFound
		}
		return domain.LegalHold{}, err
	}
	return domain.LegalHold{
		AccountID: document.AccountID,
		Reason:    document.Reason,
		PlacedBy:  document.PlacedBy,
		PlacedAt:  document.PlacedAt,
		LiftedAt:  document.LiftedAt,
	}, nil
}

func (repository *RequestRepository) Lift(ctx context.Context, accountID string, now time.Time) error {
	lifted := now.UTC()
	result, err := repository.holds().UpdateOne(ctx,
		bson.M{"_id": accountID, "liftedAt": nil},
		bson.M{"$set": bson.M{"liftedAt": lifted}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrHoldNotFound
	}
	return nil
}

func toDocument(request domain.PrivacyRequest) requestDocument {
	return requestDocument{
		ID:          request.ID(),
		AccountID:   request.AccountID(),
		Kind:        string(request.Kind()),
		Status:      string(request.Status()),
		DueAt:       request.DueAt(),
		Version:     request.Version(),
		CreatedAt:   request.CreatedAt(),
		CompletedAt: request.CompletedAt(),
	}
}

func toDomain(document requestDocument) domain.PrivacyRequest {
	return domain.ReconstituteRequest(
		document.ID,
		document.AccountID,
		domain.Kind(document.Kind),
		domain.Status(document.Status),
		document.DueAt,
		document.Version,
		document.CreatedAt,
		document.CompletedAt,
	)
}
