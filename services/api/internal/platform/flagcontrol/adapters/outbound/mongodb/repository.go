package mongodb

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/application"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct {
	database          *mongo.Database
	proposals, audits *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db, db.Collection("platform_flag_control_proposals"), db.Collection("platform_flag_control_audits")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.proposals.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("flag_control_command_unique")},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "expiresAt", Value: 1}}, Options: options.Index().SetName("flag_control_expiry")},
	}); err != nil {
		return err
	}
	_, err := r.audits.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "proposalId", Value: 1}, {Key: "version", Value: 1}, {Key: "kind", Value: 1}}, Options: options.Index().SetUnique(true).SetName("flag_control_audit_unique")},
		{Keys: bson.D{{Key: "at", Value: 1}}, Options: options.Index().SetName("flag_control_audit_time")},
	})
	return err
}

type proposalDoc struct {
	ID          string             `bson:"_id"`
	CommandID   string             `bson:"commandId"`
	Fingerprint string             `bson:"fingerprint"`
	ProposerKey string             `bson:"proposerKey"`
	ApproverKey string             `bson:"approverKey,omitempty"`
	Capability  domain.Capability  `bson:"capability"`
	Environment domain.Environment `bson:"environment"`
	Market      domain.Market      `bson:"market"`
	Action      domain.Action      `bson:"action"`
	Reason      domain.Reason      `bson:"reason"`
	Status      domain.Status      `bson:"status"`
	Version     uint64             `bson:"version"`
	CreatedAt   time.Time          `bson:"createdAt"`
	ExpiresAt   time.Time          `bson:"expiresAt"`
	ApprovedAt  *time.Time         `bson:"approvedAt,omitempty"`
	AppliedAt   *time.Time         `bson:"appliedAt,omitempty"`
}
type auditDoc struct {
	ID         string           `bson:"_id"`
	ProposalID string           `bson:"proposalId"`
	ActorKey   string           `bson:"actorKey"`
	Kind       domain.AuditKind `bson:"kind"`
	Version    uint64           `bson:"version"`
	At         time.Time        `bson:"at"`
}

func (r *Repository) CreateWithAudit(ctx context.Context, p domain.Proposal, a domain.Audit) error {
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		if _, e := r.proposals.InsertOne(tx, toDoc(p)); e != nil {
			return e
		}
		_, e := r.audits.InsertOne(tx, toAuditDoc(a))
		return e
	})
	if mongo.IsDuplicateKeyError(err) {
		return application.ErrApplied
	}
	return err
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Proposal, error) {
	var d proposalDoc
	if err := r.proposals.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		return domain.Proposal{}, mapped(err)
	}
	return fromDoc(d)
}
func (r *Repository) FindByCommand(ctx context.Context, command string) (domain.Proposal, error) {
	var d proposalDoc
	if err := r.proposals.FindOne(ctx, bson.M{"commandId": command}).Decode(&d); err != nil {
		return domain.Proposal{}, mapped(err)
	}
	return fromDoc(d)
}
func (r *Repository) SaveWithAudit(ctx context.Context, p domain.Proposal, expected uint64, a domain.Audit) error {
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, e := r.proposals.ReplaceOne(tx, bson.M{"_id": p.ID(), "version": expected}, toDoc(p))
		if e != nil {
			return e
		}
		if result.MatchedCount == 0 {
			return application.ErrConflict
		}
		_, e = r.audits.InsertOne(tx, toAuditDoc(a))
		return e
	})
	if mongo.IsDuplicateKeyError(err) {
		return application.ErrApplied
	}
	return err
}
func toAuditDoc(a domain.Audit) auditDoc {
	return auditDoc{a.ID, a.ProposalID, a.ActorKey, a.Kind, a.Version, a.At}
}
func toDoc(p domain.Proposal) proposalDoc {
	s := p.State()
	return proposalDoc{s.ID, s.CommandID, s.Fingerprint, s.ProposerKey, s.ApproverKey, s.Capability, s.Environment, s.Market, s.Action, s.Reason, s.Status, s.Version, s.CreatedAt, s.ExpiresAt, s.ApprovedAt, s.AppliedAt}
}
func fromDoc(d proposalDoc) (domain.Proposal, error) {
	return domain.Rehydrate(domain.State{
		ID: d.ID, CommandID: d.CommandID, Fingerprint: d.Fingerprint,
		ProposerKey: d.ProposerKey, ApproverKey: d.ApproverKey,
		Capability: d.Capability, Environment: d.Environment, Market: d.Market,
		Action: d.Action, Reason: d.Reason, Status: d.Status, Version: d.Version,
		CreatedAt: d.CreatedAt, ExpiresAt: d.ExpiresAt,
		ApprovedAt: d.ApprovedAt, AppliedAt: d.AppliedAt,
	})
}
func mapped(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrNotFound
	}
	return err
}
