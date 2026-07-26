package outbox

import (
	"context"
	"testing"
	"time"
)

func TestAppendValidation(t *testing.T) {
	// Validation must fail before any database access, so a nil store
	// database is intentional here.
	store := NewStore(nil, time.Now)
	base := Record{
		ID:            "evt-1",
		AggregateType: "member",
		AggregateID:   "member-1",
		EventType:     "member.registered",
		Payload:       []byte(`{"ok":true}`),
		OccurredAt:    time.Now(),
	}

	cases := map[string]struct {
		mutate  func(*Record)
		wantErr error
	}{
		"missing id":             {func(r *Record) { r.ID = "" }, ErrIDRequired},
		"missing aggregate type": {func(r *Record) { r.AggregateType = "" }, ErrAggregateTypeRequired},
		"missing aggregate id":   {func(r *Record) { r.AggregateID = "" }, ErrAggregateIDRequired},
		"missing event type":     {func(r *Record) { r.EventType = "" }, ErrEventTypeRequired},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			record := base
			tc.mutate(&record)
			if err := store.Append(context.Background(), record); err != tc.wantErr {
				t.Fatalf("Append error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
