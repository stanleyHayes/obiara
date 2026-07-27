package mongodb

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/p2gate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/p2gate/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func New(db *mongo.Database) *Repository {
	return &Repository{collection: db.Collection("p2_gate_link_proposals")}
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("p2_gate_command_unique")},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0).SetName("p2_gate_expiry")},
		{Keys: bson.D{{Key: "courtshipRef", Value: 1}, {Key: "reviewerRef", Value: 1}}, Options: options.Index().SetName("p2_gate_pair_reviewer")},
	})
	return err
}

func (r *Repository) Create(ctx context.Context, proposal domain.Proposal) error {
	_, err := r.collection.InsertOne(ctx, proposal)
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	var existing domain.Proposal
	if findErr := r.collection.FindOne(ctx, bson.M{"commandId": proposal.CommandID}).Decode(&existing); findErr != nil {
		return application.ErrConflict
	}
	if equalProposal(existing, proposal) {
		return application.ErrApplied
	}
	return application.ErrConflict
}

func equalProposal(a, b domain.Proposal) bool {
	if a.ID != b.ID || a.CommandID != b.CommandID || a.CourtshipRef != b.CourtshipRef ||
		a.ReviewerRef != b.ReviewerRef || a.PackVersion != b.PackVersion ||
		a.TokenRef != b.TokenRef || a.WatermarkRef != b.WatermarkRef ||
		a.OTPRequired != b.OTPRequired || a.NoForward != b.NoForward ||
		a.DeliveryProposed != b.DeliveryProposed || !a.CreatedAt.Equal(b.CreatedAt) ||
		!a.ExpiresAt.Equal(b.ExpiresAt) || len(a.Items) != len(b.Items) {
		return false
	}
	for i := range a.Items {
		if a.Items[i] != b.Items[i] {
			return false
		}
	}
	return true
}
