package mongodb

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func New(db *mongo.Database) *Repository { return &Repository{db.Collection("matchmaker_engagements")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, e := r.c.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "appliedids", Value: 1}}, Options: options.Index().SetUnique(true).SetName("matchmaker_command_unique")}, {Keys: bson.D{{Key: "sealedproposalref", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true).SetName("proposal_unique")}})
	return e
}
func (r *Repository) Create(ctx context.Context, x domain.Engagement) error {
	s := x.State()
	_, e := r.c.InsertOne(ctx, s)
	return r.dupe(ctx, s.AppliedIDs[0], e)
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Engagement, error) {
	var s domain.State
	e := r.c.FindOne(ctx, bson.M{"id": id}).Decode(&s)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return domain.Engagement{}, application.ErrNotFound
	}
	if e != nil {
		return domain.Engagement{}, e
	}
	return domain.Rehydrate(s)
}
func (r *Repository) FindByCommand(ctx context.Context, command string) (domain.Engagement, error) {
	var state domain.State
	err := r.c.FindOne(ctx, bson.M{"appliedids": command}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Engagement{}, application.ErrNotFound
	}
	if err != nil {
		return domain.Engagement{}, err
	}
	return domain.Rehydrate(state)
}
func (r *Repository) ListForMember(ctx context.Context, memberKey string) ([]domain.Engagement, error) {
	cursor, e := r.c.Find(ctx, bson.M{"memberkey": memberKey}, options.Find().SetSort(bson.D{{Key: "bookedat", Value: -1}}))
	if e != nil {
		return nil, e
	}
	defer cursor.Close(ctx)
	var states []domain.State
	if e = cursor.All(ctx, &states); e != nil {
		return nil, e
	}
	out := make([]domain.Engagement, 0, len(states))
	for _, state := range states {
		engagement, err := domain.Rehydrate(state)
		if err != nil {
			return nil, err
		}
		out = append(out, engagement)
	}
	return out, nil
}
func (r *Repository) Save(ctx context.Context, x domain.Engagement, expected uint64, command string) error {
	s := x.State()
	if len(s.Events) != int(expected+1) {
		return domain.ErrInvalid
	}
	result, e := r.c.UpdateOne(ctx, bson.M{"id": s.ID, "revision": expected}, bson.M{"$set": bson.M{"sealedproposalref": s.SealedProposalRef, "memberconsented": s.MemberConsented, "candidateconsented": s.CandidateConsented, "exposed": s.Exposed, "completed": s.Completed, "revision": s.Revision}, "$push": bson.M{"events": s.Events[len(s.Events)-1], "appliedids": command}})
	if e != nil {
		return r.dupe(ctx, command, e)
	}
	if result.MatchedCount == 0 {
		if r.c.FindOne(ctx, bson.M{"appliedids": command}).Err() == nil {
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
	if r.c.FindOne(ctx, bson.M{"appliedids": id}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
