package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/application"
	"github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ records, leases *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("compliance_retention_records"), db.Collection("compliance_retention_leases")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, e := r.records.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("retention_command_unique")}, {Keys: bson.D{{Key: "subjectKey", Value: 1}, {Key: "policy.dataClass", Value: 1}, {Key: "policy.purpose", Value: 1}}, Options: options.Index().SetUnique(true).SetName("retention_scope_unique")}, {Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetName("retention_due_review_no_ttl")}}); e != nil {
		return e
	}
	_, e := r.leases.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0).SetName("retention_coordination_lease_ttl")})
	return e
}

type doc struct {
	ID              string           `bson:"_id"`
	SubjectKey      string           `bson:"subjectKey"`
	Policy          domain.Policy    `bson:"policy"`
	Status          domain.Status    `bson:"status"`
	CreatedAt       time.Time        `bson:"createdAt"`
	ExpiresAt       time.Time        `bson:"expiresAt"`
	ErasedAt        time.Time        `bson:"erasedAt,omitempty"`
	Hold            domain.Hold      `bson:"hold"`
	VerificationKey string           `bson:"verificationKey,omitempty"`
	Revision        uint64           `bson:"revision"`
	Events          []domain.Event   `bson:"events"`
	Commands        []domain.Applied `bson:"commands"`
	Counters        domain.Counters  `bson:"counters"`
}

func (r *Repository) Create(ctx context.Context, x domain.Record) error {
	_, e := r.records.InsertOne(ctx, toDoc(x))
	return r.dupe(ctx, x.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Record, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Record, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Record, error) {
	var d doc
	if e := r.records.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Record{}, application.ErrNotFound
		}
		return domain.Record{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, SubjectKey: d.SubjectKey, Policy: d.Policy, Status: d.Status, CreatedAt: d.CreatedAt, ExpiresAt: d.ExpiresAt, ErasedAt: d.ErasedAt, Hold: d.Hold, VerificationKey: d.VerificationKey, Revision: d.Revision, Events: d.Events, Commands: d.Commands, Counters: d.Counters})
}
func (r *Repository) Append(ctx context.Context, x domain.Record, expected uint64, command string) error {
	es, cs := x.Events(), x.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	res, e := r.records.UpdateOne(ctx, bson.M{"_id": x.ID(), "revision": expected}, bson.M{"$set": bson.M{"status": x.Status(), "hold": x.Hold(), "erasedAt": x.ErasedAt(), "verificationKey": x.VerificationKey(), "revision": x.Revision(), "counters": x.Counters()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
	if e != nil {
		return r.dupe(ctx, command, e)
	}
	if res.MatchedCount == 0 {
		if r.records.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) dupe(ctx context.Context, command string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.records.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func toDoc(x domain.Record) doc {
	return doc{ID: x.ID(), SubjectKey: x.SubjectKey(), Policy: x.Policy(), Status: x.Status(), CreatedAt: x.CreatedAt(), ExpiresAt: x.ExpiresAt(), ErasedAt: x.ErasedAt(), Hold: x.Hold(), VerificationKey: x.VerificationKey(), Revision: x.Revision(), Events: x.Events(), Commands: x.Commands(), Counters: x.Counters()}
}
