package mongodb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type Repository struct {
	database *mongo.Database
	keyer    application.Keyer
}

func NewRepository(database *mongo.Database, keyer application.Keyer) *Repository {
	return &Repository{database: database, keyer: keyer}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.database.Collection("verification_admin_audit").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "caseId", Value: 1}, {Key: "occurredAt", Value: -1}},
		Options: options.Index().SetName("verification_admin_audit_case"),
	})
	return err
}

type verificationDocument struct {
	ID        string `bson:"_id"`
	AccountID string `bson:"accountId"`
	// The card itself is not stored; the write side persists only the
	// last-four mask the desk displays.
	CardMask    string     `bson:"cardMask"`
	Status      string     `bson:"status"`
	ProviderRef string     `bson:"providerRef,omitempty"`
	Reason      string     `bson:"reason,omitempty"`
	DateOfBirth time.Time  `bson:"dateOfBirth"`
	Version     int64      `bson:"version"`
	CreatedAt   time.Time  `bson:"createdAt"`
	DecidedAt   *time.Time `bson:"decidedAt,omitempty"`
}

type commandDocument struct {
	ID          string `bson:"_id"`
	RequestHash string `bson:"requestHash"`
	CaseID      string `bson:"caseId"`
	Outcome     string `bson:"outcome"`
}

func (repository *Repository) ListQueued(ctx context.Context, limit int) ([]application.CaseSummary, error) {
	cursor, err := repository.database.Collection("identity_verifications").Find(
		ctx, bson.M{"status": "queued_manual"},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}, {Key: "_id", Value: 1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	result := make([]application.CaseSummary, 0)
	for cursor.Next(ctx) {
		var document verificationDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		summary, err := repository.summary(document)
		if err != nil {
			return nil, err
		}
		result = append(result, summary)
	}
	return result, cursor.Err()
}

func (repository *Repository) Detail(ctx context.Context, caseID string) (application.CaseDetail, error) {
	document, err := repository.find(ctx, caseID)
	if err != nil {
		return application.CaseDetail{}, err
	}
	return repository.detail(document)
}

func (repository *Repository) AccessEvidence(ctx context.Context, access application.EvidenceAccess) (application.Evidence, error) {
	document, err := repository.find(ctx, access.CaseID)
	if err != nil {
		return application.Evidence{}, err
	}
	audit := bson.M{
		"caseId": access.CaseID, "actorId": access.ActorID, "eventType": "evidence_access",
		"purpose": access.Purpose, "reason": access.Reason,
		"correlationId": access.CorrelationID, "occurredAt": access.OccurredAt,
	}
	if _, err := repository.database.Collection("verification_admin_audit").InsertOne(ctx, audit); err != nil {
		return application.Evidence{}, err
	}
	return application.Evidence{
		CaseID: document.ID, MaskedCard: document.CardMask,
		AgeBand:        ageBand(document.DateOfBirth, access.OccurredAt),
		ProviderStatus: reasonCode(document.Reason),
	}, nil
}

func (repository *Repository) Decide(ctx context.Context, command application.DecisionCommand) (application.DecisionResult, error) {
	requestHash := decisionHash(command)
	if replay, found, err := repository.replay(ctx, command.IdempotencyKey, requestHash); found || err != nil {
		return replay, err
	}
	status := "rejected"
	if command.Outcome == application.OutcomeApprove {
		status = "approved"
	}
	session, err := repository.database.Client().StartSession()
	if err != nil {
		return application.DecisionResult{}, err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(transactionContext context.Context) (any, error) {
		var updated verificationDocument
		err := repository.database.Collection("identity_verifications").FindOneAndUpdate(
			transactionContext,
			bson.M{"_id": command.CaseID, "status": "queued_manual", "version": command.ExpectedVersion},
			bson.M{"$set": bson.M{
				"status": status, "reason": command.Reason,
				"providerRef": "manual:" + command.ActorID, "decidedAt": command.OccurredAt,
			}, "$inc": bson.M{"version": 1}},
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&updated)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.classifyDecisionMiss(transactionContext, command)
		}
		if err != nil {
			return nil, err
		}
		if _, err := repository.database.Collection("verification_admin_commands").InsertOne(transactionContext, commandDocument{
			ID: command.IdempotencyKey, RequestHash: requestHash, CaseID: command.CaseID, Outcome: string(command.Outcome),
		}); err != nil {
			return nil, err
		}
		_, err = repository.database.Collection("verification_admin_audit").InsertOne(transactionContext, bson.M{
			"caseId": command.CaseID, "actorId": command.ActorID, "eventType": "decision",
			"outcome": command.Outcome, "reason": command.Reason,
			"idempotencyKey": command.IdempotencyKey, "correlationId": command.CorrelationID,
			"occurredAt": command.OccurredAt, "fromVersion": command.ExpectedVersion,
			"toVersion": command.ExpectedVersion + 1,
		})
		return nil, err
	})
	if err != nil {
		if replay, found, replayErr := repository.replay(ctx, command.IdempotencyKey, requestHash); found || replayErr != nil {
			return replay, replayErr
		}
		return application.DecisionResult{}, err
	}
	detail, err := repository.Detail(ctx, command.CaseID)
	return application.DecisionResult{Case: detail, Outcome: command.Outcome}, err
}

