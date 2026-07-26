package telemetry

import (
	"context"
	"sync"
	"time"
)

// Attribute is a low-cardinality metric dimension. Never put member or
// request identifiers in metric attributes.
type Attribute struct {
	Key   string
	Value string
}

// Metrics is the adapter boundary used by request, dependency, and worker
// instrumentation. Production can bind it to OpenTelemetry without coupling
// application code to an exporter.
type Metrics interface {
	Add(context.Context, string, int64, ...Attribute)
	Observe(context.Context, string, time.Duration, ...Attribute)
}

// NoopMetrics is appropriate when metrics export is intentionally disabled.
type NoopMetrics struct{}

func (NoopMetrics) Add(context.Context, string, int64, ...Attribute)             {}
func (NoopMetrics) Observe(context.Context, string, time.Duration, ...Attribute) {}

// MemoryMetrics is a race-safe deterministic adapter for tests and local
// diagnostics. It intentionally ignores dimensions to avoid accidental
// high-cardinality aggregation.
type MemoryMetrics struct {
	mu           sync.RWMutex
	counters     map[string]int64
	observations map[string][]time.Duration
}

func NewMemoryMetrics() *MemoryMetrics {
	return &MemoryMetrics{
		counters:     make(map[string]int64),
		observations: make(map[string][]time.Duration),
	}
}

func (metrics *MemoryMetrics) Add(
	_ context.Context,
	name string,
	delta int64,
	_ ...Attribute,
) {
	if metrics == nil || !validMetricName(name) {
		return
	}
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	metrics.counters[name] += delta
}

func (metrics *MemoryMetrics) Observe(
	_ context.Context,
	name string,
	value time.Duration,
	_ ...Attribute,
) {
	if metrics == nil || !validMetricName(name) || value < 0 {
		return
	}
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	metrics.observations[name] = append(metrics.observations[name], value)
}

// Snapshot returns independent copies safe for concurrent callers.
func (metrics *MemoryMetrics) Snapshot() (map[string]int64, map[string][]time.Duration) {
	if metrics == nil {
		return map[string]int64{}, map[string][]time.Duration{}
	}
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	counters := make(map[string]int64, len(metrics.counters))
	for name, value := range metrics.counters {
		counters[name] = value
	}
	observations := make(map[string][]time.Duration, len(metrics.observations))
	for name, values := range metrics.observations {
		observations[name] = append([]time.Duration(nil), values...)
	}
	return counters, observations
}

func validMetricName(name string) bool {
	if name == "" || len(name) > 96 {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '.', r == '_':
		default:
			return false
		}
	}
	return true
}
