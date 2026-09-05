// Package retention automates the binding retention table (E15-S08; Doc
// 09 §7; plan §15: retention automation with proof-of-deletion records).
// Policies are declarative and each run writes an immutable proof record
// per policy execution. Legal holds are never touched by this runner —
// they live outside retention automation by design.
package retention

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Action is the retention operation.
type Action string

const (
	ActionDelete     Action = "delete"
	ActionStripField Action = "strip_field"
	ActionAggregate  Action = "aggregate"
)

// Policy declares one retention rule over a collection.
type Policy struct {
	Name       string
	Collection string
	Action     Action
	DateField  string
	MaxAge     time.Duration
	// Fields lists the fields stripped for ActionStripField.
	Fields []string
}

// BindingPolicies is the Doc 09 §7 table for artifacts that exist today.
// Media-backed classes (liveness raw media 30d, room voice 180d, Gate
// packs 30d) activate when their persistence adapters land.
var BindingPolicies = []Policy{
	{
		Name: "analytics_pseudonymize_90d", Collection: "analytics_events",
		Action: ActionStripField, DateField: "occurredAt", MaxAge: 90 * 24 * time.Hour,
		Fields: []string{"subjectRef"},
	},
	{
		Name: "analytics_aggregate_13mo", Collection: "analytics_events",
		Action: ActionAggregate, DateField: "occurredAt", MaxAge: 13 * 30 * 24 * time.Hour,
	},
	{
		Name: "privacy_requests_completed_90d", Collection: "privacy_requests",
		Action: ActionDelete, DateField: "completedAt", MaxAge: 90 * 24 * time.Hour,
	},
	// National-ID data (M1-02). Until these three policies existed, a member
	// who verified kept their date of birth and both photographs of their
	// Ghana Card in Mongo permanently: the case Update path never touches
	// dateOfBirth, and identity_documents deliberately carries no TTL.
	//
	// The images are deleted rather than stripped, and at 90 days rather than
	// 30, because the document store's own comment gives the reason: a review
	// queue "can take days", and an image that expired mid-review would leave
	// a member unverifiable with nothing explaining why. Ninety days is far
	// past any plausible queue while still being an end date.
	{
		Name: "identity_documents_delete_90d", Collection: "identity_documents",
		Action: ActionDelete, DateField: "createdAt", MaxAge: 90 * 24 * time.Hour,
	},
	// The date of birth is stripped rather than the case deleted: the case is
	// the proof that the check happened and what it decided, and that proof
	// should outlive the personal data it was derived from. Thirty days after
	// a decision, nothing needs the birth date any more.
	{
		Name: "identity_dob_strip_decided_30d", Collection: "identity_verifications",
		Action: ActionStripField, DateField: "decidedAt", MaxAge: 30 * 24 * time.Hour,
		Fields: []string{"dateOfBirth"},
	},
	// A case that is never decided has no decidedAt, so the policy above can
	// never match it and its birth date would be kept forever. This is the
	// backstop for abandoned and indefinitely queued submissions, set well
	// beyond any review so it cannot strip one that is still being worked.
	{
		Name: "identity_dob_strip_stale_180d", Collection: "identity_verifications",
		Action: ActionStripField, DateField: "createdAt", MaxAge: 180 * 24 * time.Hour,
		Fields: []string{"dateOfBirth"},
	},
}

// Report is one run's outcome per policy.
type Report struct {
	Policy     string
	Collection string
	Matched    int
}

// Runner executes the policies in batches.
type Runner struct {
	database  *mongo.Database
	policies  []Policy
	clock     func() time.Time
	batchSize int
}

func NewRunner(database *mongo.Database, policies []Policy, clock func() time.Time) *Runner {
	return &Runner{database: database, policies: policies, clock: clock, batchSize: 500}
}

// RunOnce executes every policy once, writing a proof record per policy.
func (runner *Runner) RunOnce(ctx context.Context) ([]Report, error) {
	var reports []Report
	for _, policy := range runner.policies {
		matched, err := runner.apply(ctx, policy)
		if err != nil {
			return reports, fmt.Errorf("policy %s: %w", policy.Name, err)
		}
		if err := runner.recordProof(ctx, policy, matched); err != nil {
			return reports, fmt.Errorf("policy %s proof: %w", policy.Name, err)
		}
		reports = append(reports, Report{Policy: policy.Name, Collection: policy.Collection, Matched: matched})
	}
	return reports, nil
}

func (runner *Runner) apply(ctx context.Context, policy Policy) (int, error) {
	cutoff := runner.clock().UTC().Add(-policy.MaxAge)
	collection := runner.database.Collection(policy.Collection)
	filter := bson.M{policy.DateField: bson.M{"$exists": true, "$lt": cutoff}}

	switch policy.Action {
	case ActionDelete:
		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			return 0, err
		}
		return int(result.DeletedCount), nil

	case ActionStripField:
		unset := bson.M{}
		for _, field := range policy.Fields {
			unset[field] = ""
		}
		result, err := collection.UpdateMany(ctx, filter, bson.M{"$unset": unset})
		if err != nil {
			return 0, err
		}
		return int(result.ModifiedCount), nil

	case ActionAggregate:
		return runner.aggregate(ctx, collection, filter, policy)

	default:
		return 0, fmt.Errorf("unknown retention action %q", policy.Action)
	}
}

// aggregate rolls events into per-name-per-day counts before deleting them.
func (runner *Runner) aggregate(ctx context.Context, collection *mongo.Collection, filter bson.M, policy Policy) (int, error) {
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(bson.M{"name": 1, policy.DateField: 1}).SetLimit(int64(runner.batchSize)))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	counts := map[string]int{}
	processed := 0
	for cursor.Next(ctx) {
		var document struct {
			Name string    `bson:"name"`
			At   time.Time `bson:"occurredAt"`
		}
		if err := cursor.Decode(&document); err != nil {
			return processed, err
		}
		key := document.Name + "|" + document.At.UTC().Format("2006-01-02")
		counts[key]++
		processed++
	}
	if err := cursor.Err(); err != nil {
		return processed, err
	}

	aggregates := runner.database.Collection("analytics_aggregates")
	for key, count := range counts {
		if _, err := aggregates.UpdateOne(ctx,
			bson.M{"_id": key},
			bson.M{"$inc": bson.M{"count": count}, "$set": bson.M{"aggregatedAt": runner.clock().UTC()}},
			options.UpdateOne().SetUpsert(true)); err != nil {
			return processed, err
		}
	}

	if processed > 0 {
		if _, err := collection.DeleteMany(ctx, filter); err != nil {
			return processed, err
		}
	}
	return processed, nil
}

type proofDocument struct {
	Policy      string    `bson:"_id"`
	Collection  string    `bson:"collection"`
	Action      string    `bson:"action"`
	Matched     int       `bson:"matched"`
	ProcessedAt time.Time `bson:"processedAt"`
}

// recordProof writes the immutable proof-of-retention record. One record
// per policy, updated at each run — the record proves the last execution
// and its outcome, never member data.
func (runner *Runner) recordProof(ctx context.Context, policy Policy, matched int) error {
	_, err := runner.database.Collection("retention_audit").UpdateOne(ctx,
		bson.M{"_id": policy.Name},
		bson.M{
			"$set": bson.M{"collection": policy.Collection, "action": string(policy.Action), "matched": matched, "processedAt": runner.clock().UTC()},
			"$inc": bson.M{"runs": 1},
		},
		options.UpdateOne().SetUpsert(true))
	return err
}
