package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/internal/privacy/domain"
)

// ErasureRunner executes deletion (FR-106: within 30 days, cryptographic
// erasure of voice/biometric blobs). Media bytes live in object storage
// behind the media kernel; object deletion wires in when media persistence
// adapters land. Until then the runner erases every metadata record and
// tombstones the account. Legal-hold checks happen before requests open
// (E03-S10), never here.
type ErasureRunner struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewErasureRunner(database *mongo.Database, clock func() time.Time) *ErasureRunner {
	return &ErasureRunner{database: database, clock: clock}
}

// erasureTarget maps one collection to its owner field for deletion.
type erasureTarget struct {
	collection string
	field      string
}

var erasureTargets = []erasureTarget{
	{collection: "members", field: "_id"},
	{collection: "profiles", field: "_id"},
	{collection: "doorway_questions", field: "_id"},
	{collection: "photo_vault", field: "memberId"},
	{collection: "media_assets", field: "ownerId"},
	{collection: "playback_records", field: "listenerId"},
	{collection: "otp_challenges", field: "phone"}, // applied via account phone lookup
	{collection: "voice_assets", field: "accountId"},
	{collection: "photo_assets", field: "accountId"},
	{collection: "sessions", field: "memberId"},
	// privacy_export_archives are delivery artifacts for the member's own
	// export; they expire on their own lifecycle (TTL), not on deletion.
}

// Erase deletes the account's personal records, tombstones the account
// (status=deleted, retaining the phone for the FR-102 one-account rule)
// and writes a proof-of-deletion audit record that never contains the
// subject identifier (agent_plan.md §15: proof-of-deletion records).
// Erasure is replay-safe. An active legal hold stops erasure even at
// execution time — defense in depth behind the request-time hold check
// (E03-S10, Doc 09).
func (runner *ErasureRunner) Erase(ctx context.Context, requestID, accountID string) error {
	var hold struct {
		LiftedAt *time.Time `bson:"liftedAt"`
	}
	holdErr := runner.database.Collection("legal_holds").FindOne(ctx, bson.M{"_id": accountID}).Decode(&hold)
	if holdErr == nil && hold.LiftedAt == nil {
		return domain.ErrLegalHoldActive
	}
	if holdErr != nil && holdErr != mongo.ErrNoDocuments {
		return holdErr
	}

	erased := make(map[string]int64, len(erasureTargets)-1)
	for _, target := range erasureTargets {
		if target.collection == "otp_challenges" {
			continue // keyed by phone; handled below after account lookup
		}
		result, err := runner.database.Collection(target.collection).DeleteMany(ctx, bson.M{target.field: accountID})
		if err != nil {
			return fmt.Errorf("erase %s: %w", target.collection, err)
		}
		erased[target.collection] = result.DeletedCount
	}

	accounts := runner.database.Collection("accounts")
	var account struct {
		Phone string `bson:"phone"`
	}
	if err := accounts.FindOne(ctx, bson.M{"_id": accountID}).Decode(&account); err == nil && account.Phone != "" {
		if _, err := runner.database.Collection("otp_challenges").DeleteMany(ctx, bson.M{"phone": account.Phone}); err != nil {
			return fmt.Errorf("erase otp_challenges: %w", err)
		}
	}

	tombstone := runner.clock().UTC()
	// Tombstone when an account record exists. The processor only executes
	// requests opened for real accounts, so a missing record means the
	// account was already erased or belongs to a member context the runner
	// does not tombstone — not a reason to abort mid-erasure.
	if _, err := accounts.UpdateOne(ctx,
		bson.M{"_id": accountID},
		bson.M{"$set": bson.M{"status": "deleted", "deletedAt": tombstone}}); err != nil {
		return fmt.Errorf("tombstone account: %w", err)
	}

	// Proof-of-deletion audit: keyed by request, free of subject data, and
	// upserted so replays stay safe (agent_plan.md §15).
	if _, err := runner.database.Collection("privacy_erasure_audit").UpdateOne(ctx,
		bson.M{"_id": requestID},
		bson.M{"$set": bson.M{"erasedCollections": erased, "erasedAt": tombstone}},
		options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("record erasure audit: %w", err)
	}
	return nil
}
