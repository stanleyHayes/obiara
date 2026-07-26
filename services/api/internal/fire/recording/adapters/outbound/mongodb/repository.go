package mongodb

import (
	"context"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/recording/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/recording/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database} }
func (r *Repository) heads() *mongo.Collection {
	return r.database.Collection("fire_recording_policies")
}
func (r *Repository) events() *mongo.Collection {
	return r.database.Collection("fire_recording_events")
}

type proposalDocument struct {
	Purpose        string `bson:"purpose"`
	RetentionNanos int64  `bson:"retentionNanos"`
}
type headDocument struct {
	ID           string            `bson:"_id"`
	FireKey      string            `bson:"fireKey"`
	HostKey      string            `bson:"hostKey"`
	Participants []string          `bson:"participants"`
	Consents     []string          `bson:"consents"`
	Proposal     *proposalDocument `bson:"proposal,omitempty"`
	Active       bool              `bson:"active"`
	RecordingRef string            `bson:"recordingEnvelopeRef,omitempty"`
	Revision     uint64            `bson:"revision"`
}
type eventDocument struct {
	PolicyID       string    `bson:"policyId"`
	Sequence       uint64    `bson:"sequence"`
	Action         string    `bson:"action"`
	CommandID      string    `bson:"commandId"`
	ActorKey       string    `bson:"actorKey"`
	SubjectKey     string    `bson:"subjectKey"`
	Fingerprint    string    `bson:"fingerprint"`
	Purpose        string    `bson:"purpose,omitempty"`
	RetentionNanos int64     `bson:"retentionNanos,omitempty"`
	At             time.Time `bson:"at"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("recording_command_unique")}, {Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetUnique(true).SetName("recording_sequence_unique")}}
	_, err := r.events().Indexes().CreateMany(ctx, models)
	return err
}
func (r *Repository) Create(ctx context.Context, p domain.Policy) error {
	s := p.State()
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
func (r *Repository) Find(ctx context.Context, id string) (domain.Policy, error) {
	var h headDocument
	if err := r.heads().FindOne(ctx, bson.M{"_id": id}).Decode(&h); err != nil {
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
	for _, e := range docs {
		events = append(events, fromEvent(e))
	}
	var proposal *domain.Proposal
	if h.Proposal != nil {
		proposal = &domain.Proposal{Purpose: domain.Purpose(h.Proposal.Purpose), Retention: time.Duration(h.Proposal.RetentionNanos)}
	}
	return domain.Rehydrate(domain.State{ID: h.ID, FireKey: h.FireKey, HostKey: h.HostKey, Participants: h.Participants, Consents: h.Consents, Proposal: proposal, Active: h.Active, RecordingRef: h.RecordingRef, Revision: h.Revision, Events: events})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Policy, error) {
	var e eventDocument
	if err := r.events().FindOne(ctx, bson.M{"commandId": id}).Decode(&e); err != nil {
		return domain.Policy{}, err
	}
	return r.Find(ctx, e.PolicyID)
}
func (r *Repository) Append(ctx context.Context, p domain.Policy, expected uint64, commandID string) error {
	s := p.State()
	e := s.Events[len(s.Events)-1]
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, x := r.heads().UpdateOne(tx, bson.M{"_id": s.ID, "revision": expected}, bson.M{"$set": toHeadSet(s)})
		if x != nil {
			return x
		}
		if result.MatchedCount == 0 {
			return application.ErrConcurrentChange
		}
		_, x = r.events().InsertOne(tx, toEvent(s.ID, e))
		return x
	})
	if apimongo.IsDuplicateKey(err) {
		return application.ErrCommandApplied
	}
	return err
}
func proposalDoc(p *domain.Proposal) *proposalDocument {
	if p == nil {
		return nil
	}
	return &proposalDocument{string(p.Purpose), int64(p.Retention)}
}
func toHead(s domain.State) headDocument {
	return headDocument{s.ID, s.FireKey, s.HostKey, s.Participants, s.Consents, proposalDoc(s.Proposal), s.Active, s.RecordingRef, s.Revision}
}
func toHeadSet(s domain.State) bson.M {
	return bson.M{"participants": s.Participants, "consents": s.Consents, "proposal": proposalDoc(s.Proposal), "active": s.Active, "recordingEnvelopeRef": s.RecordingRef, "revision": s.Revision}
}
func toEvent(id string, e domain.Event) eventDocument {
	return eventDocument{id, e.Sequence, string(e.Action), e.CommandID, e.ActorKey, e.SubjectKey, e.Fingerprint, string(e.Purpose), int64(e.Retention), e.At}
}
func fromEvent(e eventDocument) domain.Event {
	return domain.Event{Sequence: e.Sequence, Action: domain.Action(e.Action), CommandID: e.CommandID, ActorKey: e.ActorKey, SubjectKey: e.SubjectKey, Fingerprint: e.Fingerprint, Purpose: domain.Purpose(e.Purpose), Retention: time.Duration(e.RetentionNanos), At: e.At}
}
