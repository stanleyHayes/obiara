package mongodb

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("courtship_paces")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "commands.id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("pace_command"),
	})
	return err
}

type document struct {
	ID            string           `bson:"_id"`
	Members       []string         `bson:"members"`
	RelightGrants []string         `bson:"relightGrants"`
	Status        domain.Status    `bson:"status"`
	ResponseAt    time.Time        `bson:"responseAt,omitempty"`
	ArchiveAt     time.Time        `bson:"archiveAt,omitempty"`
	Revision      uint64           `bson:"revision"`
	Events        []domain.Event   `bson:"events"`
	Commands      []domain.Applied `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, pace domain.Pace) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(pace))
	return err
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Pace, error) {
	var stored document
	if err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&stored); err != nil {
		return domain.Pace{}, err
	}
	return domain.Rehydrate(domain.State{
		ID: stored.ID, Members: stored.Members, RelightGrants: stored.RelightGrants,
		Status: stored.Status, ResponseAt: stored.ResponseAt, ArchiveAt: stored.ArchiveAt,
		Revision: stored.Revision, Events: stored.Events, Commands: stored.Commands,
	})
}

func (repository *Repository) Save(ctx context.Context, pace domain.Pace, expected uint64, _ string) error {
	events, commands := pace.Events(), pace.Commands()
	result, err := repository.collection.UpdateOne(
		ctx,
		bson.M{"_id": pace.ID(), "revision": expected},
		bson.M{
			"$set": bson.M{
				"relightGrants": pace.RelightGrants(), "status": pace.Status(),
				"responseAt": pace.ResponseAt(), "archiveAt": pace.ArchiveAt(), "revision": pace.Revision(),
			},
			"$push": bson.M{"events": events[len(events)-1], "commands": commands[len(commands)-1]},
		},
	)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrStaleRevision
	}
	return nil
}

func toDocument(pace domain.Pace) document {
	events := pace.Events()
	if events == nil {
		events = []domain.Event{}
	}
	commands := pace.Commands()
	if commands == nil {
		commands = []domain.Applied{}
	}
	return document{
		ID: pace.ID(), Members: pace.Members(), RelightGrants: pace.RelightGrants(),
		Status: pace.Status(), ResponseAt: pace.ResponseAt(), ArchiveAt: pace.ArchiveAt(),
		Revision: pace.Revision(), Events: events, Commands: commands,
	}
}
