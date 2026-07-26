package mongodb

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func New(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("membership_passes")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "appliedids", Value: 1}}, Options: options.Index().SetUnique(true).SetName("membership_command_unique")},
		{Keys: bson.D{{Key: "receiptref", Value: 1}}, Options: options.Index().SetUnique(true).SetName("membership_receipt_unique")},
		{Keys: bson.D{{Key: "refundrequestref", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("membership_refund_unique")},
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, pass domain.Pass) error {
	state := pass.State()
	_, err := repository.collection.InsertOne(ctx, state)
	return repository.duplicate(ctx, state.AppliedIDs[0], err)
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Pass, error) {
	var state domain.State
	err := repository.collection.FindOne(ctx, bson.M{"id": id}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Pass{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Pass{}, err
	}
	return domain.Rehydrate(state)
}

func (repository *Repository) Save(ctx context.Context, pass domain.Pass, expected uint64, commandID string) error {
	state := pass.State()
	if len(state.Events) != int(expected+1) || len(state.AppliedIDs) != int(expected+1) {
		return domain.ErrInvalid
	}
	result, err := repository.collection.UpdateOne(
		ctx,
		bson.M{"id": state.ID, "revision": expected},
		bson.M{
			"$set": bson.M{
				"cancelledat":        state.CancelledAt,
				"refundrequestref":   state.RefundRequestRef,
				"refundconfirmedref": state.RefundConfirmedRef,
				"revision":           state.Revision,
			},
			"$push": bson.M{
				"events":     state.Events[len(state.Events)-1],
				"appliedids": commandID,
			},
		},
	)
	if err != nil {
		return repository.duplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		if repository.collection.FindOne(ctx, bson.M{"appliedids": commandID}).Err() == nil {
			return application.ErrApplied
		}
		var current domain.State
		if repository.collection.FindOne(ctx, bson.M{"id": state.ID}).Decode(&current) == nil {
			for _, applied := range current.AppliedIDs {
				if applied == commandID {
					return application.ErrApplied
				}
			}
		}
		return application.ErrConflict
	}
	return nil
}

func (repository *Repository) duplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if repository.collection.FindOne(ctx, bson.M{"appliedids": commandID}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
