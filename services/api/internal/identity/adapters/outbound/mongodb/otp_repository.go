package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// OtpChallengeRepository persists OTP challenges. Plaintext codes never
// reach this adapter; documents hold hashes only.
type OtpChallengeRepository struct {
	database *mongo.Database
}

func NewOtpChallengeRepository(database *mongo.Database) *OtpChallengeRepository {
	return &OtpChallengeRepository{database: database}
}

func (repository *OtpChallengeRepository) collection() *mongo.Collection {
	return repository.database.Collection("otp_challenges")
}

type otpDocument struct {
	ID         string     `bson:"_id"`
	Channel    string     `bson:"channel"`
	Contact    string     `bson:"contact"`
	CodeHash   string     `bson:"codeHash"`
	ExpiresAt  time.Time  `bson:"expiresAt"`
	Attempts   int        `bson:"attempts"`
	SentCount  int        `bson:"sentCount"`
	CreatedAt  time.Time  `bson:"createdAt"`
	ConsumedAt *time.Time `bson:"consumedAt,omitempty"`
}

func (repository *OtpChallengeRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "channel", Value: 1}, {Key: "contact", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("otp_contact_latest"),
		},
		{
			// Challenges are ephemeral verification artifacts; expire the
			// documents one hour after creation (legal-retention-safe: they
			// carry no content beyond a hash and counters).
			Keys:    bson.D{{Key: "createdAt", Value: 1}},
			Options: options.Index().SetName("otp_ttl").SetExpireAfterSeconds(3600),
		},
	})
	return err
}

func (repository *OtpChallengeRepository) Create(ctx context.Context, challenge domain.OtpChallenge) error {
	_, err := repository.collection().InsertOne(ctx, toOtpDocument(challenge))
	return err
}

func (repository *OtpChallengeRepository) LatestByContact(ctx context.Context, contact domain.Contact) (domain.OtpChallenge, error) {
	var document otpDocument
	err := repository.collection().FindOne(ctx,
		bson.M{"channel": string(contact.Channel()), "contact": contact.Value()},
		options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.OtpChallenge{}, application.ErrChallengeNotFound
		}
		return domain.OtpChallenge{}, err
	}
	return toOtpDomain(document), nil
}

func (repository *OtpChallengeRepository) Update(ctx context.Context, challenge domain.OtpChallenge) error {
	document := toOtpDocument(challenge)
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{"_id": document.ID},
		bson.M{"$set": bson.M{"attempts": document.Attempts, "consumedAt": document.ConsumedAt}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrChallengeNotFound
	}
	return nil
}

func toOtpDocument(challenge domain.OtpChallenge) otpDocument {
	return otpDocument{
		ID:         challenge.ID(),
		Channel:    string(challenge.Contact().Channel()),
		Contact:    challenge.Contact().Value(),
		CodeHash:   challenge.CodeHash(),
		ExpiresAt:  challenge.ExpiresAt(),
		Attempts:   challenge.Attempts(),
		SentCount:  challenge.SentCount(),
		CreatedAt:  challenge.CreatedAt(),
		ConsumedAt: challenge.ConsumedAt(),
	}
}

func toOtpDomain(document otpDocument) domain.OtpChallenge {
	return domain.ReconstituteChallenge(
		document.ID,
		domain.ReconstituteContact(domain.Channel(document.Channel), document.Contact),
		document.CodeHash,
		document.ExpiresAt,
		document.Attempts,
		document.SentCount,
		document.CreatedAt,
		document.ConsumedAt,
	)
}
