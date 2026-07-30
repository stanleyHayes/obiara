// Package suban is the composition root of the Suban bounded context
// slice for the character ledger (E15-S04). Marks are recomputed per read;
// fairness dashboards and anti-gaming collusion checks land with E15-S05.
package suban

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/suban/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/application"
	explanationmongo "github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/adapters/outbound/mongodb"
	explanationapp "github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/application"
)

type Module struct {
	Suban       application.SubanService
	Explanation explanationapp.Service
}

type explanationAuthority struct{}

func (explanationAuthority) RequireSelf(_ context.Context, actor, subject string) error {
	if actor == "" || actor != subject {
		return errors.New("self access required")
	}
	return nil
}

func (explanationAuthority) RequireAppealReviewer(context.Context, string) error {
	return errors.New("admin reviewer authority is required")
}

type explanationIDs struct{}

func (explanationIDs) NewID() string {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}

type explanationClock struct{}

func (explanationClock) Now() time.Time { return time.Now() }

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	appeals := explanationmongo.New(database)
	if err := appeals.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Suban: application.NewSubanService(store, time.Now, newID),
		Explanation: explanationapp.New(
			explanationAuthority{}, store, appeals, explanationIDs{}, explanationClock{},
		),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "sub_" + base64.RawURLEncoding.EncodeToString(id)
}
