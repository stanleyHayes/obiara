package mongodb

import (
	"context"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database} }
func (r *Repository) policies() *mongo.Collection {
	return r.database.Collection("cloth_gate_policies")
}
func (r *Repository) events() *mongo.Collection { return r.database.Collection("cloth_gate_events") }

type capabilityDocument struct {
	ReviewerKey string `bson:"reviewerKey"`
	QuestionKey string `bson:"questionKey"`
	MaterialKey string `bson:"materialKey"`
}
type grantDocument struct {
	MemberKey  string             `bson:"memberKey"`
	Capability capabilityDocument `bson:"capability"`
}
type policyDocument struct {
	ID       string          `bson:"_id"`
	Version  string          `bson:"version"`
	Members  [2]string       `bson:"members"`
	Grants   []grantDocument `bson:"grants"`
	Revision uint64          `bson:"revision"`
}
type eventDocument struct {
	PolicyID    string             `bson:"policyId"`
	Sequence    uint64             `bson:"sequence"`
	Action      string             `bson:"action"`
	CommandID   string             `bson:"commandId"`
	ActorKey    string             `bson:"actorKey"`
	Fingerprint string             `bson:"fingerprint"`
	Capability  capabilityDocument `bson:"capability"`
	At          time.Time          `bson:"at"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("cloth_gate_command_unique")}, {Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetUnique(true).SetName("cloth_gate_sequence_unique")}}
	_, err := r.events().Indexes().CreateMany(ctx, models)
	return err
}
func (r *Repository) Create(ctx context.Context, p domain.Policy) error {
	s := p.State()
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		if _, e := r.policies().InsertOne(tx, toPolicy(s)); e != nil {
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
func (r *Repository) Find(ctx context.Context, id string) (domain.Policy, error) {
	var document policyDocument
	if err := r.policies().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		return domain.Policy{}, err
	}
	cursor, err := r.events().Find(ctx, bson.M{"policyId": id}, options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}))
	if err != nil {
		return domain.Policy{}, err
	}
	defer cursor.Close(ctx)
	var docs []eventDocument
	if err = cursor.All(ctx, &docs); err != nil {
		return domain.Policy{}, err
	}
	events := make([]domain.Event, 0, len(docs))
	for _, doc := range docs {
		events = append(events, fromEvent(doc))
	}
	grants := make([]domain.MemberGrant, 0, len(document.Grants))
	for _, g := range document.Grants {
		grants = append(grants, domain.MemberGrant{MemberKey: g.MemberKey, Capability: fromCapability(g.Capability)})
	}
	return domain.Rehydrate(domain.State{ID: document.ID, Version: document.Version, Members: document.Members, Grants: grants, Revision: document.Revision, Events: events})
}
func (r *Repository) FindByCommand(ctx context.Context, commandID string) (domain.Policy, error) {
	var event eventDocument
	if err := r.events().FindOne(ctx, bson.M{"commandId": commandID}).Decode(&event); err != nil {
		return domain.Policy{}, err
	}
	return r.Find(ctx, event.PolicyID)
}
func (r *Repository) Append(ctx context.Context, p domain.Policy, expected uint64, commandID string) error {
	s := p.State()
	event := s.Events[len(s.Events)-1]
	grants := make([]grantDocument, 0, len(s.Grants))
	for _, g := range s.Grants {
		grants = append(grants, grantDocument{g.MemberKey, toCapability(g.Capability)})
	}
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, e := r.policies().UpdateOne(tx, bson.M{"_id": s.ID, "revision": expected}, bson.M{"$set": bson.M{"grants": grants, "revision": s.Revision}})
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
func toCapability(c domain.Capability) capabilityDocument {
	return capabilityDocument{c.ReviewerKey, c.QuestionKey, c.MaterialKey}
}
func fromCapability(c capabilityDocument) domain.Capability {
	return domain.Capability{ReviewerKey: c.ReviewerKey, QuestionKey: c.QuestionKey, MaterialKey: c.MaterialKey}
}
func toPolicy(s domain.State) policyDocument {
	grants := make([]grantDocument, 0, len(s.Grants))
	for _, g := range s.Grants {
		grants = append(grants, grantDocument{g.MemberKey, toCapability(g.Capability)})
	}
	return policyDocument{s.ID, s.Version, s.Members, grants, s.Revision}
}
func toEvent(id string, e domain.Event) eventDocument {
	return eventDocument{id, e.Sequence, string(e.Action), e.CommandID, e.ActorKey, e.Fingerprint, toCapability(e.Capability), e.At}
}
func fromEvent(e eventDocument) domain.Event {
	return domain.Event{Sequence: e.Sequence, Action: domain.Action(e.Action), CommandID: e.CommandID, ActorKey: e.ActorKey, Fingerprint: e.Fingerprint, Capability: fromCapability(e.Capability), At: e.At}
}
