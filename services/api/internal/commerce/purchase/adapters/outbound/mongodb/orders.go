// Package mongodb remembers what each collection was opened to buy.
package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase"
)

type Orders struct{ orders *mongo.Collection }

func NewOrders(database *mongo.Database) *Orders {
	return &Orders{orders: database.Collection("purchase_orders")}
}

func (store *Orders) EnsureIndexes(ctx context.Context) error {
	_, err := store.orders.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "memberId", Value: 1}, {Key: "recordedAt", Value: -1}},
			Options: options.Index().SetName("purchase_order_member"),
		},
	})
	return err
}

type document struct {
	IntentID      string `bson:"_id"`
	SKUKey        string `bson:"skuKey"`
	SKUVersion    uint64 `bson:"skuVersion"`
	MemberID      string `bson:"memberId"`
	AmountPesewas int64  `bson:"amountPesewas"`
	Code          string `bson:"code,omitempty"`
}

// Record writes the order once.
//
// Upsert on the intent id rather than insert, because a retried purchase
// carrying the same idempotency key reaches the same collection, and the
// second attempt describes the same order. Failing it would refuse a retry
// that is doing nothing wrong.
func (store *Orders) Record(ctx context.Context, order purchase.Order) error {
	_, err := store.orders.UpdateOne(ctx,
		bson.M{"_id": order.IntentID},
		bson.M{"$set": document{
			IntentID: order.IntentID, SKUKey: order.SKUKey, SKUVersion: order.SKUVersion,
			MemberID: order.MemberID, AmountPesewas: order.AmountPesewas, Code: order.Code,
		}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (store *Orders) Find(ctx context.Context, intentID string) (purchase.Order, error) {
	var stored document
	if err := store.orders.FindOne(ctx, bson.M{"_id": intentID}).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return purchase.Order{}, purchase.ErrOrderNotFound
		}
		return purchase.Order{}, err
	}
	return purchase.Order{
		IntentID: stored.IntentID, SKUKey: stored.SKUKey, SKUVersion: stored.SKUVersion,
		MemberID: stored.MemberID, AmountPesewas: stored.AmountPesewas, Code: stored.Code,
	}, nil
}
