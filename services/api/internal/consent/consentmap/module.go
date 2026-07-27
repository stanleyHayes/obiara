// Package consentmap is the composition root of the consent-map registry
// (Doc 08 §8). Feature PRs name their consent-map row; enforcement ports
// in other contexts consult StateFor through composition bridges.
package consentmap

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/application"
)

type Module struct {
	ConsentMap application.ConsentMapService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		ConsentMap: application.NewConsentMapService(store, store, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "cons_" + base64.RawURLEncoding.EncodeToString(id)
}
