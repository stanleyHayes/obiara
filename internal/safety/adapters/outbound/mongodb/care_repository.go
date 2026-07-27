package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

// CareRepository persists care cases and quietening windows.
type CareRepository struct {
	database *mongo.Database
}

func NewCareRepository(database *mongo.Database) *CareRepository {
	return &CareRepository{database: database}
}

func (repository *CareRepository) cases() *mongo.Collection {
	return repository.database.Collection("care_cases")
}

func (repository *CareRepository) quietening() *mongo.Collection {
	return repository.database.Collection("care_quietening")
}

type careDocument struct {
	ID         string     `bson:"_id"`
	SubjectID  string     `bson:"subjectId"`
	Signal     string     `bson:"signal"`
	Status     string     `bson:"status"`
	Scripts    []string   `bson:"scripts,omitempty"`
	Version    int64      `bson:"version"`
	CreatedAt  time.Time  `bson:"createdAt"`
	ResolvedAt *time.Time `bson:"resolvedAt,omitempty"`
}

func (repository *CareRepository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.cases().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: 1}},
			Options: options.Index().SetName("care_open_fifo"),
		},
		{
			Keys:    bson.D{{Key: "subjectId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("care_subject"),
		},
	}); err != nil {
		return err
	}
	_, err := repository.quietening().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "until", Value: 1}},
		Options: options.Index().SetName("care_quietening_ttl").SetExpireAfterSeconds(0),
	})
	return err
}

func (repository *CareRepository) Create(ctx context.Context, careCase domain.CareCase) error {
	_, err := repository.cases().InsertOne(ctx, toCareDocument(careCase))
	return err
}

func (repository *CareRepository) FindByID(ctx context.Context, id string) (domain.CareCase, error) {
	var document careDocument
	if err := repository.cases().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.CareCase{}, application.ErrCaseNotFound
		}
		return domain.CareCase{}, err
	}
	return toCareDomain(document), nil
}

func (repository *CareRepository) Update(ctx context.Context, careCase domain.CareCase) error {
	document := toCareDocument(careCase)
	result, err := repository.cases().UpdateOne(ctx,
		bson.M{"_id": document.ID, "version": document.Version - 1},
		bson.M{"$set": bson.M{"status": document.Status, "scripts": document.Scripts, "resolvedAt": document.ResolvedAt, "version": document.Version}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrCaseNotFound
	}
	return nil
}

func (repository *CareRepository) NextOpen(ctx context.Context, limit int) ([]domain.CareCase, error) {
	if limit < 1 {
		limit = 50
	}
	cursor, err := repository.cases().Find(ctx,
		bson.M{"status": bson.M{"$in": bson.A{string(domain.CareOpen), string(domain.CareEngaged)}}},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cases []domain.CareCase
	for cursor.Next(ctx) {
		var document careDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		cases = append(cases, toCareDomain(document))
	}
	return cases, cursor.Err()
}

// Set records a quietening window (upsert: a fresh closure refreshes it).
func (repository *CareRepository) Set(ctx context.Context, subjectID string, until time.Time) error {
	_, err := repository.quietening().UpdateOne(ctx,
		bson.M{"_id": subjectID},
		bson.M{"$set": bson.M{"until": until.UTC()}},
		options.UpdateOne().SetUpsert(true))
	return err
}

// QuietUntil returns the quietening deadline when one is active.
func (repository *CareRepository) QuietUntil(ctx context.Context, subjectID string, now time.Time) (bool, error) {
	var document struct {
		Until time.Time `bson:"until"`
	}
	err := repository.quietening().FindOne(ctx, bson.M{"_id": subjectID}).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return now.UTC().Before(document.Until), nil
}

func toCareDocument(careCase domain.CareCase) careDocument {
	scripts := make([]string, 0, len(careCase.Scripts()))
	for _, script := range careCase.Scripts() {
		scripts = append(scripts, string(script))
	}
	return careDocument{
		ID: careCase.ID(), SubjectID: careCase.SubjectID(), Signal: string(careCase.Signal()),
		Status: string(careCase.Status()), Scripts: scripts, Version: careCase.Version(),
		CreatedAt: careCase.CreatedAt(), ResolvedAt: careCase.ResolvedAt(),
	}
}

func toCareDomain(document careDocument) domain.CareCase {
	scripts := make([]domain.ScriptKey, 0, len(document.Scripts))
	for _, script := range document.Scripts {
		scripts = append(scripts, domain.ScriptKey(script))
	}
	return domain.ReconstituteCareCase(document.ID, document.SubjectID, domain.Signal(document.Signal), domain.CareStatus(document.Status), scripts, document.Version, document.CreatedAt, document.ResolvedAt)
}
