package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ collection *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db.Collection("courtship_safety")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("safety_command")})
	return err
}

type document struct {
	ID        string              `bson:"_id"`
	Members   []string            `bson:"members"`
	Blocked   bool                `bson:"blocked"`
	BlockedAt time.Time           `bson:"blockedAt,omitempty"`
	Revision  uint64              `bson:"revision"`
	Events    []domain.Event      `bson:"events"`
	Reviews   []domain.CareReview `bson:"careReviews"`
	Commands  []domain.Applied    `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, s domain.Safety) error {
	_, err := r.collection.InsertOne(ctx, toDoc(s))
	return err
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Safety, error) {
	var d document
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		return domain.Safety{}, err
	}
	return domain.Rehydrate(domain.State{ID: d.ID, Members: d.Members, Blocked: d.Blocked, BlockedAt: d.BlockedAt, Revision: d.Revision, Events: d.Events, Reviews: d.Reviews, Commands: d.Commands})
}
func (r *Repository) Save(ctx context.Context, s domain.Safety, expected uint64, commandID string) error {
	events, commands := s.Events(), s.Commands()
	// An aggregate with no events has never been acted on, so there is
	// nothing to append and the caller wanted Create. Returning an error
	// beats indexing past the end of an empty slice, which in an HTTP
	// handler is a panic rather than a rejected request.
	if len(events) == 0 || len(commands) == 0 {
		return domain.ErrInvalid
	}
	update := bson.M{"$set": bson.M{"blocked": s.Blocked(), "blockedAt": s.BlockedAt(), "revision": s.Revision()}, "$push": bson.M{"events": events[len(events)-1], "commands": commands[len(commands)-1]}}
	if events[len(events)-1].Action == domain.ActionReport {
		reviews := s.Reviews()
		update["$push"].(bson.M)["careReviews"] = reviews[len(reviews)-1]
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": s.ID(), "revision": expected}, update)
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
func toDoc(s domain.Safety) document {
	events := s.Events()
	if events == nil {
		events = []domain.Event{}
	}
	reviews := s.Reviews()
	if reviews == nil {
		reviews = []domain.CareReview{}
	}
	commands := s.Commands()
	if commands == nil {
		commands = []domain.Applied{}
	}
	return document{ID: s.ID(), Members: s.Members(), Blocked: s.Blocked(), BlockedAt: s.BlockedAt(), Revision: s.Revision(), Events: events, Reviews: reviews, Commands: commands}
}
