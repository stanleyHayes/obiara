package ampe

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/domain"
)

const (
	presenceWindow    = 15 * time.Second
	presenceRetention = 24 * time.Hour
)

type presenceKeyer interface {
	Key(string, string) (string, error)
}

type presenceDoc struct {
	ID        string    `bson:"_id"`
	RoundID   string    `bson:"roundId"`
	PlayerKey string    `bson:"playerKey"`
	FirstSeen time.Time `bson:"firstSeen"`
	LastSeen  time.Time `bson:"lastSeen"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

type presenceRepository struct{ collection *mongo.Collection }
type presenceStore interface {
	touch(context.Context, string, string, time.Time) (presenceDoc, error)
	find(context.Context, string, string) (presenceDoc, error)
}

func newPresenceRepository(database *mongo.Database) *presenceRepository {
	return &presenceRepository{collection: database.Collection("ampe_presence")}
}

func (repository *presenceRepository) ensureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "roundId", Value: 1}, {Key: "playerKey", Value: 1}}, Options: options.Index().SetUnique(true).SetName("ampe_presence_round_player_unique")},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0).SetName("ampe_presence_ttl")},
	})
	return err
}

func (repository *presenceRepository) touch(ctx context.Context, roundID, playerKey string, now time.Time) (presenceDoc, error) {
	id := strings.TrimSpace(roundID) + ":" + playerKey
	after := options.After
	var result presenceDoc
	err := repository.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"lastSeen": now.UTC(), "expiresAt": now.UTC().Add(presenceRetention)},
			"$setOnInsert": bson.M{
				"roundId": roundID, "playerKey": playerKey, "firstSeen": now.UTC(),
			},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(after),
	).Decode(&result)
	return result, err
}

func (repository *presenceRepository) find(ctx context.Context, roundID, playerKey string) (presenceDoc, error) {
	var result presenceDoc
	err := repository.collection.FindOne(ctx, bson.M{
		"roundId": strings.TrimSpace(roundID), "playerKey": playerKey,
	}).Decode(&result)
	return result, err
}

type Presence struct {
	rounds     application.Service
	repository presenceStore
	keyer      presenceKeyer
	now        func() time.Time
}

func (presence Presence) Observe(ctx context.Context, command application.Command) (application.Projection, error) {
	if presence.repository == nil || presence.keyer == nil || presence.now == nil {
		return application.Projection{}, application.ErrNotAvailable
	}
	now := presence.now().UTC()
	actorKey, err := presence.keyer.Key("ampe:player", strings.TrimSpace(command.ActorID))
	if err != nil {
		return application.Projection{}, application.ErrNotAvailable
	}
	self, err := presence.repository.touch(ctx, command.RoundID, actorKey, now)
	if err != nil {
		return application.Projection{}, application.ErrNotAvailable
	}
	projection, err := presence.rounds.View(ctx, command)
	if err != nil {
		return application.Projection{}, err
	}
	if !projection.You.Connected {
		projection, err = presence.apply(ctx, command, command.ActorID, domain.ActionReconnect, projection.Sequence)
		if err != nil {
			return application.Projection{}, err
		}
	}
	otherKey, err := presence.keyer.Key("ampe:player", strings.TrimSpace(command.SecondPlayerID))
	if err != nil {
		return application.Projection{}, application.ErrNotAvailable
	}
	other, findErr := presence.repository.find(ctx, command.RoundID, otherKey)
	otherRecent := findErr == nil && now.Sub(other.LastSeen) <= presenceWindow
	graceElapsed := now.Sub(self.FirstSeen) > presenceWindow
	if ((errors.Is(findErr, mongo.ErrNoDocuments) && graceElapsed) || (findErr == nil && !otherRecent)) && projection.Other.Connected {
		projection, err = presence.apply(ctx, command, command.SecondPlayerID, domain.ActionDisconnect, projection.Sequence)
	} else if otherRecent && !projection.Other.Connected {
		projection, err = presence.apply(ctx, command, command.SecondPlayerID, domain.ActionReconnect, projection.Sequence)
	}
	if err != nil {
		if errors.Is(err, application.ErrConflict) {
			return presence.rounds.View(ctx, command)
		}
		return application.Projection{}, err
	}
	return presence.rounds.View(ctx, command)
}

func (presence Presence) apply(ctx context.Context, command application.Command, actorID string, action domain.Action, sequence uint64) (application.Projection, error) {
	actorKey, err := presence.keyer.Key("ampe:player", strings.TrimSpace(actorID))
	if err != nil {
		return application.Projection{}, application.ErrNotAvailable
	}
	command.ID = "presence:" + string(action) + ":" + command.RoundID + ":" + actorKey[:12] + ":" + fmt.Sprintf("%d", sequence)
	command.ActorID = actorID
	command.ExpectedSequence = sequence
	return presence.rounds.Apply(ctx, command, action, "")
}
