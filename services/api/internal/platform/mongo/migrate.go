package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Migration is one ordered, idempotent schema/data change. IDs must be
// unique, non-empty and added in append-only order; applied migrations are
// never edited (agent_plan.md §12: expand, migrate/backfill, contract).
type Migration struct {
	ID    string
	Apply func(context.Context, *mongo.Database) error
}

// Runner applies migrations once each, recording them in the
// schema_migrations collection. A unique _id per migration makes concurrent
// starters safe: the loser sees a duplicate key and treats the migration as
// applied.
type Runner struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewRunner(database *mongo.Database, clock func() time.Time) *Runner {
	return &Runner{database: database, clock: clock}
}

var (
	ErrMigrationIDRequired = errors.New("migration id is required")
	ErrMigrationDuplicate  = errors.New("migration ids must be unique")
	ErrMigrationApplyNil   = errors.New("migration apply function is required")
)

type migrationDocument struct {
	ID        string    `bson:"_id"`
	AppliedAt time.Time `bson:"appliedAt"`
}

// Run validates and applies migrations in order. Migrations recorded in a
// previous run are skipped.
func (runner *Runner) Run(ctx context.Context, migrations []Migration) error {
	if err := validate(migrations); err != nil {
		return err
	}
	records := runner.database.Collection("schema_migrations")
	for _, migration := range migrations {
		var existing migrationDocument
		err := records.FindOne(ctx, bson.M{"_id": migration.ID}).Decode(&existing)
		if err == nil {
			continue
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("check migration %q: %w", migration.ID, err)
		}
		if err := migration.Apply(ctx, runner.database); err != nil {
			return fmt.Errorf("apply migration %q: %w", migration.ID, err)
		}
		_, err = records.InsertOne(ctx, migrationDocument{ID: migration.ID, AppliedAt: runner.clock().UTC()})
		if err != nil && !IsDuplicateKey(err) {
			return fmt.Errorf("record migration %q: %w", migration.ID, err)
		}
	}
	return nil
}

func validate(migrations []Migration) error {
	seen := make(map[string]struct{}, len(migrations))
	for _, migration := range migrations {
		if migration.ID == "" {
			return ErrMigrationIDRequired
		}
		if migration.Apply == nil {
			return fmt.Errorf("%q: %w", migration.ID, ErrMigrationApplyNil)
		}
		if _, ok := seen[migration.ID]; ok {
			return fmt.Errorf("%q: %w", migration.ID, ErrMigrationDuplicate)
		}
		seen[migration.ID] = struct{}{}
	}
	return nil
}
