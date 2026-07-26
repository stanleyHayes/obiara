package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

// CaseRepository persists T&S cases (kept separate from the report/block
// Repository because both contexts use short method names).
type CaseRepository struct {
	database *mongo.Database
}

func NewCaseRepository(database *mongo.Database) *CaseRepository {
	return &CaseRepository{database: database}
}

type caseDocument struct {
	ID         string     `bson:"_id"`
	ReportID   string     `bson:"reportId"`
	SubjectID  string     `bson:"subjectId"`
	Tier       string     `bson:"tier"`
	Queue      string     `bson:"queue"`
	SLADueAt   time.Time  `bson:"slaDueAt"`
	Status     string     `bson:"status"`
	AssignedTo string     `bson:"assignedTo,omitempty"`
	Version    int64      `bson:"version"`
	CreatedAt  time.Time  `bson:"createdAt"`
	ResolvedAt *time.Time `bson:"resolvedAt,omitempty"`
}

func (repository *CaseRepository) cases() *mongo.Collection {
	return repository.database.Collection("safety_cases")
}

func (repository *CaseRepository) EnsureCaseIndexes(ctx context.Context) error {
	_, err := repository.cases().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// One case per report: queue building is replay-safe.
			Keys:    bson.D{{Key: "reportId", Value: 1}},
			Options: options.Index().SetName("cases_report_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "queue", Value: 1}, {Key: "status", Value: 1}, {Key: "slaDueAt", Value: 1}},
			Options: options.Index().SetName("cases_queue_sla"),
		},
	})
	return err
}

func (repository *CaseRepository) Create(ctx context.Context, safetyCase domain.Case) error {
	_, err := repository.cases().InsertOne(ctx, toCaseDocument(safetyCase))
	if apimongo.IsDuplicateKey(err) {
		return domain.ErrReportAlreadyQueued
	}
	return err
}

func (repository *CaseRepository) FindByID(ctx context.Context, id string) (domain.Case, error) {
	var document caseDocument
	if err := repository.cases().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Case{}, application.ErrCaseNotFound
		}
		return domain.Case{}, err
	}
	return toCaseDomain(document), nil
}

func (repository *CaseRepository) Update(ctx context.Context, safetyCase domain.Case) error {
	document := toCaseDocument(safetyCase)
	result, err := repository.cases().UpdateOne(ctx,
		bson.M{"_id": document.ID, "version": document.Version - 1},
		bson.M{"$set": bson.M{
			"status": document.Status, "assignedTo": document.AssignedTo,
			"resolvedAt": document.ResolvedAt, "version": document.Version,
		}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrCaseNotFound
	}
	return nil
}

func (repository *CaseRepository) NextQueued(ctx context.Context, queue domain.Queue, limit int) ([]domain.Case, error) {
	if limit < 1 {
		limit = 50
	}
	cursor, err := repository.cases().Find(ctx,
		bson.M{"queue": string(queue), "status": bson.M{"$in": bson.A{string(domain.CaseQueued), string(domain.CaseInReview)}}},
		options.Find().SetSort(bson.D{{Key: "slaDueAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cases []domain.Case
	for cursor.Next(ctx) {
		var document caseDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		cases = append(cases, toCaseDomain(document))
	}
	return cases, cursor.Err()
}

func (repository *CaseRepository) CountBreached(ctx context.Context, now time.Time) (int, error) {
	count, err := repository.cases().CountDocuments(ctx, bson.M{
		"status":   bson.M{"$ne": string(domain.CaseResolved)},
		"slaDueAt": bson.M{"$lt": now.UTC()},
	})
	return int(count), err
}

func toCaseDocument(safetyCase domain.Case) caseDocument {
	return caseDocument{
		ID:         safetyCase.ID(),
		ReportID:   safetyCase.ReportID(),
		SubjectID:  safetyCase.SubjectID(),
		Tier:       string(safetyCase.Tier()),
		Queue:      string(safetyCase.Queue()),
		SLADueAt:   safetyCase.SLADueAt(),
		Status:     string(safetyCase.Status()),
		AssignedTo: safetyCase.AssignedTo(),
		Version:    safetyCase.Version(),
		CreatedAt:  safetyCase.CreatedAt(),
		ResolvedAt: safetyCase.ResolvedAt(),
	}
}

func toCaseDomain(document caseDocument) domain.Case {
	return domain.ReconstituteCase(
		document.ID, document.ReportID, document.SubjectID,
		domain.Tier(document.Tier), domain.Queue(document.Queue),
		document.SLADueAt, domain.CaseStatus(document.Status),
		document.AssignedTo, document.Version, document.CreatedAt, document.ResolvedAt,
	)
}
