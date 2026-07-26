// Package suban is the composition root of the Suban bounded context
// slice for the character ledger (E15-S04). Marks are recomputed per read;
// fairness dashboards and anti-gaming collusion checks land with E15-S05.
package suban

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/suban/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/application"
)

type Module struct {
	Suban application.SubanService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{Suban: application.NewSubanService(store, time.Now, newID)}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "sub_" + base64.RawURLEncoding.EncodeToString(id)
}
