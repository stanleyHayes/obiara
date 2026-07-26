package mongodb

import (
	"context"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/control/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/control/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database} }
func (r *Repository) heads() *mongo.Collection           { return r.database.Collection("fire_controls") }
func (r *Repository) events() *mongo.Collection          { return r.database.Collection("fire_control_events") }

type memberDocument struct {
	Key     string `bson:"key"`
	Role    string `bson:"role"`
	Muted   bool   `bson:"muted"`
	Ejected bool   `bson:"ejected"`
}
type headDocument struct {
	ID       string           `bson:"_id"`
	FireKey  string           `bson:"fireKey"`
	Members  []memberDocument `bson:"members"`
	Revision uint64           `bson:"revision"`
}
type eventDocument struct {
	ControlID   string    `bson:"controlId"`
	Sequence    uint64    `bson:"sequence"`
	Action      string    `bson:"action"`
	CommandID   string    `bson:"commandId"`
	ActorKey    string    `bson:"actorKey"`
	TargetKey   string    `bson:"targetKey"`
	ReasonCode  string    `bson:"reasonCode"`
	Fingerprint string    `bson:"fingerprint"`
	At          time.Time `bson:"at"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("fire_control_command_unique")}, {Keys: bson.D{{Key: "controlId", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetUnique(true).SetName("fire_control_sequence_unique")}}
	_, err := r.events().Indexes().CreateMany(ctx, models)
	return err
}
func (r *Repository) Create(ctx context.Context, f domain.Fire) error {
	s := f.State()
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
func (r *Repository) Find(ctx context.Context, id string) (domain.Fire, error) {
	var head headDocument
	if err := r.heads().FindOne(ctx, bson.M{"_id": id}).Decode(&head); err != nil {
		return domain.Fire{}, err
	}
	cursor, err := r.events().Find(ctx, bson.M{"controlId": id}, options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}))
	if err != nil {
		return domain.Fire{}, err
	}
	defer cursor.Close(ctx)
	var docs []eventDocument
	if err = cursor.All(ctx, &docs); err != nil {
		return domain.Fire{}, err
	}
	members := make([]domain.Member, 0, len(head.Members))
	for _, m := range head.Members {
		members = append(members, domain.Member{Key: m.Key, Role: domain.Role(m.Role), Muted: m.Muted, Ejected: m.Ejected})
	}
	events := make([]domain.Event, 0, len(docs))
	for _, e := range docs {
		events = append(events, fromEvent(e))
	}
	return domain.Rehydrate(domain.State{ID: head.ID, FireKey: head.FireKey, Members: members, Revision: head.Revision, Events: events})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Fire, error) {
	var event eventDocument
	if err := r.events().FindOne(ctx, bson.M{"commandId": id}).Decode(&event); err != nil {
		return domain.Fire{}, err
	}
	return r.Find(ctx, event.ControlID)
}
func (r *Repository) Append(ctx context.Context, f domain.Fire, expected uint64, commandID string) error {
	s := f.State()
	event := s.Events[len(s.Events)-1]
	members := make([]memberDocument, 0, len(s.Members))
	for _, m := range s.Members {
		members = append(members, memberDocument{m.Key, string(m.Role), m.Muted, m.Ejected})
	}
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, e := r.heads().UpdateOne(tx, bson.M{"_id": s.ID, "revision": expected}, bson.M{"$set": bson.M{"members": members, "revision": s.Revision}})
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
func toHead(s domain.State) headDocument {
	members := make([]memberDocument, 0, len(s.Members))
	for _, m := range s.Members {
		members = append(members, memberDocument{m.Key, string(m.Role), m.Muted, m.Ejected})
	}
	return headDocument{s.ID, s.FireKey, members, s.Revision}
}
func toEvent(id string, e domain.Event) eventDocument {
	return eventDocument{id, e.Sequence, string(e.Action), e.CommandID, e.ActorKey, e.TargetKey, e.ReasonCode, e.Fingerprint, e.At}
}
func fromEvent(e eventDocument) domain.Event {
	return domain.Event{Sequence: e.Sequence, Action: domain.Action(e.Action), CommandID: e.CommandID, ActorKey: e.ActorKey, TargetKey: e.TargetKey, ReasonCode: e.ReasonCode, Fingerprint: e.Fingerprint, At: e.At}
}
