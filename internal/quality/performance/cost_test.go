package performance

import (
	"encoding/json"
	"go.yaml.in/yaml/v3"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type costFile struct {
	Version          int
	Currency, Status string
	Assumptions      struct {
		FixedComputeCents         int64 `yaml:"fixed_compute_cents"`
		FixedDatabaseCents        int64 `yaml:"fixed_database_cents"`
		StorageGiBCents           int64 `yaml:"storage_gib_cents"`
		EgressGiBCents            int64 `yaml:"egress_gib_cents"`
		NotificationThousandCents int64 `yaml:"notification_thousand_cents"`
		AIMinuteThousandCents     int64 `yaml:"ai_minute_thousand_cents"`
		LiveSeatHourThousandCents int64 `yaml:"live_seat_hour_thousand_cents"`
	} `yaml:"assumptions"`
	Scenarios []struct {
		Name          string
		MAU           int64 `yaml:"mau"`
		StorageGiB    int64 `yaml:"storage_gib"`
		EgressGiB     int64 `yaml:"egress_gib"`
		Notifications int64 `yaml:"notifications"`
		AIMinutes     int64 `yaml:"ai_minutes"`
		LiveSeatHours int64 `yaml:"live_seat_hours"`
	} `yaml:"scenarios"`
}

func loadCosts(t *testing.T) costFile {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	raw, e := os.ReadFile(filepath.Join(filepath.Dir(file), "../../../deploy/performance/cost-assumptions.yaml"))
	if e != nil {
		t.Fatal(e)
	}
	var c costFile
	if e = yaml.Unmarshal(raw, &c); e != nil {
		t.Fatal(e)
	}
	return c
}
func TestCommittedCostEvidenceIsCompleteMonotonicAndMachineReadable(t *testing.T) {
	c := loadCosts(t)
	if c.Version != 1 || c.Currency != "USD_cents" || c.Status != "planning_assumption_not_provider_quote" || len(c.Scenarios) != 3 {
		t.Fatalf("invalid cost contract %#v", c)
	}
	a := CostAssumptions(c.Assumptions)
	var previous int64
	for _, raw := range c.Scenarios {
		s := CostScenario(raw)
		evidence, e := Estimate(a, s)
		if e != nil {
			t.Fatal(e)
		}
		if s.Name == "architecture_100k_mau" && (evidence.MonthlyCents < 1_200_000 || evidence.MonthlyCents > 2_200_000) {
			t.Fatalf("100k MAU evidence %d cents is outside Doc 07 order-of-magnitude envelope", evidence.MonthlyCents)
		}
		if evidence.MonthlyCents <= previous {
			t.Fatalf("cost not monotonic: %#v", evidence)
		}
		previous = evidence.MonthlyCents
		encoded, e := json.Marshal(evidence)
		if e != nil || !json.Valid(encoded) {
			t.Fatal("invalid evidence JSON")
		}
	}
}
func TestCostFormulaRoundsMeteredUnitsUp(t *testing.T) {
	e, err := Estimate(CostAssumptions{NotificationThousandCents: 100}, CostScenario{Name: "one", MAU: 1, Notifications: 1})
	if err != nil || e.NotificationCents != 100 || e.MonthlyCents != 100 {
		t.Fatalf("%#v %v", e, err)
	}
}
func BenchmarkCostEstimate(b *testing.B) {
	a := CostAssumptions{FixedComputeCents: 1}
	s := CostScenario{Name: "bench", MAU: 1}
	for i := 0; i < b.N; i++ {
		_, _ = Estimate(a, s)
	}
}
