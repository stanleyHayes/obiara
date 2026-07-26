package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/application"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

// DoorwayRepository persists doorway questions (one per member).
type DoorwayRepository struct {
	database *mongo.Database
}

func NewDoorwayRepository(database *mongo.Database) *DoorwayRepository {
	return &DoorwayRepository{database: database}
}

func (repository *DoorwayRepository) collection() *mongo.Collection {
	return repository.database.Collection("doorway_questions")
}

type doorwayDocument struct {
	MemberID  string    `bson:"_id"`
	Text      string    `bson:"text"`
	Custom    bool      `bson:"custom"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func (repository *DoorwayRepository) Upsert(ctx context.Context, question domain.DoorwayQuestion) error {
	_, err := repository.collection().ReplaceOne(ctx,
		bson.M{"_id": question.MemberID()},
		doorwayDocument{
			MemberID:  question.MemberID(),
			Text:      question.Text(),
			Custom:    question.Custom(),
			UpdatedAt: question.UpdatedAt(),
		},
		options.Replace().SetUpsert(true))
	return err
}

func (repository *DoorwayRepository) FindByMember(ctx context.Context, memberID string) (domain.DoorwayQuestion, error) {
	var document doorwayDocument
	if err := repository.collection().FindOne(ctx, bson.M{"_id": memberID}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.DoorwayQuestion{}, application.ErrDoorwayQuestionMissing
		}
		return domain.DoorwayQuestion{}, err
	}
	return domain.NewDoorwayQuestion(document.MemberID, document.Text, document.Custom, document.UpdatedAt)
}

// VaultRepository persists vault item metadata.
type VaultRepository struct {
	database *mongo.Database
}

func NewVaultRepository(database *mongo.Database) *VaultRepository {
	return &VaultRepository{database: database}
}

func (repository *VaultRepository) collection() *mongo.Collection {
	return repository.database.Collection("photo_vault")
}

type vaultDocument struct {
	ID        string    `bson:"_id"`
	MemberID  string    `bson:"memberId"`
	AssetID   string    `bson:"assetId"`
	Position  int       `bson:"position"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (repository *VaultRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "memberId", Value: 1}, {Key: "position", Value: 1}},
			Options: options.Index().SetName("vault_member_position_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "assetId", Value: 1}},
			Options: options.Index().SetName("vault_asset"),
		},
	})
	return err
}

func (repository *VaultRepository) Add(ctx context.Context, item domain.VaultItem) error {
	_, err := repository.collection().InsertOne(ctx, vaultDocument{
		ID:        item.ID(),
		MemberID:  item.MemberID(),
		AssetID:   item.AssetID(),
		Position:  item.Position(),
		CreatedAt: item.CreatedAt(),
	})
	if apimongo.IsDuplicateKey(err) {
		return application.ErrVaultItemConflict
	}
	return err
}

func (repository *VaultRepository) Remove(ctx context.Context, itemID string) error {
	result, err := repository.collection().DeleteOne(ctx, bson.M{"_id": itemID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrVaultItemMissing
	}
	return nil
}

func (repository *VaultRepository) ListByMember(ctx context.Context, memberID string) ([]domain.VaultItem, error) {
	cursor, err := repository.collection().Find(ctx,
		bson.M{"memberId": memberID},
		options.Find().SetSort(bson.D{{Key: "position", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []domain.VaultItem
	for cursor.Next(ctx) {
		var document vaultDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		item, err := domain.NewVaultItem(document.ID, document.MemberID, document.AssetID, document.Position, document.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, cursor.Err()
}

func (repository *VaultRepository) CountByMember(ctx context.Context, memberID string) (int, error) {
	count, err := repository.collection().CountDocuments(ctx, bson.M{"memberId": memberID})
	return int(count), err
}
