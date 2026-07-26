package mongodb

import (
	"context"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database} }
func (r *Repository) heads() *mongo.Collection           { return r.database.Collection("cloth_ceremonies") }
func (r *Repository) events() *mongo.Collection {
	return r.database.Collection("cloth_ceremony_events")
}

type announcementDocument struct {
	DestinationKey string   `bson:"destinationKey"`
	Kind           string   `bson:"kind"`
	Consents       []string `bson:"consents"`
	Published      bool     `bson:"published"`
}
type headDocument struct {
	ID            string                `bson:"_id"`
	Members       [2]string             `bson:"members"`
	Confirmations []string              `bson:"confirmations"`
	Announcement  *announcementDocument `bson:"announcement,omitempty"`
	Revision      uint64                `bson:"revision"`
}
type eventDocument struct {
	CeremonyID     string    `bson:"ceremonyId"`
	Sequence       uint64    `bson:"sequence"`
	Action         string    `bson:"action"`
	CommandID      string    `bson:"commandId"`
	ActorKey       string    `bson:"actorKey"`
	Fingerprint    string    `bson:"fingerprint"`
	DestinationKey string    `bson:"destinationKey,omitempty"`
	At             time.Time `bson:"at"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("ceremony_command_unique")}, {Keys: bson.D{{Key: "ceremonyId", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetUnique(true).SetName("ceremony_sequence_unique")}}
	_, err := r.events().Indexes().CreateMany(ctx, models)
	return err
}
func (r *Repository) Create(ctx context.Context, c domain.Ceremony) error {
	s := c.State()
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		if _, e := r.heads().InsertOne(tx, toHead(s)); e != nil {
			return e
		}
		_, e := r.events().InsertOne(tx, toEvent(s.ID, s.Events[0]))
		return e
	})
	if apimongo.IsDuplicateKey(err) {
		return application.ErrCommandApplied
	}
	return err
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Ceremony, error) {
	var head headDocument
	if err := r.heads().FindOne(ctx, bson.M{"_id": id}).Decode(&head); err != nil {
		return domain.Ceremony{}, err
	}
	cursor, err := r.events().Find(ctx, bson.M{"ceremonyId": id}, options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}))
	if err != nil {
		return domain.Ceremony{}, err
	}
	defer cursor.Close(ctx)
	var docs []eventDocument
	if err = cursor.All(ctx, &docs); err != nil {
		return domain.Ceremony{}, err
	}
	events := make([]domain.Event, 0, len(docs))
	for _, d := range docs {
		events = append(events, fromEvent(d))
	}
	var announcement *domain.Announcement
	if head.Announcement != nil {
		announcement = &domain.Announcement{DestinationKey: head.Announcement.DestinationKey, Kind: head.Announcement.Kind, Consents: head.Announcement.Consents, Published: head.Announcement.Published}
	}
	return domain.Rehydrate(domain.State{ID: head.ID, Members: head.Members, Confirmations: head.Confirmations, Announcement: announcement, Revision: head.Revision, Events: events})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Ceremony, error) {
	var event eventDocument
	if err := r.events().FindOne(ctx, bson.M{"commandId": id}).Decode(&event); err != nil {
		return domain.Ceremony{}, err
	}
	return r.Find(ctx, event.CeremonyID)
}
func (r *Repository) Append(ctx context.Context, c domain.Ceremony, expected uint64, commandID string) error {
	s := c.State()
	event := s.Events[len(s.Events)-1]
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, e := r.heads().UpdateOne(tx, bson.M{"_id": s.ID, "revision": expected}, bson.M{"$set": bson.M{"confirmations": s.Confirmations, "announcement": toAnnouncement(s.Announcement), "revision": s.Revision}})
		if e != nil {
			return e
		}
		if result.MatchedCount == 0 {
			return application.ErrConcurrentChange
		}
		_, e = r.events().InsertOne(tx, toEvent(s.ID, event))
		return e
	})
	if apimongo.IsDuplicateKey(err) {
		return application.ErrCommandApplied
	}
	return err
}
func toAnnouncement(a *domain.Announcement) *announcementDocument {
	if a == nil {
		return nil
	}
	return &announcementDocument{a.DestinationKey, a.Kind, a.Consents, a.Published}
}
func toHead(s domain.State) headDocument {
	return headDocument{s.ID, s.Members, s.Confirmations, toAnnouncement(s.Announcement), s.Revision}
}
func toEvent(id string, e domain.Event) eventDocument {
	return eventDocument{id, e.Sequence, string(e.Action), e.CommandID, e.ActorKey, e.Fingerprint, e.DestinationKey, e.At}
}
func fromEvent(e eventDocument) domain.Event {
	return domain.Event{Sequence: e.Sequence, Action: domain.Action(e.Action), CommandID: e.CommandID, ActorKey: e.ActorKey, Fingerprint: e.Fingerprint, DestinationKey: e.DestinationKey, At: e.At}
}
