package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/application"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ db *mongo.Database }

func New(db *mongo.Database) *Repository { return &Repository{db} }

type ad struct {
	Command, Actor, Reason, Correlation, From, To string
	At                                            time.Time
}
type doc struct {
	ID                                             string `bson:"_id"`
	Kind, SubjectRef, EvidenceRef, Summary, Status string
	Version                                        uint64
	Created                                        time.Time
	Audit                                          []ad
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.db.Collection("community_audit_cases").Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "status", Value: 1}, {Key: "created", Value: 1}}}, {Keys: bson.D{{Key: "audit.command", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)}})
	return e
}
func (r *Repository) List(ctx context.Context, l int) ([]domain.Case, error) {
	cur, e := r.db.Collection("community_audit_cases").Find(ctx, bson.M{"status": string(domain.StatusQueued)}, options.Find().SetLimit(int64(l)))
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, e
	}
	out := make([]domain.Case, 0, len(ds))
	for _, d := range ds {
		c, x := toDomain(d)
		if x != nil {
			return nil, x
		}
		out = append(out, c)
	}
	return out, nil
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Case, error) {
	var d doc
	if e := r.db.Collection("community_audit_cases").FindOne(ctx, bson.M{"_id": id}).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Case{}, application.ErrNotFound
		}
		return domain.Case{}, e
	}
	return toDomain(d)
}
func (r *Repository) RecordEvidenceAccess(ctx context.Context, id, actor, purpose string, at time.Time) error {
	_, e := r.db.Collection("community_evidence_access").InsertOne(ctx, bson.M{"caseId": id, "actorKey": actor, "purposeKey": purpose, "at": at})
	return e
}
func (r *Repository) Decide(ctx context.Context, c domain.Case, v uint64, cmd string) error {
	a := c.Audit()[len(c.Audit())-1]
	x, e := r.db.Collection("community_audit_cases").UpdateOne(ctx, bson.M{"_id": c.ID(), "version": v, "audit.command": bson.M{"$ne": cmd}}, bson.M{"$set": bson.M{"status": c.Status(), "version": c.Version()}, "$push": bson.M{"audit": ad{a.Command(), a.Actor(), a.Reason(), a.Correlation(), string(a.From()), string(a.To()), a.At()}}})
	if e != nil {
		return e
	}
	if x.MatchedCount == 0 {
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) Seed(ctx context.Context, c domain.Case) error {
	_, e := r.db.Collection("community_audit_cases").InsertOne(ctx, toDoc(c))
	return e
}
func toDoc(c domain.Case) doc {
	aa := make([]ad, 0, len(c.Audit()))
	for _, a := range c.Audit() {
		aa = append(aa, ad{a.Command(), a.Actor(), a.Reason(), a.Correlation(), string(a.From()), string(a.To()), a.At()})
	}
	return doc{c.ID(), string(c.Kind()), c.SubjectRef(), c.EvidenceRef(), c.SummaryCode(), string(c.Status()), c.Version(), c.Created(), aa}
}
func toDomain(d doc) (domain.Case, error) {
	aa := make([]domain.Audit, 0, len(d.Audit))
	for _, a := range d.Audit {
		x, e := domain.NewAudit(a.Command, a.Actor, a.Reason, a.Correlation, domain.Status(a.From), domain.Status(a.To), a.At)
		if e != nil {
			return domain.Case{}, e
		}
		aa = append(aa, x)
	}
	return domain.Rehydrate(d.ID, domain.Kind(d.Kind), d.SubjectRef, d.EvidenceRef, d.Summary, domain.Status(d.Status), d.Version, d.Created, aa)
}
