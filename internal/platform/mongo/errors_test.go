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
