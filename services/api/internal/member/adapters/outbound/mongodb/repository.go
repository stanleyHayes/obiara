package mongodb

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/member/domain"
)

var ErrMemberNotFound = errors.New("member not found")

type Repository struct {
	collection *mongo.Collection
}

type memberDocument struct {
	ID        string    `bson:"_id"`
	Email     string    `bson:"email"`
	CreatedAt time.Time `bson:"createdAt"`
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("members")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetName("members_email_unique").SetUnique(true),
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, member domain.Member) error {
	_, err := repository.collection.InsertOne(ctx, memberDocument{
		ID:        member.ID(),
		Email:     member.Email(),
		CreatedAt: member.CreatedAt(),
	})
	if apimongo.IsDuplicateKey(err) {
		return domain.ErrDuplicateMember
	}
	return err
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Member, error) {
	var document memberDocument
	if err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Member{}, ErrMemberNotFound
		}
		return domain.Member{}, err
	}
	return domain.NewMember(document.ID, document.Email, document.CreatedAt)
}

// FindByEmail reads a member by the address they registered with.
//
// Added for one question the affiliate scheme has to be able to answer:
// whether the party being admitted is a member. Members are never affiliates
// (agent_plan.md §41), and without a way to check, the rule is a comment.
func (repository *Repository) FindByEmail(
	ctx context.Context, email string,
) (domain.Member, error) {
	var document memberDocument
	if err := repository.collection.FindOne(ctx, bson.M{
		"email": strings.ToLower(strings.TrimSpace(email)),
	}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Member{}, ErrMemberNotFound
		}
		return domain.Member{}, err
	}
	return domain.NewMember(document.ID, document.Email, document.CreatedAt)
}
