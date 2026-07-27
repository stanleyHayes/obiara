// Package performance provides bounded local load and cost evidence. It is not
// a production capacity claim or a substitute for device/network field tests.
package performance

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var ErrInvalidProfile = errors.New("invalid bounded load profile")

const (
	MaxRequests    = 10_000
	MaxConcurrency = 64
)

type Profile struct {
	Name         string        `json:"name"`
	Requests     int           `json:"requests"`
	Concurrency  int           `json:"concurrency"`
	MaxP95       time.Duration `json:"maxP95"`
	MaxErrorRate float64       `json:"maxErrorRate"`
}
type Result struct {
	Profile                     string  `json:"profile"`
	Samples                     int     `json:"samples"`
	Errors                      int     `json:"errors"`
	ErrorRate                   float64 `json:"errorRate"`
	P50, P90, P95, P99, Elapsed time.Duration
}

func (r Result) Within(p Profile) bool {
	return r.Samples == p.Requests && r.ErrorRate <= p.MaxErrorRate && r.P95 <= p.MaxP95
}
func (r Result) JSON() ([]byte, error) { return json.Marshal(r) }

func Run(ctx context.Context, p Profile, operation func(context.Context, int) error) (Result, error) {
	if p.Name == "" || p.Requests < 1 || p.Requests > MaxRequests || p.Concurrency < 1 || p.Concurrency > MaxConcurrency || p.Concurrency > p.Requests || p.MaxP95 <= 0 || p.MaxErrorRate < 0 || p.MaxErrorRate > 1 || operation == nil {
		return Result{}, ErrInvalidProfile
	}
	start := time.Now()
	jobs := make(chan int)
	samples := make(chan time.Duration, p.Requests)
	var failures atomic.Int64
	var wg sync.WaitGroup
	for range p.Concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				began := time.Now()
				if operation(ctx, index) != nil {
					failures.Add(1)
				}
				samples <- time.Since(began)
			}
		}()
	}
	go func() {
		defer close(jobs)
		for i := 0; i < p.Requests; i++ {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	wg.Wait()
	close(samples)
	durations := make([]time.Duration, 0, p.Requests)
	for duration := range samples {
		durations = append(durations, duration)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	result := Result{Profile: p.Name, Samples: len(durations), Errors: int(failures.Load()), Elapsed: time.Since(start)}
	if len(durations) > 0 {
		result.P50 = percentile(durations, 50)
		result.P90 = percentile(durations, 90)
		result.P95 = percentile(durations, 95)
		result.P99 = percentile(durations, 99)
		result.ErrorRate = float64(result.Errors) / float64(result.Samples)
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	return result, nil
}
func percentile(sorted []time.Duration, p int) time.Duration {
	index := (len(sorted)*p+99)/100 - 1
	if index < 0 {
		index = 0
	}
	return sorted[index]
}
