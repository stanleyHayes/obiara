// Package scamarc is the composition root of the scam-arc detection slice
// (E11-S11). Consent defaults to monitoring-on per Doc 08 §8 until the
// consent registry ships.
package scamarc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/application"
)

type Module struct {
	ScamArc application.ScamArcService
}

// ConsentAllowAll is the default monitoring consent (Doc 08 §8: scam-arc
// room monitoring defaults to on with opt-out). The registry bridge
// replaces this when per-room consent records exist.
type ConsentAllowAll struct{}

func (ConsentAllowAll) MonitoringAllowed(context.Context, string) (bool, error) { return true, nil }

func NewModule(ctx context.Context, database *mongo.Database, consent application.MonitoringConsent, cases application.CaseOpener) (Module, error) {
	store := mongodb.NewStore(database)
	if err := store.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	if consent == nil {
		consent = ConsentAllowAll{}
	}
	return Module{
		ScamArc: application.NewScamArcService(store, consent, cases, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "sig_" + base64.RawURLEncoding.EncodeToString(id)
}
