package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// AccountRepository persists phone-bound accounts with a unique active
// phone index (FR-102: exactly one active account per verified identity).
type AccountRepository struct {
	database *mongo.Database
}

func NewAccountRepository(database *mongo.Database) *AccountRepository {
	return &AccountRepository{database: database}
}

func (repository *AccountRepository) collection() *mongo.Collection {
	return repository.database.Collection("accounts")
}

func (repository *AccountRepository) transitions() *mongo.Collection {
	return repository.database.Collection("tier_transitions")
}

type accountDocument struct {
	ID        string    `bson:"_id"`
	Phone     string    `bson:"phone"`
	Status    string    `bson:"status"`
	Tier      int       `bson:"tier"`
	Version   int64     `bson:"version"`
	CreatedAt time.Time `bson:"createdAt"`
}

type transitionDocument struct {
	AccountID  string    `bson:"accountId"`
	From       int       `bson:"from"`
	To         int       `bson:"to"`
	Reason     string    `bson:"reason"`
	ActorID    string    `bson:"actorId"`
	OccurredAt time.Time `bson:"occurredAt"`
}

func (repository *AccountRepository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetName("accounts_phone_unique").SetUnique(true),
	}); err != nil {
		return err
	}
	_, err := repository.transitions().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "accountId", Value: 1}, {Key: "occurredAt", Value: 1}},
		Options: options.Index().SetName("tier_transitions_account"),
	})
	return err
}

func (repository *AccountRepository) Create(ctx context.Context, account domain.Account) error {
	_, err := repository.collection().InsertOne(ctx, toAccountDocument(account))
	if apimongo.IsDuplicateKey(err) {
		return application.ErrAccountExists
	}
	return err
}

func (repository *AccountRepository) FindByPhone(ctx context.Context, phone string) (domain.Account, error) {
	return repository.findOne(ctx, bson.M{"phone": phone})
}

func (repository *AccountRepository) FindByID(ctx context.Context, id string) (domain.Account, error) {
	return repository.findOne(ctx, bson.M{"_id": id})
}

func (repository *AccountRepository) findOne(ctx context.Context, filter bson.M) (domain.Account, error) {
	var document accountDocument
	if err := repository.collection().FindOne(ctx, filter).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Account{}, application.ErrAccountNotFound
		}
		return domain.Account{}, err
	}
	return domain.ReconstituteAccount(
		document.ID,
		document.Phone,
		domain.AccountStatus(document.Status),
		domain.Tier(document.Tier),
		document.Version,
		document.CreatedAt,
	), nil
}

// UpdateWithAudit applies a tier transition atomically: the account update
// (optimistic-concurrency pinned) and the immutable audit record commit in
// one transaction (agent_plan.md §7.4).
func (repository *AccountRepository) UpdateWithAudit(ctx context.Context, account domain.Account, transition domain.TierTransition) error {
	return apimongo.WithTransaction(ctx, repository.database.Client(), func(sc context.Context) error {
		document := toAccountDocument(account)
		result, err := repository.collection().UpdateOne(sc,
			bson.M{"_id": document.ID, "version": document.Version - 1},
			bson.M{"$set": bson.M{"tier": document.Tier, "status": document.Status, "version": document.Version}})
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return application.ErrStaleSession
		}
		_, err = repository.transitions().InsertOne(sc, transitionDocument{
			AccountID:  transition.AccountID,
			From:       int(transition.From),
			To:         int(transition.To),
			Reason:     transition.Reason,
			ActorID:    transition.ActorID,
			OccurredAt: transition.OccurredAt,
		})
		return err
	})
}

func toAccountDocument(account domain.Account) accountDocument {
	return accountDocument{
		ID:        account.ID(),
		Phone:     account.Phone(),
		Status:    string(account.Status()),
		Tier:      int(account.Tier()),
		Version:   account.Version(),
		CreatedAt: account.CreatedAt(),
	}
}
