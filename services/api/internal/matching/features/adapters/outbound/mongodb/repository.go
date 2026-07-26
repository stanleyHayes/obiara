package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/application"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	definitions *mongo.Collection
	grants      *mongo.Collection
	decisions   *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{definitions: db.Collection("matching_feature_definitions"), grants: db.Collection("matching_feature_grants"), decisions: db.Collection("matching_feature_decisions")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.definitions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "key", Value: 1}, {Key: "version", Value: 1}}, Options: options.Index().SetUnique(true).SetName("feature_version_unique")},
		{Keys: bson.D{{Key: "key", Value: 1}, {Key: "version", Value: -1}}, Options: options.Index().SetName("feature_current")},
	}); err != nil {
		return err
	}
	if _, err := r.grants.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "memberKey", Value: 1}, {Key: "featureKey", Value: 1}}, Options: options.Index().SetUnique(true).SetName("member_feature_unique")},
		{Keys: bson.D{{Key: "commands.id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("feature_command_unique")},
		{Keys: bson.D{{Key: "memberKey", Value: 1}, {Key: "status", Value: 1}}, Options: options.Index().SetName("member_effective")},
	}); err != nil {
		return err
	}
	_, err := r.decisions.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "pair", Value: 1}, {Key: "evaluatedAt", Value: -1}}, Options: options.Index().SetName("pair_decisions")})
	return err
}
func (r *Repository) Put(ctx context.Context, d domain.Definition) error {
	_, err := r.definitions.InsertOne(ctx, d)
	if mongo.IsDuplicateKeyError(err) {
		return application.ErrConflict
	}
	return err
}
func (r *Repository) Find(ctx context.Context, member, feature string) (domain.Grant, error) {
	return r.findGrant(ctx, bson.M{"memberKey": member, "featureKey": feature})
}

// FindDefinition resolves one immutable allowlisted version.
func (r *Repository) FindDefinition(ctx context.Context, key string, version uint64) (domain.Definition, error) {
	var d domain.Definition
	err := r.definitions.FindOne(ctx, bson.M{"key": key, "version": version}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Definition{}, application.ErrNotFound
	}
	return d, err
}
func (r *Repository) Current(ctx context.Context, key string) (domain.Definition, error) {
	var d domain.Definition
	err := r.definitions.FindOne(ctx, bson.M{"key": key}, options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Definition{}, application.ErrNotFound
	}
	return d, err
}
func (r *Repository) ListCurrent(ctx context.Context) ([]domain.Definition, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$sort", Value: bson.D{{Key: "key", Value: 1}, {Key: "version", Value: -1}}}},
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$key"}, {Key: "value", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}}}}},
		{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$value"}}}},
	}
	cur, err := r.definitions.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Definition
	return out, cur.All(ctx, &out)
}

type grantDoc struct {
	ID             string           `bson:"_id"`
	MemberKey      string           `bson:"memberKey"`
	FeatureKey     string           `bson:"featureKey"`
	Purpose        string           `bson:"purpose"`
	FeatureVersion uint64           `bson:"featureVersion"`
	GrantVersion   uint64           `bson:"grantVersion"`
	Status         domain.Status    `bson:"status"`
	GrantedAt      time.Time        `bson:"grantedAt"`
	WithdrawnAt    time.Time        `bson:"withdrawnAt,omitempty"`
	Revision       uint64           `bson:"revision"`
	Events         []domain.Event   `bson:"events"`
	Commands       []domain.Applied `bson:"commands"`
}

