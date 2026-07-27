// Package mongodb persists call sessions (metadata only; FR-304).
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/calls/application"
	"github.com/stanleyHayes/obiara/services/api/internal/calls/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) collection() *mongo.Collection {
	return repository.database.Collection("call_sessions")
}

type callDocument struct {
	ID           string     `bson:"_id"`
	RoomID       string     `bson:"roomId"`
	Participants []string   `bson:"participants"`
	Status       string     `bson:"status"`
	Version      int64      `bson:"version"`
	CreatedAt    time.Time  `bson:"createdAt"`
	EndedAt      *time.Time `bson:"endedAt,omitempty"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "roomId", Value: 1}, {Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("calls_room_recent"),
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, call domain.Call) error {
	_, err := repository.collection().InsertOne(ctx, callDocument{
		ID:           call.ID(),
		RoomID:       call.RoomID(),
		Participants: []string{call.Participants()[0], call.Participants()[1]},
		Status:       string(call.Status()),
		Version:      call.Version(),
		CreatedAt:    call.CreatedAt(),
	})
	return err
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Call, error) {
	var document callDocument
	if err := repository.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Call{}, application.ErrCallNotFound
		}
		return domain.Call{}, err
	}
	return domain.ReconstituteCall(
		document.ID, document.RoomID,
		[2]string{document.Participants[0], document.Participants[1]},
		domain.CallStatus(document.Status), document.Version, document.CreatedAt, document.EndedAt,
	), nil
}

func (repository *Repository) Update(ctx context.Context, call domain.Call) error {
	result, err := repository.collection().UpdateOne(ctx,
		bson.M{"_id": call.ID(), "version": call.Version() - 1},
		bson.M{"$set": bson.M{"status": string(call.Status()), "endedAt": call.EndedAt(), "version": call.Version()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrCallNotFound
	}
	return nil
}

// RoomMembership checks room membership against the courtship rooms
// collection (read model; codex's room schema uses `members`).
type RoomMembership struct {
	database *mongo.Database
}

func NewRoomMembership(database *mongo.Database) *RoomMembership {
	return &RoomMembership{database: database}
}

func (membership *RoomMembership) IsMember(ctx context.Context, roomID, memberID string) (bool, error) {
	count, err := membership.database.Collection("rooms").CountDocuments(ctx,
		bson.M{"_id": roomID, "members": memberID})
	return count > 0, err
}
