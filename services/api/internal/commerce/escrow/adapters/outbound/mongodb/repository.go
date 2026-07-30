package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	ledgerprivacy "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/adapters/outbound/privacy"
	ledgerdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

type Repository struct {
	database        *mongo.Database
	c               *mongo.Collection
	settlementKeyer *ledgerprivacy.Keyer
}

func New(db *mongo.Database) *Repository {
	return &Repository{database: db, c: db.Collection("commerce_escrows")}
}
func NewWithSettlementSecret(db *mongo.Database, secret []byte) (*Repository, error) {
	keyer, err := ledgerprivacy.New(secret)
	if err != nil {
		return nil, err
	}
	return &Repository{database: db, c: db.Collection("commerce_escrows"), settlementKeyer: &keyer}, nil
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "appliedids", Value: 1}}, Options: options.Index().SetUnique(true).SetName("escrow_command_unique")}, {Keys: bson.D{{Key: "fundingref", Value: 1}}, Options: options.Index().SetUnique(true).SetName("escrow_funding_unique")}, {Keys: bson.D{{Key: "ownerkey", Value: 1}, {Key: "events.0.at", Value: -1}}, Options: options.Index().SetName("escrow_owner")}, {Keys: bson.D{{Key: "engagementid", Value: 1}}, Options: options.Index().SetUnique(true).SetName("escrow_engagement_unique")}})
	return e
}
func (r *Repository) FindByCommand(ctx context.Context, command string) (domain.Escrow, error) {
	var state domain.State
	err := r.c.FindOne(ctx, bson.M{"appliedids": command}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Escrow{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Escrow{}, err
	}
	return domain.Rehydrate(state)
}
func (r *Repository) ListForOwner(ctx context.Context, ownerKey string) ([]domain.Escrow, error) {
	cursor, err := r.c.Find(ctx, bson.M{"ownerkey": ownerKey}, options.Find().SetSort(bson.D{{Key: "events.0.at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var states []domain.State
	if err := cursor.All(ctx, &states); err != nil {
		return nil, err
	}
	out := make([]domain.Escrow, 0, len(states))
	for _, state := range states {
		escrow, hydrateErr := domain.Rehydrate(state)
		if hydrateErr != nil {
			return nil, hydrateErr
		}
		out = append(out, escrow)
	}
	return out, nil
}
func (r *Repository) Create(ctx context.Context, x domain.Escrow) error {
	s := x.State()
	_, e := r.c.InsertOne(ctx, s)
	return r.dupe(ctx, s.AppliedIDs[0], e)
}
func (r *Repository) CreateAudited(ctx context.Context, escrow domain.Escrow, actorID string) error {
	state := escrow.State()
	return apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		if _, err := r.c.InsertOne(tx, state); err != nil {
			return r.dupe(tx, state.AppliedIDs[0], err)
		}
		_, err := r.database.Collection("admin_access").InsertOne(tx, bson.M{
			"actorId": actorID, "action": "admin.escrow.funding",
			"target": state.ID, "at": state.Events[0].At,
		})
		return err
	})
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Escrow, error) {
	var s domain.State
	e := r.c.FindOne(ctx, bson.M{"id": id}).Decode(&s)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return domain.Escrow{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Escrow{}, e
	}
	return domain.Rehydrate(s)
}
func (r *Repository) Save(ctx context.Context, x domain.Escrow, expected uint64, command string) error {
	s := x.State()
	result, e := r.c.ReplaceOne(ctx, bson.M{"id": s.ID, "revision": expected}, s)
	if e != nil {
		return r.dupe(ctx, command, e)
	}
	if result.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"appliedids": command}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) SaveAudited(ctx context.Context, escrow domain.Escrow, expected uint64, command, actorID, action string) error {
	state := escrow.State()
	return apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, err := r.c.ReplaceOne(tx, bson.M{"id": state.ID, "revision": expected}, state)
		if err != nil {
			return r.dupe(tx, command, err)
		}
		if result.MatchedCount == 0 {
			if r.c.FindOne(tx, bson.M{"appliedids": command}).Err() == nil {
				return application.ErrApplied
			}
			return application.ErrConflict
		}
		_, err = r.database.Collection("admin_access").InsertOne(tx, bson.M{
			"actorId": actorID, "action": action, "target": state.ID,
			"commandId": command, "at": state.Events[len(state.Events)-1].At,
		})
		return err
	})
}
func (r *Repository) SettleAudited(ctx context.Context, escrow domain.Escrow, expected uint64, command, actorID string, statement domain.Statement) error {
	if r.settlementKeyer == nil {
		return application.ErrUnavailable
	}
	liability, err := r.settlementKeyer.Key("ledger-account:liability", "escrow:"+statement.EscrowID)
	if err != nil {
		return err
	}
	payable, err := r.settlementKeyer.Key("ledger-account:liability", "matchmaker-payable:"+escrow.State().EngagementID)
	if err != nil {
		return err
	}
	revenue, err := r.settlementKeyer.Key("ledger-account:revenue", "agyina-platform-fee")
	if err != nil {
		return err
	}
	reference, err := r.settlementKeyer.Key("ledger-reference:sale_settlement", statement.Ref)
	if err != nil {
		return err
	}
	lines := []ledgerdomain.Line{
		{AccountKey: liability, Class: ledgerdomain.ClassLiability, Side: ledgerdomain.SideDebit, Minor: int64(statement.GrossPesewas)},
		{AccountKey: payable, Class: ledgerdomain.ClassLiability, Side: ledgerdomain.SideCredit, Minor: int64(statement.NetPesewas)},
	}
	if statement.FeePesewas > 0 {
		lines = append(lines, ledgerdomain.Line{AccountKey: revenue, Class: ledgerdomain.ClassRevenue, Side: ledgerdomain.SideCredit, Minor: int64(statement.FeePesewas)})
	}
	posting, err := ledgerdomain.NewPosting(statement.Ref, command, reference, ledgerdomain.PurposeSaleSettlement, ledgerdomain.CurrencyGHS, lines, statement.SettledAt)
	if err != nil {
		return application.ErrInvalid
	}
	state := escrow.State()
	return apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, replaceErr := r.c.ReplaceOne(tx, bson.M{"id": state.ID, "revision": expected}, state)
		if replaceErr != nil {
			return r.dupe(tx, command, replaceErr)
		}
		if result.MatchedCount == 0 {
			if r.c.FindOne(tx, bson.M{"appliedids": command}).Err() == nil {
				return application.ErrApplied
			}
			return application.ErrConflict
		}
		_, insertErr := r.database.Collection("commerce_ledger_postings").InsertOne(tx, bson.M{
			"_id": posting.ID(), "commandId": posting.CommandID(), "fingerprint": posting.Fingerprint(),
			"referenceKey": posting.ReferenceKey(), "purpose": posting.Purpose(), "currency": posting.Currency(),
			"lines": posting.Lines(), "postedAt": posting.PostedAt(),
		})
		if insertErr != nil {
			if mongo.IsDuplicateKeyError(insertErr) {
				return application.ErrApplied
			}
			return insertErr
		}
		_, insertErr = r.database.Collection("admin_access").InsertOne(tx, bson.M{
			"actorId": actorID, "action": "admin.escrow.settlement", "target": state.ID,
			"commandId": command, "at": statement.SettledAt,
		})
		return insertErr
	})
}
func (r *Repository) dupe(ctx context.Context, id string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"appliedids": id}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
