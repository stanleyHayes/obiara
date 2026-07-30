package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/cohort/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/cohort/domain"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{collection: database.Collection("competition_cohorts")}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("competition_cohort_command_unique")},
		{Keys: bson.D{{Key: "memberKeys", Value: 1}, {Key: "status", Value: 1}}, Options: options.Index().SetName("competition_cohort_member_status")},
		{Keys: bson.D{{Key: "competitionId", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("competition_cohort_competition_unique")},
	})
	return err
}

type document struct {
	ID            string           `bson:"_id"`
	Capacity      int              `bson:"capacity"`
	MemberKeys    []string         `bson:"memberKeys"`
	Status        domain.Status    `bson:"status"`
	CompetitionID string           `bson:"competitionId,omitempty"`
	Revision      uint64           `bson:"revision"`
	Commands      []domain.Applied `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, cohort domain.Cohort) error {
	_, err := repository.collection.InsertOne(ctx, toDocument(cohort))
	return repository.duplicate(ctx, cohort.Commands()[0].ID, err)
}

func (repository *Repository) Find(ctx context.Context, id string) (domain.Cohort, error) {
	return repository.find(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindByCommand(ctx context.Context, commandID string) (domain.Cohort, error) {
	return repository.find(ctx, bson.M{"commands.id": commandID})
}

func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Cohort, error) {
	var stored document
	if err := repository.collection.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Cohort{}, application.ErrNotFound
		}
		return domain.Cohort{}, err
	}
	return domain.Rehydrate(domain.State{
		ID: stored.ID, Capacity: stored.Capacity, MemberKeys: stored.MemberKeys,
		Status: stored.Status, CompetitionID: stored.CompetitionID,
		Revision: stored.Revision, Commands: stored.Commands,
	})
}

func (repository *Repository) Append(ctx context.Context, cohort domain.Cohort, expected uint64, commandID string) error {
	commands := cohort.Commands()
	if len(commands) != int(cohort.Revision()) || len(commands) != int(expected+1) {
		return domain.ErrInvalid
	}
	result, err := repository.collection.UpdateOne(
		ctx,
		bson.M{"_id": cohort.ID(), "revision": expected},
		bson.M{
			"$set": bson.M{
				"memberKeys": cohort.MemberKeys(), "status": cohort.Status(),
				"competitionId": cohort.CompetitionID(), "revision": cohort.Revision(),
			},
			"$push": bson.M{"commands": commands[len(commands)-1]},
		},
	)
	if err != nil {
		return repository.duplicate(ctx, commandID, err)
	}
	if result.MatchedCount == 0 {
		if repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}

func (repository *Repository) duplicate(ctx context.Context, commandID string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if repository.collection.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}

func toDocument(cohort domain.Cohort) document {
	return document{
		ID: cohort.ID(), Capacity: cohort.Capacity(), MemberKeys: cohort.MemberKeys(),
		Status: cohort.Status(), CompetitionID: cohort.CompetitionID(),
		Revision: cohort.Revision(), Commands: cohort.Commands(),
	}
}
