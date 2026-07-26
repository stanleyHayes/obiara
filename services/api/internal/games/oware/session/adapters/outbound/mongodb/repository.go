package mongodb

import (
	"context"
	"errors"
	oware "github.com/stanleyHayes/obiara/services/api/internal/games/oware/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/application"
	session "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(d *mongo.Database) *Repository { return &Repository{d.Collection("oware_sessions")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("oware_session_command_unique")}, {Keys: bson.D{{Key: "roomRef", Value: 1}, {Key: "status", Value: 1}}, Options: options.Index().SetName("oware_session_room_status")}})
	return e
}

type doc struct {
	ID         string            `bson:"_id"`
	RoomRef    string            `bson:"roomRef"`
	Players    []string          `bson:"players"`
	Houses     [12]int           `bson:"houses"`
	Captured   [2]int            `bson:"captured"`
	GameOver   bool              `bson:"gameOver"`
	Winner     int               `bson:"winner"`
	Turn       int               `bson:"turn"`
	MoveWindow time.Duration     `bson:"moveWindow"`
	Deadline   time.Time         `bson:"deadline"`
	Status     session.Status    `bson:"status"`
	Revision   uint64            `bson:"revision"`
	Events     []session.Event   `bson:"events"`
	Commands   []session.Applied `bson:"commands"`
}

func (r *Repository) Create(ctx context.Context, s session.Session) error {
	_, e := r.c.InsertOne(ctx, toDoc(s))
	return r.dupe(ctx, s.Commands()[0].ID, e)
}
func (r *Repository) Find(ctx context.Context, id string) (session.Session, error) {
	return r.find(ctx, bson.M{"_id": id})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (session.Session, error) {
	return r.find(ctx, bson.M{"commands.id": id})
}
func (r *Repository) find(ctx context.Context, f bson.M) (session.Session, error) {
	var d doc
	if e := r.c.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return session.Session{}, application.ErrNotFound
		}
		return session.Session{}, e
	}
	return session.Rehydrate(session.State{ID: d.ID, RoomRef: d.RoomRef, Players: d.Players, Houses: d.Houses, Captured: d.Captured, GameOver: d.GameOver, Winner: d.Winner, Turn: owarePlayer(d.Turn), MoveWindow: d.MoveWindow, Deadline: d.Deadline, Status: d.Status, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) Append(ctx context.Context, s session.Session, expected uint64, id string) error {
	es, cs := s.Events(), s.Commands()
	if len(es) != int(expected+1) || len(cs) != int(expected+1) {
		return session.ErrInvalid
	}
	b := s.Board()
	x, e := r.c.UpdateOne(ctx, bson.M{"_id": s.ID(), "revision": expected}, bson.M{"$set": bson.M{"houses": b.Houses(), "captured": b.Captured(), "gameOver": b.GameOver(), "winner": b.Winner(), "turn": int(s.Turn()), "deadline": s.Deadline(), "status": s.Status(), "revision": s.Revision()}, "$push": bson.M{"events": es[len(es)-1], "commands": cs[len(cs)-1]}})
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
func toDoc(s session.Session) doc {
	b := s.Board()
	return doc{ID: s.ID(), RoomRef: s.RoomRef(), Players: s.Players(), Houses: b.Houses(), Captured: b.Captured(), GameOver: b.GameOver(), Winner: b.Winner(), Turn: int(s.Turn()), MoveWindow: s.MoveWindow(), Deadline: s.Deadline(), Status: s.Status(), Revision: s.Revision(), Events: s.Events(), Commands: s.Commands()}
}
func owarePlayer(value int) oware.Player { return oware.Player(value) }