func grantID(member, feature string) string { return member + ":" + feature }
func toGrantDoc(g domain.Grant) grantDoc {
	return grantDoc{ID: grantID(g.MemberKey(), g.FeatureKey()), MemberKey: g.MemberKey(), FeatureKey: g.FeatureKey(), Purpose: g.Purpose(), FeatureVersion: g.FeatureVersion(), GrantVersion: g.GrantVersion(), Status: g.Status(), GrantedAt: g.GrantedAt(), WithdrawnAt: g.WithdrawnAt(), Revision: g.Revision(), Events: g.Events(), Commands: g.Commands()}
}
func (r *Repository) Create(ctx context.Context, g domain.Grant) error {
	_, err := r.grants.InsertOne(ctx, toGrantDoc(g))
	return r.duplicate(ctx, g.Commands()[0].ID, err)
}
func (r *Repository) findGrant(ctx context.Context, filter bson.M) (domain.Grant, error) {
	var d grantDoc
	if err := r.grants.FindOne(ctx, filter).Decode(&d); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Grant{}, application.ErrNotFound
		}
		return domain.Grant{}, err
	}
	return domain.RehydrateGrant(domain.GrantState{MemberKey: d.MemberKey, FeatureKey: d.FeatureKey, Purpose: d.Purpose, FeatureVersion: d.FeatureVersion, GrantVersion: d.GrantVersion, Status: d.Status, GrantedAt: d.GrantedAt, WithdrawnAt: d.WithdrawnAt, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
}
func (r *Repository) FindByCommand(ctx context.Context, id string) (domain.Grant, error) {
	return r.findGrant(ctx, bson.M{"commands.id": id})
}
func (r *Repository) ListEffective(ctx context.Context, member string) ([]domain.Grant, error) {
	cur, err := r.grants.Find(ctx, bson.M{"memberKey": member, "status": domain.StatusGranted})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var docs []grantDoc
	if err = cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]domain.Grant, 0, len(docs))
	for _, d := range docs {
		g, e := domain.RehydrateGrant(domain.GrantState{MemberKey: d.MemberKey, FeatureKey: d.FeatureKey, Purpose: d.Purpose, FeatureVersion: d.FeatureVersion, GrantVersion: d.GrantVersion, Status: d.Status, GrantedAt: d.GrantedAt, WithdrawnAt: d.WithdrawnAt, Revision: d.Revision, Events: d.Events, Commands: d.Commands})
		if e != nil {
			return nil, e
		}
		out = append(out, g)
	}
	return out, nil
}
func (r *Repository) Append(ctx context.Context, g domain.Grant, expected uint64, command string) error {
	events, commands := g.Events(), g.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalid
	}
	result, err := r.grants.UpdateOne(ctx, bson.M{"_id": grantID(g.MemberKey(), g.FeatureKey()), "revision": expected}, bson.M{
		"$set":  bson.M{"purpose": g.Purpose(), "featureVersion": g.FeatureVersion(), "grantVersion": g.GrantVersion(), "status": g.Status(), "grantedAt": g.GrantedAt(), "withdrawnAt": g.WithdrawnAt(), "revision": g.Revision()},
		"$push": bson.M{"events": events[len(events)-1], "commands": commands[len(commands)-1]},
	})
	if err != nil {
		return r.duplicate(ctx, command, err)
	}
	if result.MatchedCount == 0 {
		if r.grants.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
			return application.ErrApplied
		}
		return application.ErrConflict
	}
	return nil
}
func (r *Repository) duplicate(ctx context.Context, command string, err error) error {
	if err == nil || !mongo.IsDuplicateKeyError(err) {
		return err
	}
	if r.grants.FindOne(ctx, bson.M{"commands.id": command}).Err() == nil {
		return application.ErrApplied
	}
	return application.ErrConflict
}
func (r *Repository) CreateDecision(ctx context.Context, d domain.Decision) error {
	_, err := r.decisions.InsertOne(ctx, bson.M{"_id": d.ID, "pair": d.Pair, "features": d.Features, "evaluatedAt": d.EvaluatedAt})
	if mongo.IsDuplicateKeyError(err) {
		return application.ErrConflict
	}
	return err
}
func (r *Repository) FindDecision(ctx context.Context, id string) (domain.Decision, error) {
	var raw struct {
		ID          string                  `bson:"_id"`
		Pair        []string                `bson:"pair"`
		Features    []domain.EnabledFeature `bson:"features"`
		EvaluatedAt time.Time               `bson:"evaluatedAt"`
	}
	if err := r.decisions.FindOne(ctx, bson.M{"_id": id}).Decode(&raw); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Decision{}, application.ErrNotFound
		}
		return domain.Decision{}, err
	}
	return domain.NewDecision(raw.ID, raw.Pair[0], raw.Pair[1], raw.Features, raw.EvaluatedAt)
}
