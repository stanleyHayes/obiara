package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/games/conduct/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/conduct/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository {
	return &Repository{d.Collection("game_conduct_signals")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("conduct_command_unique")}, {Keys: bson.D{{Key: "eventKey", Value: 1}}, Options: options.Index().SetUnique(true).SetName("conduct_event_unique")}})
	return e
}

type doc struct {
	ID         string             `bson:"_id"`
	GameKey    string             `bson:"gameKey"`
	SubjectKey string             `bson:"subjectKey"`
	EventKey   string             `bson:"eventKey"`
	GameEvent  domain.GameEvent   `bson:"gameEvent"`
	Kind       domain.Kind        `bson:"kind"`
	Reason     domain.Reason      `bson:"reason"`
	Provenance domain.Provenance  `bson:"provenance"`
	RecordedAt time.Time          `bson:"recordedAt"`
	Appeal     domain.AppealState `bson:"appeal"`
	AppealedAt time.Time          `bson:"appealedAt,omitempty"`
	ResolvedAt time.Time          `bson:"resolvedAt,omitempty"`
	Revision   uint64             `bson:"revision"`
	Events     []domain.Event     `bson:"events"`
	Commands   []domain.Applied   `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, s domain.Signal) error {
	_, e := r.c.InsertOne(ctx, toDoc(s))
	return r.dupe(ctx, s.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Signal, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Signal, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (domain.Signal, error) {
	var d doc
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Signal{}, application.ErrNotFound
		}
		return domain.Signal{}, e
	}
	return domain.Rehydrate(domain.State{ID: d.ID, GameKey: d.GameKey, SubjectKey: d.SubjectKey, EventKey: d.EventKey, GameEvent: d.GameEvent, Kind: d.Kind, Reason: d.Reason, Provenance: d.Provenance, RecordedAt: d.RecordedAt, Appeal: d.Appeal, AppealedAt: d.AppealedAt, ResolvedAt: d.ResolvedAt, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, s domain.Signal, expected uint64, id string) error {
	es, cs := s.Events(), s.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return domain.ErrInvalid
	}
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": s.ID(), "revision": expected}, bson.M{"$set": bson.M{"appeal": s.AppealState(), "appealedAt": s.AppealedAt(), "resolvedAt": s.ResolvedAt(), "revision": s.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
	if e != nil {
		return r.dupe(ctx, id, e)
	}
	if x.MatchedCount == 0 {
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
func toDoc(s domain.Signal) doc {
	return doc{ID: s.ID(), GameKey: s.GameKey(), SubjectKey: s.SubjectKey(), EventKey: s.EventKey(), GameEvent: s.GameEvent(), Kind: s.Kind(), Reason: s.Reason(), Provenance: s.Provenance(), RecordedAt: s.RecordedAt(), Appeal: s.AppealState(), AppealedAt: s.AppealedAt(), ResolvedAt: s.ResolvedAt(), Revision: s.Revision(), Events: s.Events(), Commands: s.Commands()}
}
