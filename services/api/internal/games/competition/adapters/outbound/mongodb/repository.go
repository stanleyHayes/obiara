package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("game_competitions")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("competition_command_unique")}, {Keys: bson.D{{Key: "matches.resultKey", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("competition_result_unique")}})
	return e
}

type doc struct {
	ID        string               `bson:"_id"`
	CohortKey string               `bson:"cohortKey"`
	Entrants  []string             `bson:"entrants"`
	Matches   []domain.Match       `bson:"matches"`
	Ladder    []domain.LadderEntry `bson:"ladder"`
	Reviews   []domain.Review      `bson:"reviews"`
	Status    domain.Status        `bson:"status"`
	Revision  uint64               `bson:"revision"`
	Events    []domain.Event       `bson:"events"`
	Commands  []domain.Applied     `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, x domain.Competition) error {
	_, e := r.c.InsertOne(ctx, toDoc(x))
	return r.dupe(ctx, x.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Competition, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Competition, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Competition, error) {
	var d doc
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Competition{}, application.ErrNotFound
		}
		return domain.Competition{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, CohortKey: d.CohortKey, Entrants: d.Entrants, Matches: d.Matches, Ladder: d.Ladder, Reviews: d.Reviews, Status: d.Status, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, x domain.Competition, expected uint64, id string) error {
	es, cs := x.Events(), x.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	res, e := r.c.UpdateOne(ctx, bson.M{"_id": x.ID(), "revision": expected}, bson.M{"$set": bson.M{"matches": x.Matches(), "ladder": x.Ladder(), "reviews": x.Reviews(), "status": x.Status(), "revision": x.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
	if e != nil {
		return r.dupe(ctx, id, e)
	}
	if res.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) dupe(ctx context.Context, id string, e error) error {
	if e == nil || !mongo.IsDuplicateKeyError(e) {
		return e
	}
	if r.c.FindOne(ctx, bson.M{"commands.id": id}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func toDoc(x domain.Competition) doc {
	return doc{ID: x.ID(), CohortKey: x.CohortKey(), Entrants: x.Entrants(), Matches: x.Matches(), Ladder: x.Ladder(), Reviews: x.Reviews(), Status: x.Status(), Revision: x.Revision(), Events: x.Events(), Commands: x.Commands()}
}
