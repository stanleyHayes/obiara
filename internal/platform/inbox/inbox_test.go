package inbox

import (
	"context"
	"testing"
	"time"
)

func TestAlreadyProcessedValidation(t *testing.T) {
	// Validation must fail before any database access, so a nil store
	// database is intentional here.
	store := NewStore(nil, time.Now)

	if _, err := store.AlreadyProcessed(context.Background(), "", "msg-1"); err != ErrConsumerRequired {
		t.Fatalf("empty consumer error = %v, want %v", err, ErrConsumerRequired)
	}
	if _, err := store.AlreadyProcessed(context.Background(), "relay", ""); err != ErrMessageIDRequired {
		t.Fatalf("empty message id error = %v, want %v", err, ErrMessageIDRequired)
	}
}
