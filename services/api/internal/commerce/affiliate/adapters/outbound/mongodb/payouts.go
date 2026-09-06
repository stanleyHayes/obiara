package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
)

type Payouts struct{ payouts *mongo.Collection }

func NewPayouts(database *mongo.Database) *Payouts {
	return &Payouts{payouts: database.Collection("affiliate_payouts")}
}

func (store *Payouts) EnsureIndexes(ctx context.Context) error {
	_, err := store.payouts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "requestedAt", Value: 1}},
			Options: options.Index().SetName("affiliate_payout_queue"),
		},
		{
			// One open request per affiliate. A second would let the same
			// balance be approved twice before either was spent.
			Keys: bson.D{{Key: "affiliateId", Value: 1}},
			Options: options.Index().SetName("affiliate_payout_open").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"status": string(domain.PayoutRequested)}),
		},
	})
	return err
}

type payoutDocument struct {
	ID               string              `bson:"_id"`
	AffiliateID      string              `bson:"affiliateId"`
	GrossPesewas     int64               `bson:"grossPesewas"`
	WithholdingBasis int64               `bson:"withholdingBasisPoints"`
	WithheldPesewas  int64               `bson:"withheldPesewas"`
	NetPesewas       int64               `bson:"netPesewas"`
	Status           domain.PayoutStatus `bson:"status"`
	ApproverKey      string              `bson:"approverKey,omitempty"`
	Reference        string              `bson:"reference,omitempty"`
	TransferCode     string              `bson:"transferCode,omitempty"`
	RequestedAt      time.Time           `bson:"requestedAt"`
	DecidedAt        *time.Time          `bson:"decidedAt,omitempty"`
}

func (store *Payouts) Create(ctx context.Context, payout domain.Payout) error {
	_, err := store.payouts.InsertOne(ctx, toPayoutDocument(payout))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// Already an open request for this affiliate. A second would let
			// one balance be approved twice.
			return application.ErrConflict
		}
		return application.ErrUnavailable
	}
	return nil
}

func (store *Payouts) Find(ctx context.Context, id string) (domain.Payout, error) {
	var stored payoutDocument
	if err := store.payouts.FindOne(ctx, bson.M{"_id": id}).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Payout{}, application.ErrNotFound
		}
		return domain.Payout{}, application.ErrUnavailable
	}
	return toPayoutDomain(stored)
}

// Save writes a transition, and only over the status it was decided from.
//
// The status guard is what stops two operators approving one payout: both read
// it as requested, and only one write lands.
func (store *Payouts) Save(
	ctx context.Context, payout domain.Payout, expected domain.PayoutStatus,
) error {
	result, err := store.payouts.UpdateOne(ctx,
		bson.M{"_id": payout.ID(), "status": string(expected)},
		bson.M{"$set": bson.M{
			"status":       payout.Status(),
			"approverKey":  payout.ApproverKey(),
			"reference":    payout.Reference(),
			"transferCode": payout.TransferCode(),
			"decidedAt":    payout.DecidedAt(),
		}})
	if err != nil {
		return application.ErrUnavailable
	}
	if result.MatchedCount == 0 {
		return application.ErrConflict
	}
	return nil
}

func (store *Payouts) Pending(ctx context.Context, limit int) ([]domain.Payout, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	cursor, err := store.payouts.Find(ctx,
		bson.M{"status": string(domain.PayoutRequested)},
		options.Find().SetSort(bson.D{{Key: "requestedAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, application.ErrUnavailable
	}
	defer cursor.Close(ctx)

	pending := make([]domain.Payout, 0, limit)
	for cursor.Next(ctx) {
		var stored payoutDocument
		if err := cursor.Decode(&stored); err != nil {
			return nil, application.ErrUnavailable
		}
		payout, err := toPayoutDomain(stored)
		if err != nil {
			continue
		}
		pending = append(pending, payout)
	}
	return pending, cursor.Err()
}

func toPayoutDocument(payout domain.Payout) payoutDocument {
	return payoutDocument{
		ID: payout.ID(), AffiliateID: payout.AffiliateID(),
		GrossPesewas: payout.GrossPesewas(), WithholdingBasis: payout.WithholdingBasis(),
		WithheldPesewas: payout.WithheldPesewas(), NetPesewas: payout.NetPesewas(),
		Status: payout.Status(), ApproverKey: payout.ApproverKey(),
		Reference: payout.Reference(), TransferCode: payout.TransferCode(),
		RequestedAt: payout.RequestedAt(), DecidedAt: payout.DecidedAt(),
	}
}

func toPayoutDomain(stored payoutDocument) (domain.Payout, error) {
	return domain.RehydratePayout(
		stored.ID, stored.AffiliateID, stored.GrossPesewas, stored.WithholdingBasis,
		stored.WithheldPesewas, stored.NetPesewas, stored.Status, stored.ApproverKey,
		stored.Reference, stored.TransferCode, stored.RequestedAt, stored.DecidedAt,
	)
}
