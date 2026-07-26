package mongodb

import (
	"context"
	"errors"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database: database} }
func (r *Repository) heads() *mongo.Collection {
	return r.database.Collection("seed_allowance_ledgers")
}
func (r *Repository) entries() *mongo.Collection {
	return r.database.Collection("seed_allowance_entries")
}

type headDocument struct {
	ID              string    `bson:"_id"`
	WeeklyAllowance int64     `bson:"weeklyAllowance"`
	Balance         int64     `bson:"balance"`
	WeekStart       time.Time `bson:"weekStart"`
	Version         int64     `bson:"version"`
}
type entryDocument struct {
	LedgerID     string    `bson:"ledgerId"`
	Sequence     int64     `bson:"sequence"`
	Kind         string    `bson:"kind"`
	CommandID    string    `bson:"commandId"`
	Fingerprint  string    `bson:"fingerprint"`
	Units        int64     `bson:"units"`
	BalanceAfter int64     `bson:"balanceAfter"`
	WeekStart    time.Time `bson:"weekStart"`
	OccurredAt   time.Time `bson:"occurredAt"`
	ActorKey     string    `bson:"actorKey"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.entries().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "ledgerId", Value: 1}, {Key: "sequence", Value: 1}},
		Options: options.Index().SetName("seed_allowance_ledger_sequence_unique").SetUnique(true),
	}); err != nil {
		return err
	}
	_, err := r.entries().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "ledgerId", Value: 1}, {Key: "commandId", Value: 1}},
		Options: options.Index().SetName("seed_allowance_command_unique").SetUnique(true),
	})
	return err
}

func (r *Repository) Create(ctx context.Context, ledger domain.Ledger) error {
	entries := ledger.Entries()
	return apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		if _, err := r.heads().InsertOne(tx, toHead(ledger)); err != nil {
			return err
		}
		_, err := r.entries().InsertOne(tx, toEntry(ledger.SubjectKey(), entries[0]))
		return err
	})
}

func (r *Repository) Find(ctx context.Context, subjectKey string) (domain.Ledger, error) {
	var head headDocument
	if err := r.heads().FindOne(ctx, bson.M{"_id": subjectKey}).Decode(&head); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Ledger{}, application.ErrNotFound
		}
		return domain.Ledger{}, err
	}
	cursor, err := r.entries().Find(ctx, bson.M{"ledgerId": subjectKey}, options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}))
	if err != nil {
		return domain.Ledger{}, err
	}
	defer cursor.Close(ctx)
	var docs []entryDocument
	if err = cursor.All(ctx, &docs); err != nil {
		return domain.Ledger{}, err
	}
	entries := make([]domain.Entry, 0, len(docs))
	for _, doc := range docs {
		entries = append(entries, fromEntry(doc))
	}
	return domain.Rehydrate(head.ID, head.WeeklyAllowance, head.Balance, head.WeekStart, head.Version, entries)
}

func (r *Repository) Save(ctx context.Context, ledger domain.Ledger, expectedVersion int64) error {
	newEntries, err := ledger.EntriesAfter(expectedVersion)
	if err != nil {
		return err
	}
	if len(newEntries) == 0 {
		return nil
	}
	return apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, err := r.heads().UpdateOne(tx,
			bson.M{"_id": ledger.SubjectKey(), "version": expectedVersion},
			bson.M{"$set": bson.M{"balance": ledger.Balance(), "weekStart": ledger.WeekStart(), "version": ledger.Version()}})
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return application.ErrConcurrentChange
		}
		documents := make([]any, 0, len(newEntries))
		for _, entry := range newEntries {
			documents = append(documents, toEntry(ledger.SubjectKey(), entry))
		}
		_, err = r.entries().InsertMany(tx, documents)
		return err
	})
}

func toHead(l domain.Ledger) headDocument {
	return headDocument{l.SubjectKey(), l.WeeklyAllowance(), l.Balance(), l.WeekStart(), l.Version()}
}
func toEntry(id string, e domain.Entry) entryDocument {
	return entryDocument{id, e.Sequence, string(e.Kind), e.CommandID, e.Fingerprint, e.Units, e.BalanceAfter, e.WeekStart, e.OccurredAt, e.ActorKey}
}
func fromEntry(e entryDocument) domain.Entry {
	return domain.Entry{Sequence: e.Sequence, Kind: domain.EntryKind(e.Kind), CommandID: e.CommandID, Fingerprint: e.Fingerprint, Units: e.Units, BalanceAfter: e.BalanceAfter, WeekStart: e.WeekStart, OccurredAt: e.OccurredAt, ActorKey: e.ActorKey}
}
