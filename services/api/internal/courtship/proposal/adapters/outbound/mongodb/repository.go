package mongodb

import (
	"context"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repository struct{ database *mongo.Database }

func NewRepository(database *mongo.Database) *Repository { return &Repository{database} }
func (r *Repository) proposals() *mongo.Collection {
	return r.database.Collection("courtship_proposals")
}
func (r *Repository) events() *mongo.Collection {
	return r.database.Collection("courtship_proposal_events")
}

type proposalDocument struct {
	ID           string    `bson:"_id"`
	Kind         string    `bson:"kind"`
	SenderKey    string    `bson:"senderKey"`
	RecipientKey string    `bson:"recipientKey"`
	DetailKey    string    `bson:"encryptedDetailRef"`
	Status       string    `bson:"status"`
	ExpiresAt    time.Time `bson:"expiresAt"`
	Revision     uint64    `bson:"revision"`
}
type eventDocument struct {
	ProposalID  string    `bson:"proposalId"`
	Sequence    uint64    `bson:"sequence"`
	Action      string    `bson:"action"`
	CommandID   string    `bson:"commandId"`
	ActorKey    string    `bson:"actorKey"`
	Fingerprint string    `bson:"fingerprint"`
	At          time.Time `bson:"at"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("proposal_command_unique")}, {Keys: bson.D{{Key: "proposalId", Value: 1}, {Key: "sequence", Value: 1}}, Options: options.Index().SetUnique(true).SetName("proposal_sequence_unique")}}
	_, err := r.events().Indexes().CreateMany(ctx, models)
	return err
}
func (r *Repository) Create(ctx context.Context, p domain.Proposal) error {
	s := p.State()
	return apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		if _, err := r.proposals().InsertOne(tx, toProposal(s)); err != nil {
			if apimongo.IsDuplicateKey(err) {
				return application.ErrCommandApplied
			}
			return err
		}
		_, err := r.events().InsertOne(tx, toEvent(s.ID, s.Events[0]))
		if apimongo.IsDuplicateKey(err) {
			return application.ErrCommandApplied
		}
		return err
	})
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Proposal, error) {
	var document proposalDocument
	if err := r.proposals().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		return domain.Proposal{}, err
	}
	cursor, err := r.events().Find(ctx, bson.M{"proposalId": id}, options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}))
	if err != nil {
		return domain.Proposal{}, err
	}
	defer cursor.Close(ctx)
	var docs []eventDocument
	if err = cursor.All(ctx, &docs); err != nil {
		return domain.Proposal{}, err
	}
	events := make([]domain.Event, 0, len(docs))
	for _, doc := range docs {
		events = append(events, fromEvent(doc))
	}
	return domain.Rehydrate(domain.State{ID: document.ID, Kind: domain.Type(document.Kind), SenderKey: document.SenderKey, RecipientKey: document.RecipientKey, DetailKey: document.DetailKey, Status: domain.Status(document.Status), ExpiresAt: document.ExpiresAt, Revision: document.Revision, Events: events})
}
func (r *Repository) FindByCommand(ctx context.Context, commandID string) (domain.Proposal, error) {
	var event eventDocument
	if err := r.events().FindOne(ctx, bson.M{"commandId": commandID}).Decode(&event); err != nil {
		return domain.Proposal{}, err
	}
	return r.Find(ctx, event.ProposalID)
}
func (r *Repository) Append(ctx context.Context, p domain.Proposal, expected uint64, commandID string) error {
	s := p.State()
	event := s.Events[len(s.Events)-1]
	err := apimongo.WithTransaction(ctx, r.database.Client(), func(tx context.Context) error {
		result, updateErr := r.proposals().UpdateOne(tx, bson.M{"_id": s.ID, "revision": expected, "status": string(domain.StatusPending)}, bson.M{"$set": bson.M{"status": string(s.Status), "revision": s.Revision}})
		if updateErr != nil {
			return updateErr
		}
		if result.MatchedCount == 0 {
			return application.ErrConcurrentChange
		}
		_, insertErr := r.events().InsertOne(tx, toEvent(s.ID, event))
		return insertErr
	})
	if apimongo.IsDuplicateKey(err) {
		return application.ErrCommandApplied
	}
	return err
}
func toProposal(s domain.State) proposalDocument {
	return proposalDocument{s.ID, string(s.Kind), s.SenderKey, s.RecipientKey, s.DetailKey, string(s.Status), s.ExpiresAt, s.Revision}
}
func toEvent(id string, e domain.Event) eventDocument {
	return eventDocument{id, e.Sequence, string(e.Action), e.CommandID, e.ActorKey, e.Fingerprint, e.At}
}
func fromEvent(e eventDocument) domain.Event {
	return domain.Event{Sequence: e.Sequence, Action: domain.Action(e.Action), CommandID: e.CommandID, ActorKey: e.ActorKey, Fingerprint: e.Fingerprint, At: e.At}
}
