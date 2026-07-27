package performance

import (
	"context"
	"encoding/json"
	"errors"
	"go.yaml.in/yaml/v3"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"testing/quick"
	"time"
)

type budgetFile struct {
	Version     int
	Environment string
	Profiles    []struct {
		Name             string
		Requests         int
		Concurrency      int
		MaximumP95MS     int     `yaml:"maximum_p95_ms"`
		MaximumErrorRate float64 `yaml:"maximum_error_rate"`
	} `yaml:"profiles"`
	Guardrails struct {
		MaximumRequests        int  `yaml:"maximum_requests"`
		MaximumConcurrency     int  `yaml:"maximum_concurrency"`
		ExternalTargetsAllowed bool `yaml:"external_targets_allowed"`
		ProductionDataAllowed  bool `yaml:"production_data_allowed"`
		BenchmarkIsSLA         bool `yaml:"benchmark_is_sla"`
	} `yaml:"guardrails"`
}

func TestCommittedBudgetsAreBoundedAndExplicit(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	raw, e := os.ReadFile(filepath.Join(filepath.Dir(file), "../../../deploy/performance/budgets.yaml"))
	if e != nil {
		t.Fatal(e)
	}
	var budgets budgetFile
	if e = yaml.Unmarshal(raw, &budgets); e != nil {
		t.Fatal(e)
	}
	if budgets.Version != 1 || budgets.Environment != "local_synthetic" || len(budgets.Profiles) != 2 || budgets.Guardrails.MaximumRequests != MaxRequests || budgets.Guardrails.MaximumConcurrency != MaxConcurrency || budgets.Guardrails.ExternalTargetsAllowed || budgets.Guardrails.ProductionDataAllowed || budgets.Guardrails.BenchmarkIsSLA {
		t.Fatalf("unsafe budget contract %#v", budgets)
	}
	for _, p := range budgets.Profiles {
		if _, e := Run(context.Background(), Profile{Name: p.Name, Requests: p.Requests, Concurrency: p.Concurrency, MaxP95: time.Duration(p.MaximumP95MS) * time.Millisecond, MaxErrorRate: p.MaximumErrorRate}, func(context.Context, int) error { return nil }); e != nil {
			t.Fatalf("%s: %v", p.Name, e)
		}
	}
}
func TestRunBoundedAndMachineReadable(t *testing.T) {
	p := Profile{Name: "unit", Requests: 1000, Concurrency: 8, MaxP95: 50 * time.Millisecond, MaxErrorRate: 0}
	r, e := Run(context.Background(), p, func(context.Context, int) error { return nil })
	if e != nil {
		t.Fatal(e)
	}
	if !r.Within(p) {
		t.Fatalf("outside budget: %#v", r)
	}
	raw, e := r.JSON()
	if e != nil || !json.Valid(raw) {
		t.Fatal("invalid evidence JSON")
	}
}
func TestRunRejectsUnboundedProfiles(t *testing.T) {
	for _, p := range []Profile{{Name: "large", Requests: MaxRequests + 1, Concurrency: 1, MaxP95: time.Second}, {Name: "wide", Requests: 100, Concurrency: MaxConcurrency + 1, MaxP95: time.Second}, {Name: "empty", Requests: 1, Concurrency: 1, MaxP95: 0}} {
		if _, e := Run(context.Background(), p, func(context.Context, int) error { return nil }); !errors.Is(e, ErrInvalidProfile) {
			t.Fatalf("%#v: %v", p, e)
		}
	}
}
func TestPercentileProperty(t *testing.T) {
	if e := quick.Check(func(values []uint16) bool {
		if len(values) == 0 {
			return true
		}
		d := make([]time.Duration, len(values))
		for i, v := range values {
			d[i] = time.Duration(v)
		}
		slices.Sort(d)
		p := percentile(d, 95)
		return p >= d[0] && p <= d[len(d)-1]
	}, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
func BenchmarkRunnerOverhead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Run(context.Background(), Profile{Name: "benchmark", Requests: 64, Concurrency: 8, MaxP95: time.Second}, func(context.Context, int) error { return nil })
	}
}
