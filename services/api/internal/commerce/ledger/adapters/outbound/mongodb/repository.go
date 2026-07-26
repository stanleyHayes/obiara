package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("commerce_ledger_postings")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("ledger_command_unique")}, {Keys: bson.D{{Key: "lines.accountKey", Value: 1}, {Key: "currency", Value: 1}, {Key: "postedAt", Value: 1}}, Options: options.Index().SetName("ledger_balance_recompute")}, {Keys: bson.D{{Key: "referenceKey", Value: 1}, {Key: "purpose", Value: 1}}, Options: options.Index().SetName("ledger_reference")}})
	return e
}

type doc struct {
	ID           string          `bson:"_id"`
	CommandID    string          `bson:"commandId"`
	Fingerprint  string          `bson:"fingerprint"`
	ReferenceKey string          `bson:"referenceKey"`
	Purpose      domain.Purpose  `bson:"purpose"`
	Currency     domain.Currency `bson:"currency"`
	Lines        []domain.Line   `bson:"lines"`
	PostedAt     time.Time       `bson:"postedAt"`
}

func (r *Repository) Create(ctx context.Context, p domain.Posting) error {
	_, e := r.c.InsertOne(ctx, toDoc(p))
	if e == nil {
		return nil
	}
	if !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"commandId": p.CommandID()}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Posting, error) {
	var d doc
	if e := r.c.FindOne(ctx, bson.M{"commandId": id}).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Posting{}, application.ErrNotFound
		}
		return domain.Posting{}, e
	}
	return rehydrate(d)
}
func (r *Repository) ListLines(ctx context.Context, account string, currency domain.Currency) ([]domain.BookedLine, error) {
	cur, e := r.c.Find(ctx, bson.M{"currency": currency, "lines.accountKey": account}, options.Find().SetSort(bson.D{{Key: "postedAt", Value: 1}, {Key: "_id", Value: 1}}))
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var docs []doc
	if e = cur.All(ctx, &docs); e != nil {
		return nil, e
	}
	var out []domain.BookedLine
	for _, d := range docs {
		p, x := rehydrate(d)
		if x != nil {
			return nil, x
		}
		for _, line := range p.Lines() {
			if line.AccountKey == account {
				out = append(out, domain.BookedLine{PostingID: p.ID(), Currency: p.Currency(), Line: line})
			}
		}
	}
	return out, nil
}
func rehydrate(d doc) (domain.Posting, error) {
	return domain.Rehydrate(domain.State{ID: d.ID, CommandID: d.CommandID, Fingerprint: d.Fingerprint, ReferenceKey: d.ReferenceKey, Purpose: d.Purpose, Currency: d.Currency, Lines: d.Lines, PostedAt: d.PostedAt})
}
func toDoc(p domain.Posting) doc {
	return doc{ID: p.ID(), CommandID: p.CommandID(), Fingerprint: p.Fingerprint(), ReferenceKey: p.ReferenceKey(), Purpose: p.Purpose(), Currency: p.Currency(), Lines: p.Lines(), PostedAt: p.PostedAt()}
}
