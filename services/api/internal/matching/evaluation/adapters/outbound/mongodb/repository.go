package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/application"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("matching_offline_evaluations")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("evaluation_command_unique")},
		{Keys: bson.D{{Key: "candidate", Value: 1}, {Key: "candidateVersion", Value: 1}}, Options: options.Index().SetUnique(true).SetName("candidate_version_unique")},
	})
	return e
}

type doc struct {
	ID               string           `bson:"_id"`
	Candidate        string           `bson:"candidate"`
	CandidateVersion uint64           `bson:"candidateVersion"`
	Snapshot         domain.Snapshot  `bson:"snapshot"`
	Metrics          domain.Metrics   `bson:"metrics"`
	Card             domain.ModelCard `bson:"card"`
	Approval         domain.Approval  `bson:"approval"`
	Revision         uint64           `bson:"revision"`
	Events           []domain.Event   `bson:"events"`
	Commands         []domain.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, e domain.Evaluation) error {
	_, err := r.c.InsertOne(ctx, toDoc(e))
	return r.dupe(ctx, e.Commands()[0].ID, err)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Evaluation, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Evaluation, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Evaluation, error) {
	var d doc
	if err := r.c.FindOne(ctx, f).Decode(&d); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Evaluation{}, application.ErrNotFound
		}
		return domain.Evaluation{}, err
	}
	return domain.Rehydrate(domain.State{ID: d.ID, Candidate: d.Candidate, CandidateVersion: d.CandidateVersion, Snapshot: d.Snapshot, Metrics: d.Metrics, Card: d.Card, Approval: d.Approval, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, e domain.Evaluation, expected uint64, command string) error {
	events, commands := e.Events(), e.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalid
	}
	res, err := r.c.UpdateOne(ctx, bson.M{"_id": e.ID(), "revision": expected}, bson.M{"$set": bson.M{"snapshot": e.Snapshot(), "metrics": e.Metrics(), "card": e.Card(), "approval": e.Approval(), "revision": e.Revision()}, "$push": bson.M{"events": events[len(events)-1], "commands": commands[len(commands)-1]}})
	if err != nil {
		return r.dupe(ctx, command, err)
	}
	if res.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) dupe(ctx context.Context, command string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if r.c.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func toDoc(e domain.Evaluation) doc {
	return doc{ID: e.ID(), Candidate: e.Candidate(), CandidateVersion: e.CandidateVersion(), Snapshot: e.Snapshot(), Metrics: e.Metrics(), Card: e.Card(), Approval: e.Approval(), Revision: e.Revision(), Events: e.Events(), Commands: e.Commands()}
}
