package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ db *mongo.Database }

func NewRepository(db *mongo.Database) *Repository { return &Repository{db} }

type mediaDoc struct {
	Asset, Transcript, Type string `bson:",omitempty"`
	Duration                int64
}
type auditDoc struct {
	Command, Actor string
	At             time.Time
}
type doc struct {
	ID                                      string `bson:"_id"`
	Circle, Author, Content, Kind           string
	Media                                   mediaDoc
	Starts, Ends, Created, Expires, Deleted time.Time
	Revision                                uint64
	Audit                                   []auditDoc
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.db.Collection("circle_room_entries").Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "circle", Value: 1}, {Key: "expires", Value: 1}}}, {Keys: bson.D{{Key: "audit.command", Value: 1}}, Options: options.Index().SetUnique(true)}})
	return e
}
func (r *Repository) Create(ctx context.Context, e domain.Entry) (domain.Entry, bool, error) {
	_, x := r.db.Collection("circle_room_entries").InsertOne(ctx, toDoc(e))
	if x == nil {
		return e, false, nil
	}
	if !mongo.IsDuplicateKeyError(x) {
		return domain.Entry{}, false, x
	}
	var d doc
	if y := r.db.Collection("circle_room_entries").FindOne(ctx, bson.M{"audit.command": e.Audit()[0].CommandID()}).Decode(&d); y != nil {
		return domain.Entry{}, false, y
	}
	v, y := toDomain(d)
	return v, true, y
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Entry, error) {
	var d doc
	if e := r.db.Collection("circle_room_entries").FindOne(ctx, bson.M{"_id": id}).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return domain.Entry{}, application.ErrNotFound
		}
		return domain.Entry{}, e
	}
	return toDomain(d)
}
func (r *Repository) Delete(ctx context.Context, e domain.Entry, v uint64, c string) error {
	d := toDoc(e)
	x, er := r.db.Collection("circle_room_entries").UpdateOne(ctx, bson.M{"_id": e.ID(), "revision": v, "audit.command": bson.M{"$ne": c}}, bson.M{"$set": bson.M{"deleted": e.DeletedAt(), "revision": e.Revision()}, "$push": bson.M{"audit": d.Audit[len(d.Audit)-1]}})
	if er != nil {
		return er
	}
	if x.MatchedCount == 0 {
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) List(ctx context.Context, c string, at time.Time, l int) ([]domain.Entry, error) {
	cur, e := r.db.Collection("circle_room_entries").Find(ctx, bson.M{"circle": c, "deleted": bson.M{"$eq": time.Time{}}, "expires": bson.M{"$gt": at}}, options.Find().SetLimit(int64(l)).SetSort(bson.D{{Key: "created", Value: -1}}))
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var ds []doc
	if e = cur.All(ctx, &ds); e != nil {
		return nil, e
	}
	out := make([]domain.Entry, 0, len(ds))
	for _, d := range ds {
		v, x := toDomain(d)
		if x != nil {
			return nil, x
		}
		out = append(out, v)
	}
	return out, nil
}
func toDoc(e domain.Entry) doc {
	a := make([]auditDoc, 0, len(e.Audit()))
	for _, x := range e.Audit() {
		a = append(a, auditDoc{x.CommandID(), x.ActorKey(), x.At()})
	}
	return doc{e.ID(), e.CircleID(), e.AuthorKey(), e.ContentRef(), string(e.Kind()), mediaDoc{e.Media().AssetID(), e.Media().TranscriptID(), e.Media().ContentType(), int64(e.Media().Duration())}, e.StartsAt(), e.EndsAt(), e.CreatedAt(), e.ExpiresAt(), e.DeletedAt(), e.Revision(), a}
}
func toDomain(d doc) (domain.Entry, error) {
	m := domain.MediaRef{}
	var e error
	if d.Kind == string(domain.KindVoice) {
		m, e = domain.NewMediaRef(d.Media.Asset, d.Media.Transcript, d.Media.Type, time.Duration(d.Media.Duration))
		if e != nil {
			return domain.Entry{}, e
		}
	}
	a := make([]domain.Audit, 0, len(d.Audit))
	for _, x := range d.Audit {
		v, y := domain.NewAudit(x.Command, x.Actor, x.At)
		if y != nil {
			return domain.Entry{}, y
		}
		a = append(a, v)
	}
	return domain.Rehydrate(domain.Params{ID: d.ID, CircleID: d.Circle, AuthorKey: d.Author, ContentRef: d.Content, CommandID: d.Audit[0].Command, Kind: domain.Kind(d.Kind), Media: m, StartsAt: d.Starts, EndsAt: d.Ends, CreatedAt: d.Created, ExpiresAt: d.Expires}, d.Deleted, d.Revision, a)
}
