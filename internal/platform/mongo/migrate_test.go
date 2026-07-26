package mongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestRunnerValidation(t *testing.T) {
	apply := func(context.Context, *mongo.Database) error { return nil }

	cases := map[string]struct {
		migrations []Migration
		wantErr    error
	}{
		"empty id": {
			[]Migration{{ID: "", Apply: apply}},
			ErrMigrationIDRequired,
		},
		"duplicate ids": {
			[]Migration{{ID: "0001_a", Apply: apply}, {ID: "0001_a", Apply: apply}},
			ErrMigrationDuplicate,
		},
		"nil apply": {
			[]Migration{{ID: "0001_a"}},
			ErrMigrationApplyNil,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Validation must fail before any database access, so a nil
			// runner database is intentional here.
			runner := NewRunner(nil, time.Now)
			if err := runner.Run(context.Background(), tc.migrations); err == nil {
				t.Fatalf("Run(%v) succeeded, want %v", tc.migrations, tc.wantErr)
			} else if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Run(%v) error = %v, want %v", tc.migrations, err, tc.wantErr)
			}
		})
	}
}

func TestConnectUnreachableURI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := Connect(ctx, "mongodb://localhost:29999")
	if err == nil {
		if client != nil {
			_ = client.Disconnect(context.Background())
		}
		t.Fatal("Connect to unreachable URI succeeded, want error")
	}
}
