package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct {
	facts, audits, checkpoints *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("commerce_reconciliation_facts"), db.Collection("commerce_reconciliation_audits"), db.Collection("commerce_reconciliation_checkpoints")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.facts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "eventKey", Value: 1}}, Options: options.Index().SetUnique(true).SetName("reconciliation_event_unique")},
		{Keys: bson.D{{Key: "occurredDay", Value: 1}, {Key: "occurredAt", Value: 1}}, Options: options.Index().SetName("reconciliation_daily_facts")},
	}); err != nil {
		return err
	}
	if _, err := r.audits.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "fingerprint", Value: 1}}, Options: options.Index().SetUnique(true).SetName("reconciliation_audit_unique")}); err != nil {
		return err
	}
	_, err := r.checkpoints.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "day", Value: 1}}, Options: options.Index().SetUnique(true).SetName("reconciliation_day_unique")})
	return err
}

type factDoc struct {
	ID            string                `bson:"_id"`
	ProviderKey   string                `bson:"providerKey"`
	EventKey      string                `bson:"eventKey"`
	ReferenceKey  string                `bson:"referenceKey"`
	LedgerCommand string                `bson:"ledgerCommand"`
	Fingerprint   string                `bson:"fingerprint"`
	Currency      domain.Currency       `bson:"currency"`
	Status        domain.ProviderStatus `bson:"status"`
	Minor         int64                 `bson:"minor"`
	OccurredAt    time.Time             `bson:"occurredAt"`
	ReceivedAt    time.Time             `bson:"receivedAt"`
	OccurredDay   string                `bson:"occurredDay"`
}
type auditDoc struct {
	ID          string               `bson:"_id"`
	FactID      string               `bson:"factId"`
	Fingerprint string               `bson:"fingerprint"`
	Outcome     domain.Outcome       `bson:"outcome"`
	Exception   domain.ExceptionCode `bson:"exception,omitempty"`
	RecordedAt  time.Time            `bson:"recordedAt"`
}
type checkpointDoc struct {
	ID          string    `bson:"_id"`
	Day         string    `bson:"day"`
	Fingerprint string    `bson:"fingerprint"`
	Total       int       `bson:"total"`
	Reconciled  int       `bson:"reconciled"`
	Excepted    int       `bson:"excepted"`
	CompletedAt time.Time `bson:"completedAt"`
}

func (r *Repository) AppendFact(ctx context.Context, f domain.StatementFact) error {
	d := factDoc{f.ID(), f.ProviderKey(), f.EventKey(), f.ReferenceKey(), f.LedgerCommand(), f.Fingerprint(), f.Currency(), f.Status(), f.Minor(), f.OccurredAt(), f.ReceivedAt(), f.OccurredAt().Format("2006-01-02")}
	_, err := r.facts.InsertOne(ctx, d)
	return duplicate(err)
}
func (r *Repository) FindFactByEvent(ctx context.Context, event string) (domain.StatementFact, error) {
	var d factDoc
	if err := r.facts.FindOne(ctx, bson.M{"eventKey": event}).Decode(&d); err != nil {
		return domain.StatementFact{}, mapped(err)
	}
	return factFromDoc(d)
}
func (r *Repository) FindFactByID(ctx context.Context, id string) (domain.StatementFact, error) {
	var d factDoc
	if err := r.facts.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		return domain.StatementFact{}, mapped(err)
	}
	return factFromDoc(d)
}
func (r *Repository) AppendAudit(ctx context.Context, a domain.Audit) error {
	_, err := r.audits.InsertOne(ctx, auditDoc{a.ID(), a.FactID(), a.Fingerprint(), a.Outcome(), a.Exception(), a.RecordedAt()})
	return duplicate(err)
}
func (r *Repository) ListRecentAudits(ctx context.Context, limit int) ([]domain.Audit, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	cur, err := r.audits.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "recordedAt", Value: -1}, {Key: "_id", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var docs []auditDoc
	if err = cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]domain.Audit, 0, len(docs))
	for _, d := range docs {
		audit, hydrateErr := domain.RehydrateAudit(domain.AuditState{
			ID: d.ID, FactID: d.FactID, Fingerprint: d.Fingerprint,
			Outcome: d.Outcome, Exception: d.Exception, RecordedAt: d.RecordedAt,
		})
		if hydrateErr != nil {
			return nil, hydrateErr
		}
		out = append(out, audit)
	}
	return out, nil
}
func (r *Repository) ListFactsForDay(ctx context.Context, day string) ([]domain.StatementFact, error) {
	cur, err := r.facts.Find(ctx, bson.M{"occurredDay": day}, options.Find().SetSort(bson.D{{Key: "occurredAt", Value: 1}, {Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var docs []factDoc
	if err = cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]domain.StatementFact, 0, len(docs))
	for _, d := range docs {
		f, e := factFromDoc(d)
		if e != nil {
			return nil, e
		}
		out = append(out, f)
	}
	return out, nil
}
func (r *Repository) AppendCheckpoint(ctx context.Context, c domain.Checkpoint) error {
	_, err := r.checkpoints.InsertOne(ctx, checkpointDoc{c.ID(), c.Day(), c.Fingerprint(), c.Total(), c.Reconciled(), c.Excepted(), c.CompletedAt()})
	return duplicate(err)
}
func (r *Repository) FindCheckpoint(ctx context.Context, day string) (domain.Checkpoint, error) {
	var d checkpointDoc
	if err := r.checkpoints.FindOne(ctx, bson.M{"day": day}).Decode(&d); err != nil {
		return domain.Checkpoint{}, mapped(err)
	}
	return domain.RehydrateCheckpoint(domain.CheckpointState{ID: d.ID, Day: d.Day, Fingerprint: d.Fingerprint, Total: d.Total, Reconciled: d.Reconciled, Excepted: d.Excepted, CompletedAt: d.CompletedAt})
}
func (r *Repository) ListRecentCheckpoints(ctx context.Context, limit int) ([]domain.Checkpoint, error) {
	if limit < 1 || limit > 31 {
		limit = 14
	}
	cur, err := r.checkpoints.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "day", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var docs []checkpointDoc
	if err = cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]domain.Checkpoint, 0, len(docs))
	for _, d := range docs {
		checkpoint, hydrateErr := domain.RehydrateCheckpoint(domain.CheckpointState{
			ID: d.ID, Day: d.Day, Fingerprint: d.Fingerprint, Total: d.Total,
			Reconciled: d.Reconciled, Excepted: d.Excepted, CompletedAt: d.CompletedAt,
		})
		if hydrateErr != nil {
			return nil, hydrateErr
		}
		out = append(out, checkpoint)
	}
	return out, nil
}
func factFromDoc(d factDoc) (domain.StatementFact, error) {
	return domain.RehydrateFact(domain.FactState{ID: d.ID, ProviderKey: d.ProviderKey, EventKey: d.EventKey, ReferenceKey: d.ReferenceKey, LedgerCommand: d.LedgerCommand, Fingerprint: d.Fingerprint, Currency: d.Currency, Status: d.Status, Minor: d.Minor, OccurredAt: d.OccurredAt, ReceivedAt: d.ReceivedAt})
}
func duplicate(err error) error {
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return application.ErrApplied
	}
	return err
}
func mapped(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return application.ErrNotFound
	}
	return err
}
