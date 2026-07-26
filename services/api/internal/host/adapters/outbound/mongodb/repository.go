package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/host/application"
	"github.com/stanleyHayes/obiara/services/api/internal/host/domain"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database: database} }
func (repository *Repository) collection() *mongo.Collection {
	return repository.database.Collection("host_applications")
}

type proofDocument struct {
	Reference string    `bson:"reference"`
	Kind      string    `bson:"kind"`
	IssuerKey string    `bson:"issuerKey"`
	IssuedAt  time.Time `bson:"issuedAt"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

type auditDocument struct {
	Sequence   uint64    `bson:"sequence"`
	CommandID  string    `bson:"commandId"`
	Action     string    `bson:"action"`
	Reason     string    `bson:"reason,omitempty"`
	ActorKey   string    `bson:"actorKey"`
	OccurredAt time.Time `bson:"occurredAt"`
}

type applicationDocument struct {
	ID            string          `bson:"_id"`
	SubmissionID  string          `bson:"submissionId"`
	ApplicantKey  string          `bson:"applicantKey"`
	Proof         proofDocument   `bson:"proof"`
	Status        string          `bson:"status"`
	Reason        string          `bson:"reason,omitempty"`
	ProviderRef   string          `bson:"providerRef,omitempty"`
	ApprovedUntil time.Time       `bson:"approvedUntil,omitempty"`
	RecheckDueAt  time.Time       `bson:"recheckDueAt,omitempty"`
	CreatedAt     time.Time       `bson:"createdAt"`
	UpdatedAt     time.Time       `bson:"updatedAt"`
	Version       uint64          `bson:"version"`
	Audit         []auditDocument `bson:"audit"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "submissionId", Value: 1}},
		Options: options.Index().SetName("host_submission_unique").SetUnique(true),
	}); err != nil {
		return err
	}
	_, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: 1}, {Key: "recheckDueAt", Value: 1}},
		Options: options.Index().SetName("host_recheck_due"),
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, hostApplication domain.Application) (domain.Application, bool, error) {
	_, err := repository.collection().InsertOne(ctx, toDocument(hostApplication))
	if err == nil {
		return hostApplication, false, nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return domain.Application{}, false, err
	}
	var document applicationDocument
	if findErr := repository.collection().FindOne(ctx, bson.M{"submissionId": hostApplication.SubmissionID()}).Decode(&document); findErr != nil {
		return domain.Application{}, false, findErr
	}
	existing, findErr := toDomain(document)
	return existing, true, findErr
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Application, error) {
	var document applicationDocument
	if err := repository.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Application{}, application.ErrNotFound
		}
		return domain.Application{}, err
	}
	return toDomain(document)
}

func (repository *Repository) Update(ctx context.Context, hostApplication domain.Application, expected uint64, commandID string) error {
	document := toDocument(hostApplication)
	last := document.Audit[len(document.Audit)-1]
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{
			"_id": hostApplication.ID(), "version": expected,
			"audit.commandId": bson.M{"$ne": commandID},
		},
		bson.M{
			"$set": bson.M{
				"status": document.Status, "reason": document.Reason,
				"providerRef": document.ProviderRef, "approvedUntil": document.ApprovedUntil,
				"recheckDueAt": document.RecheckDueAt, "updatedAt": document.UpdatedAt,
				"version": document.Version,
			},
			"$push": bson.M{"audit": last},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrOptimisticConflict
	}
	return nil
}

func (repository *Repository) ListRecheckDue(ctx context.Context, dueBefore time.Time, limit int) ([]domain.Application, error) {
	cursor, err := repository.collection().Find(ctx, bson.M{
		"status":       string(domain.StatusApproved),
		"recheckDueAt": bson.M{"$lte": dueBefore.UTC()},
	}, options.Find().SetSort(bson.D{{Key: "recheckDueAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []applicationDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	applications := make([]domain.Application, 0, len(documents))
	for _, document := range documents {
		value, convertErr := toDomain(document)
		if convertErr != nil {
			return nil, convertErr
		}
		applications = append(applications, value)
	}
	return applications, nil
}

func toDocument(value domain.Application) applicationDocument {
	audit := make([]auditDocument, 0, len(value.Audit()))
	for _, event := range value.Audit() {
		audit = append(audit, auditDocument{
			Sequence: event.Sequence(), CommandID: event.CommandID(),
			Action: string(event.Action()), Reason: string(event.Reason()),
			ActorKey: event.ActorKey(), OccurredAt: event.OccurredAt(),
		})
	}
	return applicationDocument{
		ID: value.ID(), SubmissionID: value.SubmissionID(), ApplicantKey: value.ApplicantKey(),
		Proof: proofDocument{
			Reference: value.Proof().Reference(), Kind: string(value.Proof().Kind()),
			IssuerKey: value.Proof().IssuerKey(), IssuedAt: value.Proof().IssuedAt(),
			ExpiresAt: value.Proof().ExpiresAt(),
		},
		Status: string(value.Status()), Reason: string(value.Reason()),
		ProviderRef: value.ProviderRef(), ApprovedUntil: value.ApprovedUntil(),
		RecheckDueAt: value.RecheckDueAt(), CreatedAt: value.CreatedAt(),
		UpdatedAt: value.UpdatedAt(), Version: value.Version(), Audit: audit,
	}
}

func toDomain(document applicationDocument) (domain.Application, error) {
	proof, err := domain.NewProof(
		document.Proof.Reference, domain.InstitutionKind(document.Proof.Kind),
		document.Proof.IssuerKey, document.Proof.IssuedAt, document.Proof.ExpiresAt,
	)
	if err != nil {
		return domain.Application{}, err
	}
	audit := make([]domain.AuditEvent, 0, len(document.Audit))
	for _, stored := range document.Audit {
		event, eventErr := domain.NewAuditEvent(
			stored.Sequence, stored.CommandID, domain.Status(stored.Action),
			domain.Reason(stored.Reason), stored.ActorKey, stored.OccurredAt,
		)
		if eventErr != nil {
			return domain.Application{}, eventErr
		}
		audit = append(audit, event)
	}
	return domain.Rehydrate(
		document.ID, document.SubmissionID, document.ApplicantKey, proof,
		domain.Status(document.Status), domain.Reason(document.Reason), document.ProviderRef,
		document.ApprovedUntil, document.RecheckDueAt, document.CreatedAt,
		document.UpdatedAt, document.Version, audit,
	)
}
