package providerdiligence

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)

func validRegistry() Registry {
	r := Registry{SchemaVersion: SchemaVersion, GeneratedAt: testNow.Add(-time.Hour), ExpiresAt: testNow.Add(time.Hour), Environment: "production", Market: "GH"}
	n := 1
	for _, id := range ProviderIDs() {
		p := Provider{ID: id, Outcome: "approved", IssuerRef: "role/procurement-owner", ReviewerRef: "role/privacy-reviewer"}
		for _, subject := range Subjects() {
			p.Evidence = append(p.Evidence, Evidence{Subject: subject, EvidenceRef: fmt.Sprintf("%064x", n), CollectedAt: testNow.Add(-time.Hour), ExpiresAt: testNow.Add(24 * time.Hour), Outcome: "accepted"})
			n++
		}
		r.Providers = append(r.Providers, p)
	}
	return r
}

func TestLoadIsStrictAndSyntheticFixtureBlocked(t *testing.T) {
	_, decision, err := Load("../../../deploy/release/examples/provider-diligence.synthetic.json", testNow)
	if err != nil || decision.Ready || len(decision.Blockers) != 4 {
		t.Fatalf("fixture must be valid but blocked: %#v %v", decision, err)
	}
	path := t.TempDir() + "/bad.json"
	raw := `{"schemaVersion":"obiara.provider-diligence.v1","generatedAt":"2026-07-27T11:00:00Z","expiresAt":"2026-07-27T13:00:00Z","environment":"production","market":"GH","providers":[],"privatePayload":"no"}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path, testNow); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown private field must fail: %v", err)
	}
	if err := os.WriteFile(path, append([]byte(raw[:len(raw)-1]), []byte("}{}")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path, testNow); err == nil {
		t.Fatal("trailing data must fail")
	}
}

func TestCompleteRegistryReady(t *testing.T) {
	d, err := Evaluate(validRegistry(), testNow)
	if err != nil || !d.Ready || len(d.Blockers) != 0 {
		t.Fatalf("expected ready: %#v %v", d, err)
	}
}

func TestFailsClosed(t *testing.T) {
	tests := map[string]func(*Registry){
		"missing provider":   func(r *Registry) { r.Providers = r.Providers[1:] },
		"missing subject":    func(r *Registry) { r.Providers[0].Evidence = r.Providers[0].Evidence[1:] },
		"synthetic provider": func(r *Registry) { r.Providers[0].Synthetic = true },
		"synthetic evidence": func(r *Registry) { r.Providers[0].Evidence[0].Synthetic = true },
		"pending provider":   func(r *Registry) { r.Providers[0].Outcome = "pending" },
		"pending evidence":   func(r *Registry) { r.Providers[0].Evidence[0].Outcome = "pending" },
		"stale evidence":     func(r *Registry) { r.Providers[0].Evidence[0].ExpiresAt = testNow.Add(-time.Second) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			r := validRegistry()
			mutate(&r)
			d, err := Evaluate(r, testNow)
			if err != nil || d.Ready {
				t.Fatalf("must return a blocked decision, got %#v %v", d, err)
			}
		})
	}
}

func TestRejectsMalformedRegistry(t *testing.T) {
	tests := map[string]func(*Registry){
		"unknown provider":   func(r *Registry) { r.Providers[0].ID = "other" },
		"duplicate provider": func(r *Registry) { r.Providers = append(r.Providers, r.Providers[0]) },
		"self review":        func(r *Registry) { r.Providers[0].ReviewerRef = r.Providers[0].IssuerRef },
		"duplicate subject": func(r *Registry) {
			r.Providers[0].Evidence = append(r.Providers[0].Evidence, r.Providers[0].Evidence[0])
		},
		"replayed ref": func(r *Registry) { r.Providers[1].Evidence[0].EvidenceRef = r.Providers[0].Evidence[0].EvidenceRef },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			r := validRegistry()
			mutate(&r)
			if _, err := Evaluate(r, testNow); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDecisionIsOrderInvariant(t *testing.T) {
	base, err := Evaluate(validRegistry(), testNow)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewSource(17))
	for i := 0; i < 1000; i++ {
		r := validRegistry()
		random.Shuffle(len(r.Providers), func(i, j int) { r.Providers[i], r.Providers[j] = r.Providers[j], r.Providers[i] })
		for p := range r.Providers {
			random.Shuffle(len(r.Providers[p].Evidence), func(i, j int) {
				r.Providers[p].Evidence[i], r.Providers[p].Evidence[j] = r.Providers[p].Evidence[j], r.Providers[p].Evidence[i]
			})
		}
		got, err := Evaluate(r, testNow)
		if err != nil || !reflect.DeepEqual(got, base) {
			t.Fatalf("round %d changed decision: %#v %v", i, got, err)
		}
	}
}
