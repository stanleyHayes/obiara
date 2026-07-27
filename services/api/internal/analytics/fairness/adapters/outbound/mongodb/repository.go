package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/application"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("analytics_fairness_reports")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "quarterKey", Value: 1}, {Key: "definitionVersion", Value: 1}}, Options: options.Index().SetUnique(true).SetName("fairness_quarter_definition_unique")},
		{Keys: bson.D{{Key: "evaluatedAt", Value: -1}}, Options: options.Index().SetName("fairness_evaluated_at")},
	})
	return err
}
func (r *Repository) Insert(ctx context.Context, report domain.Report) error {
	if _, err := domain.RehydrateReport(report); err != nil {
		return err
	}
	_, err := r.collection.InsertOne(ctx, report)
	if mongo.IsDuplicateKeyError(err) {
		return application.ErrApplied
	}
	return err
}
func (r *Repository) Find(ctx context.Context, quarter string, definition uint64) (domain.Report, error) {
	var report domain.Report
	if err := r.collection.FindOne(ctx, bson.M{"quarterKey": quarter, "definitionVersion": definition}).Decode(&report); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Report{}, application.ErrNotFound
		}
		return domain.Report{}, err
	}
	return domain.RehydrateReport(report)
}