func (repository *Repository) replay(ctx context.Context, key, requestHash string) (application.DecisionResult, bool, error) {
	var command commandDocument
	err := repository.database.Collection("verification_admin_commands").FindOne(ctx, bson.M{"_id": key}).Decode(&command)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return application.DecisionResult{}, false, nil
	}
	if err != nil {
		return application.DecisionResult{}, true, err
	}
	if command.RequestHash != requestHash {
		return application.DecisionResult{}, true, application.ErrIdempotencyConflict
	}
	detail, err := repository.Detail(ctx, command.CaseID)
	return application.DecisionResult{
		Case: detail, Outcome: application.Outcome(command.Outcome), Replayed: true,
	}, true, err
}

func (repository *Repository) classifyDecisionMiss(ctx context.Context, command application.DecisionCommand) error {
	document, err := repository.find(ctx, command.CaseID)
	if err != nil {
		return err
	}
	if document.Status != "queued_manual" {
		return application.ErrCaseClosed
	}
	return application.ErrStaleCase
}

func (repository *Repository) find(ctx context.Context, id string) (verificationDocument, error) {
	var document verificationDocument
	err := repository.database.Collection("identity_verifications").FindOne(ctx, bson.M{"_id": id}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return verificationDocument{}, application.ErrCaseNotFound
	}
	return document, err
}

func (repository *Repository) summary(document verificationDocument) (application.CaseSummary, error) {
	key, err := repository.keyer.Key("verification-admin:subject", document.AccountID)
	if err != nil {
		return application.CaseSummary{}, err
	}
	return application.CaseSummary{
		ID: document.ID, SubjectRef: "member_" + key[:12],
		ReasonCode: reasonCode(document.Reason), SubmittedAt: document.CreatedAt,
		Version: document.Version,
	}, nil
}

func (repository *Repository) detail(document verificationDocument) (application.CaseDetail, error) {
	summary, err := repository.summary(document)
	return application.CaseDetail{CaseSummary: summary, Status: document.Status}, err
}

func reasonCode(reason string) string {
	switch reason {
	case "provider unavailable":
		return "provider_outage"
	case "provider uncertain":
		return "provider_uncertain"
	default:
		return "manual_review"
	}
}

func ageBand(dateOfBirth, at time.Time) string {
	age := at.Year() - dateOfBirth.Year()
	if at.YearDay() < dateOfBirth.YearDay() {
		age--
	}
	switch {
	case age < 18:
		return "under_18"
	case age < 25:
		return "18_24"
	case age < 35:
		return "25_34"
	case age < 50:
		return "35_49"
	default:
		return "50_plus"
	}
}

func decisionHash(command application.DecisionCommand) string {
	value := fmt.Sprintf("%s\x00%s\x00%s\x00%d", command.CaseID, command.Outcome, command.Reason, command.ExpectedVersion)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
