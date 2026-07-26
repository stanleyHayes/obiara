// Package mongodb persists reports and blocks. Reporter identity stays
// inside this store for least-exposure desk access only (Doc 09 §3).
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/application"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) reports() *mongo.Collection {
	return repository.database.Collection("reports")
}

func (repository *Repository) blocks() *mongo.Collection {
	return repository.database.Collection("blocks")
}

type reportDocument struct {
	ID         string    `bson:"_id"`
	ReporterID string    `bson:"reporterId"`
	SubjectID  string    `bson:"subjectId"`
	Category   string    `bson:"category"`
	Tier       string    `bson:"tier"`
	Surface    string    `bson:"surface"`
	ContextRef string    `bson:"contextRef,omitempty"`
	Reason     string    `bson:"reason,omitempty"`
	Status     string    `bson:"status"`
	Version    int64     `bson:"version"`
	CreatedAt  time.Time `bson:"createdAt"`
}

type blockDocument struct {
	ID        string    `bson:"_id"`
	BlockerID string    `bson:"blockerId"`
	BlockedID string    `bson:"blockedId"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.reports().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "subjectId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("reports_subject"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "tier", Value: 1}, {Key: "createdAt", Value: 1}},
			Options: options.Index().SetName("reports_queue"),
		},
	}); err != nil {
		return err
	}
	_, err := repository.blocks().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "blockerId", Value: 1}, {Key: "blockedId", Value: 1}},
		Options: options.Index().SetName("blocks_edge_unique").SetUnique(true),
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, report domain.Report) error {
	_, err := repository.reports().InsertOne(ctx, reportDocument{
		ID:         report.ID(),
		ReporterID: report.ReporterID(),
		SubjectID:  report.SubjectID(),
		Category:   string(report.Category()),
		Tier:       string(report.Tier()),
		Surface:    string(report.Surface()),
		ContextRef: report.ContextRef(),
		Reason:     report.Reason(),
		Status:     string(report.Status()),
		Version:    report.Version(),
		CreatedAt:  report.CreatedAt(),
	})
	return err
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Report, error) {
	var document reportDocument
	if err := repository.reports().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Report{}, application.ErrReportNotFound
		}
		return domain.Report{}, err
	}
	return domain.ReconstituteReport(
		document.ID, document.ReporterID, document.SubjectID,
		domain.Category(document.Category), domain.Tier(document.Tier),
		domain.Surface(document.Surface), document.ContextRef, document.Reason,
		domain.Status(document.Status), document.Version, document.CreatedAt,
	), nil
}

func blockKey(blockerID, blockedID string) string {
	return blockerID + "|" + blockedID
}

func (repository *Repository) Add(ctx context.Context, block domain.Block) error {
	_, err := repository.blocks().InsertOne(ctx, blockDocument{
		ID:        blockKey(block.BlockerID(), block.BlockedID()),
		BlockerID: block.BlockerID(),
		BlockedID: block.BlockedID(),
		CreatedAt: block.CreatedAt(),
	})
	if apimongo.IsDuplicateKey(err) {
		return application.ErrBlockExists
	}
	return err
}

func (repository *Repository) Remove(ctx context.Context, blockerID, blockedID string) error {
	result, err := repository.blocks().DeleteOne(ctx, bson.M{"_id": blockKey(blockerID, blockedID)})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return application.ErrBlockNotFound
	}
	return nil
}

func (repository *Repository) Exists(ctx context.Context, blockerID, blockedID string) (bool, error) {
	count, err := repository.blocks().CountDocuments(ctx, bson.M{"_id": blockKey(blockerID, blockedID)})
	return count > 0, err
}
