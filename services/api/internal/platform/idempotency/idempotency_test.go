package idempotency

import (
	"context"
	"testing"
	"time"
)

func TestClaimValidation(t *testing.T) {
	// Validation must fail before any database access, so a nil store
	// database is intentional here.
	store := NewStore(nil, time.Now)

	if _, err := store.Claim(context.Background(), "", "key-1"); err != ErrScopeRequired {
		t.Fatalf("empty scope error = %v, want %v", err, ErrScopeRequired)
	}
	if _, err := store.Claim(context.Background(), "member.register", ""); err != ErrKeyRequired {
		t.Fatalf("empty key error = %v, want %v", err, ErrKeyRequired)
	}
}
