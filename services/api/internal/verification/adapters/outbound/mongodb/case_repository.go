// Package mongodb persists verification cases. Documents hold the minimal
// proof set (reference and decision), never card images or biometric media
// (data-classification C4 handling).
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/application"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

type CaseRepository struct {
	database *mongo.Database
}

func NewCaseRepository(database *mongo.Database) *CaseRepository {
	return &CaseRepository{database: database}
}

func (repository *CaseRepository) collection() *mongo.Collection {
	return repository.database.Collection("identity_verifications")
}

type caseDocument struct {
	ID          string     `bson:"_id"`
	AccountID   string     `bson:"accountId"`
	CardKey     string     `bson:"cardKey"`
	CardMask    string     `bson:"cardMask"`
	Status      string     `bson:"status"`
	ProviderRef string     `bson:"providerRef,omitempty"`
	Reason      string     `bson:"reason,omitempty"`
	DateOfBirth time.Time  `bson:"dateOfBirth"`
	Version     int64      `bson:"version"`
	CreatedAt   time.Time  `bson:"createdAt"`
	DecidedAt   *time.Time `bson:"decidedAt,omitempty"`
}

func (repository *CaseRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "accountId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("verifications_account"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: 1}},
			Options: options.Index().SetName("verifications_queue"),
		},
		{
			// One approved case per card, across every account. This is the
			// real enforcement of "one active account per verified
			// identity": the service checks first for a clear error, but two
			// concurrent submissions of the same card would both pass that
			// check and only one may commit.
			//
			// Partial, so that rejected and abandoned attempts do not claim
			// an identity — a member whose first submission failed must be
			// able to try again.
			Keys: bson.D{{Key: "cardKey", Value: 1}},
			Options: options.Index().SetName("verifications_identity_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"status": string(domain.StatusApproved)}),
		},
	})
	return err
}

// ApprovedAccountByCardKey reports which account, if any, already holds this
// identity.
func (repository *CaseRepository) ApprovedAccountByCardKey(ctx context.Context, cardKey string) (string, error) {
	var document caseDocument
	err := repository.collection().FindOne(ctx, bson.M{
		"cardKey": cardKey,
		"status":  string(domain.StatusApproved),
	}).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", application.ErrCaseNotFound
		}
		return "", err
	}
	return document.AccountID, nil
}

// LatestByAccount returns the account's newest case. Served by the
// verifications_account index, which already sorts createdAt descending.
func (repository *CaseRepository) LatestByAccount(ctx context.Context, accountID string) (domain.VerificationCase, error) {
	var document caseDocument
	err := repository.collection().FindOne(
		ctx,
		bson.M{"accountId": accountID},
		options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.VerificationCase{}, application.ErrCaseNotFound
		}
		return domain.VerificationCase{}, err
	}
	return toDomain(document), nil
}

func (repository *CaseRepository) Create(ctx context.Context, verificationCase domain.VerificationCase) error {
	_, err := repository.collection().InsertOne(ctx, toDocument(verificationCase))
	return err
}

func (repository *CaseRepository) FindByID(ctx context.Context, id string) (domain.VerificationCase, error) {
	var document caseDocument
	if err := repository.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.VerificationCase{}, application.ErrCaseNotFound
		}
		return domain.VerificationCase{}, err
	}
	return toDomain(document), nil
}

// Update applies optimistic concurrency on the case version.
func (repository *CaseRepository) Update(ctx context.Context, verificationCase domain.VerificationCase) error {
	document := toDocument(verificationCase)
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{"_id": document.ID, "version": document.Version - 1},
		bson.M{"$set": bson.M{
			"status":      document.Status,
			"providerRef": document.ProviderRef,
			"reason":      document.Reason,
			"decidedAt":   document.DecidedAt,
			"version":     document.Version,
		}})
	if err != nil {
		// The partial unique index rejects a second approval of the same
		// card. Two concurrent submissions both clear the service's
		// pre-check, so this is where "one active account per verified
		// identity" is actually decided.
		if apimongo.IsDuplicateKey(err) {
			return application.ErrIdentityAlreadyVerified
		}
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrCaseNotFound
	}
	return nil
}

func (repository *CaseRepository) NextQueued(ctx context.Context, limit int) ([]domain.VerificationCase, error) {
	if limit < 1 {
		limit = 50
	}
	cursor, err := repository.collection().Find(ctx,
		bson.M{"status": string(domain.StatusQueuedManual)},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cases []domain.VerificationCase
	for cursor.Next(ctx) {
		var document caseDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		cases = append(cases, toDomain(document))
	}
	return cases, cursor.Err()
}

func toDocument(verificationCase domain.VerificationCase) caseDocument {
	return caseDocument{
		ID:          verificationCase.ID(),
		AccountID:   verificationCase.AccountID(),
		CardKey:     verificationCase.CardKey(),
		CardMask:    verificationCase.CardMask(),
		Status:      string(verificationCase.Status()),
		ProviderRef: verificationCase.ProviderRef(),
		Reason:      verificationCase.Reason(),
		DateOfBirth: verificationCase.DateOfBirth(),
		Version:     verificationCase.Version(),
		CreatedAt:   verificationCase.CreatedAt(),
		DecidedAt:   verificationCase.DecidedAt(),
	}
}

func toDomain(document caseDocument) domain.VerificationCase {
	return domain.ReconstituteCase(
		document.ID,
		document.AccountID,
		document.CardKey,
		document.CardMask,
		domain.CaseStatus(document.Status),
		document.ProviderRef,
		document.Reason,
		document.DateOfBirth,
		document.Version,
		document.CreatedAt,
		document.DecidedAt,
	)
}
