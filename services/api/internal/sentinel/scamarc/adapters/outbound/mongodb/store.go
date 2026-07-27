// Package mongodb persists scam-arc signals and per-room ladder states.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/domain"
)

type Store struct {
	database *mongo.Database
}

func NewStore(database *mongo.Database) *Store {
	return &Store{database: database}
}

func (store *Store) signals() *mongo.Collection {
	return store.database.Collection("scamarc_signals")
}

func (store *Store) states() *mongo.Collection {
	return store.database.Collection("scamarc_states")
}

type signalDocument struct {
	ID         string    `bson:"_id"`
	RoomID     string    `bson:"roomId"`
	ActorID    string    `bson:"actorId"`
	Kind       string    `bson:"kind"`
	ObservedAt time.Time `bson:"observedAt"`
}

type stateDocument struct {
	RoomID    string    `bson:"_id"`
	Score     float64   `bson:"score"`
	Ladder    string    `bson:"ladder"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.signals().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "roomId", Value: 1}, {Key: "observedAt", Value: 1}},
		Options: options.Index().SetName("scamarc_room_time"),
	})
	return err
}

func (store *Store) Append(ctx context.Context, signal domain.Signal) error {
	_, err := store.signals().InsertOne(ctx, signalDocument{
		ID: signal.ID, RoomID: signal.RoomID, ActorID: signal.ActorID,
		Kind: string(signal.Kind), ObservedAt: signal.ObservedAt.UTC(),
	})
	return err
}

func (store *Store) KindsForRoom(ctx context.Context, roomID string) ([]domain.SignalKind, error) {
	cursor, err := store.signals().Find(ctx, bson.M{"roomId": roomID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kinds []domain.SignalKind
	for cursor.Next(ctx) {
		var document signalDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		kinds = append(kinds, domain.SignalKind(document.Kind))
	}
	return kinds, cursor.Err()
}

func (store *Store) SaveState(ctx context.Context, state domain.RoomState) error {
	_, err := store.states().ReplaceOne(ctx,
		bson.M{"_id": state.RoomID},
		stateDocument{RoomID: state.RoomID, Score: state.Score, Ladder: string(state.Ladder), UpdatedAt: state.UpdatedAt.UTC()},
		options.Replace().SetUpsert(true))
	return err
}

func (store *Store) StateForRoom(ctx context.Context, roomID string) (domain.RoomState, error) {
	var document stateDocument
	if err := store.states().FindOne(ctx, bson.M{"_id": roomID}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.RoomState{RoomID: roomID, Ladder: domain.LadderNone}, nil
		}
		return domain.RoomState{}, err
	}
	return domain.RoomState{
		RoomID: document.RoomID, Score: document.Score,
		Ladder: domain.LadderState(document.Ladder), UpdatedAt: document.UpdatedAt,
	}, nil
}
