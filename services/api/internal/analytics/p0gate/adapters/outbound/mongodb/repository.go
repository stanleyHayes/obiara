package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func New(db *mongo.Database) *Repository {
	return &Repository{db.Collection("analytics_p0_gate_reports")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "snapshotid", Value: 1}, {Key: "snapshotversion", Value: 1}, {Key: "definitionversion", Value: 1}}, Options: options.Index().SetUnique(true).SetName("p0_projection_unique")}, {Keys: bson.D{{Key: "windowkey", Value: 1}, {Key: "evaluatedat", Value: -1}}, Options: options.Index().SetName("p0_window_reports")}})
	return e
}
func (r *Repository) Insert(ctx context.Context, report domain.Report) error {
	_, e := r.c.InsertOne(ctx, report)
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	var existing domain.Report
	if x := r.c.FindOne(ctx, bson.M{"snapshotid": report.SnapshotID, "snapshotversion": report.SnapshotVersion, "definitionversion": report.DefinitionVersion}).Decode(&existing); x == nil && existing.SourceWatermark == report.SourceWatermark {
		return application.ErrApplied
	}
	return application.ErrConflict
}
