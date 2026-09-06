package mongo

import (
	"errors"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestIsDuplicateKey(t *testing.T) {
	duplicateWrite := mongo.WriteException{
		WriteErrors: []mongo.WriteError{{Code: 11000, Message: "E11000 duplicate key error"}},
	}
	duplicateCommand := mongo.CommandError{Code: 11000, Message: "E11000 duplicate key error"}

	cases := map[string]struct {
		err  error
		want bool
	}{
		"write exception 11000":         {duplicateWrite, true},
		"wrapped write exception":       {fmt.Errorf("insert member: %w", duplicateWrite), true},
		"command error 11000":           {duplicateCommand, true},
		"write exception other code":    {mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 121}}}, false},
		"unrelated error":               {errors.New("network down"), false},
		"nil is not reached by callers": {nil, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsDuplicateKey(tc.err); got != tc.want {
				t.Fatalf("IsDuplicateKey(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestAFreshDatabaseCanStillRetireAnIndex(t *testing.T) {
	// The case this predicate always claimed to cover and did not: a fresh
	// database has no collection, so dropping a retired index answers
	// NamespaceNotFound rather than IndexNotFound. Failing on it meant the
	// product could not boot against an empty database at all — which is
	// every new environment.
	for _, err := range []error{
		mongo.CommandError{Code: 26, Name: "NamespaceNotFound"},
		mongo.CommandError{Code: 27, Name: "IndexNotFound"},
		// Some server versions answer with only one of the two populated.
		mongo.CommandError{Code: 26},
		mongo.CommandError{Name: "NamespaceNotFound"},
	} {
		if !IsIndexNotFound(err) {
			t.Fatalf("%#v was treated as a real failure", err)
		}
	}
}

func TestARealFailureIsStillAFailure(t *testing.T) {
	// The direction that matters: tolerating everything would hide a
	// permissions error or an unreachable server behind a silent boot.
	for _, err := range []error{
		mongo.CommandError{Code: 13, Name: "Unauthorized"},
		mongo.CommandError{Code: 11000, Name: "DuplicateKey"},
		errors.New("connection refused"),
	} {
		if IsIndexNotFound(err) {
			t.Fatalf("%#v was swallowed", err)
		}
	}
}
