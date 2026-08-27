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

// AccountRepository persists contact-bound accounts with a unique index on
// the channel-and-value pair (FR-102: exactly one active account per
// verified identity).
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
	ID             string     `bson:"_id"`
	Channel        string     `bson:"channel"`
	Contact        string     `bson:"contact"`
	Status         string     `bson:"status"`
	Tier           int        `bson:"tier"`
	Version        int64      `bson:"version"`
	SuspendedUntil *time.Time `bson:"suspendedUntil,omitempty"`
	CreatedAt      time.Time  `bson:"createdAt"`
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
	// The old index keyed accounts on "phone" alone. It has to go rather
	// than sit alongside the new one: a unique index treats a missing field
	// as null, so the first two email accounts — neither carrying a phone —
	// would collide with each other on it. Dropping a index that is not
	// there is not an error worth failing a boot over.
	if err := repository.collection().Indexes().DropOne(ctx, "accounts_phone_unique"); err != nil &&
		!apimongo.IsIndexNotFound(err) {
		return err
	}
	if _, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "channel", Value: 1}, {Key: "contact", Value: 1}},
		Options: options.Index().SetName("accounts_contact_unique").SetUnique(true),
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

func (repository *AccountRepository) FindByContact(ctx context.Context, contact domain.Contact) (domain.Account, error) {
	return repository.findOne(ctx, bson.M{"channel": string(contact.Channel()), "contact": contact.Value()})
}

func (repository *AccountRepository) FindByID(ctx context.Context, id string) (domain.Account, error) {
	return repository.findOne(ctx, bson.M{"_id": id})
}

// List returns the newest accounts first. The repository returns domain
// accounts; privacy-safe projection remains the inbound adapter's duty.
func (repository *AccountRepository) List(ctx context.Context, limit int) ([]domain.Account, error) {
	if limit < 1 || limit > 200 {
		limit = 100
	}
	cursor, err := repository.collection().Find(
		ctx,
		bson.M{},
		options.Find().
			SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: 1}}).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	accounts := make([]domain.Account, 0)
	for cursor.Next(ctx) {
		var document accountDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		accounts = append(accounts, domain.ReconstituteAccount(
			document.ID, toContact(document), domain.AccountStatus(document.Status),
			domain.Tier(document.Tier), document.Version, document.SuspendedUntil, document.CreatedAt))
	}
	return accounts, cursor.Err()
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
		toContact(document),
		domain.AccountStatus(document.Status),
		domain.Tier(document.Tier),
		document.Version,
		document.SuspendedUntil,
		document.CreatedAt,
	), nil
}

// ListSuspendedExpired returns suspended accounts whose suspension ended
// before now, oldest first.
func (repository *AccountRepository) ListSuspendedExpired(ctx context.Context, now time.Time, limit int) ([]domain.Account, error) {
	if limit < 1 {
		limit = 100
	}
	cursor, err := repository.collection().Find(ctx,
		bson.M{"status": string(domain.AccountSuspended), "suspendedUntil": bson.M{"$lte": now.UTC()}},
		options.Find().SetSort(bson.D{{Key: "suspendedUntil", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []domain.Account
	for cursor.Next(ctx) {
		var document accountDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		accounts = append(accounts, domain.ReconstituteAccount(
			document.ID, toContact(document), domain.AccountStatus(document.Status),
			domain.Tier(document.Tier), document.Version, document.SuspendedUntil, document.CreatedAt))
	}
	return accounts, cursor.Err()
}

// Update persists status transitions (suspend, block, reactivate) with
// optimistic concurrency.
func (repository *AccountRepository) Update(ctx context.Context, account domain.Account) error {
	document := toAccountDocument(account)
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{"_id": document.ID, "version": document.Version - 1},
		bson.M{"$set": bson.M{"status": document.Status, "suspendedUntil": document.SuspendedUntil, "version": document.Version}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrStaleSession
	}
	return nil
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

// toContact rebuilds the stored identity.
func toContact(document accountDocument) domain.Contact {
	return domain.ReconstituteContact(domain.Channel(document.Channel), document.Contact)
}

func toAccountDocument(account domain.Account) accountDocument {
	return accountDocument{
		ID:             account.ID(),
		Channel:        string(account.Contact().Channel()),
		Contact:        account.Contact().Value(),
		Status:         string(account.Status()),
		Tier:           int(account.Tier()),
		Version:        account.Version(),
		SuspendedUntil: account.SuspendedUntil(),
		CreatedAt:      account.CreatedAt(),
	}
}
