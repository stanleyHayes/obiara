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

type accountDocument struct {
	ID        string    `bson:"_id"`
	Phone     string    `bson:"phone"`
	Status    string    `bson:"status"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (repository *AccountRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetName("accounts_phone_unique").SetUnique(true),
	})
	return err
}

func (repository *AccountRepository) Create(ctx context.Context, account domain.Account) error {
	_, err := repository.collection().InsertOne(ctx, accountDocument{
		ID:        account.ID(),
		Phone:     account.Phone(),
		Status:    string(account.Status()),
		CreatedAt: account.CreatedAt(),
	})
	if apimongo.IsDuplicateKey(err) {
		return application.ErrAccountExists
	}
	return err
}

func (repository *AccountRepository) FindByPhone(ctx context.Context, phone string) (domain.Account, error) {
	var document accountDocument
	if err := repository.collection().FindOne(ctx, bson.M{"phone": phone}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Account{}, application.ErrAccountNotFound
		}
		return domain.Account{}, err
	}
	return domain.ReconstituteAccount(document.ID, document.Phone, domain.AccountStatus(document.Status), document.CreatedAt), nil
}
