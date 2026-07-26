package invariants_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	garden "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
	sprout "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
	water "github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
)

func TestExportedAuditAndProjectionSurfacesContainOnlyOpaqueReferences(t *testing.T) {
	now := time.Now().UTC()
	a, b := opaque(1), opaque(2)
	doorway, _ := sprout.Open("doorway", a, b, now)
	doorway, _, _ = doorway.Exchange(a, opaque(3), "exchange", "fingerprint", now)
	watering, _ := water.Start("water", []string{a, b}, waterCommand("first", a, 0, now))
	item, _ := garden.New(opaque(4), opaque(5), now.Add(time.Hour), now)
	payload, err := json.Marshal(struct {
		Doorway []sprout.Exchange
		Water   []water.Event
		Garden  garden.Item
	}{doorway.Exchanges(), watering.Events(), item})
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(payload))
	for _, forbidden := range []string{"raw-user", "email", "phone", "readreceipt", "read_receipt", "publicsignal", "public_signal", "payment", "boost", "price"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("privacy or pressure field %q leaked in %s", forbidden, text)
		}
	}
}
