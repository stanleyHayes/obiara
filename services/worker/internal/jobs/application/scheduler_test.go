package application

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestBackoffProgressionAndCap(t *testing.T) {
	scheduler := NewScheduler(nil, nil, quietLogger(), time.Now)
	scheduler.baseBackoff = 100 * time.Millisecond

	want := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}
	for i, expected := range want {
		if got := scheduler.backoff(i + 1); got != expected {
			t.Fatalf("backoff(%d) = %v, want %v", i+1, got, expected)
		}
	}
	if got := scheduler.backoff(20); got != time.Minute {
		t.Fatalf("backoff(20) = %v, want cap 1m", got)
	}
}

func TestRunOnceHonoursTimeout(t *testing.T) {
	scheduler := NewScheduler(nil, nil, quietLogger(), time.Now)
	job := Job{
		Name:    "slow",
		Timeout: 50 * time.Millisecond,
		Run: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	start := time.Now()
	if err := scheduler.runOnce(context.Background(), job); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("runOnce error = %v, want deadline exceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("runOnce took %v, timeout not enforced", elapsed)
	}
}

func TestDeadLetterAfterMaxAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockDeadLetterStore(ctrl)
	store.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, letter DeadLetter) error {
			if letter.JobName != "flaky" || letter.Failures != 2 {
				t.Fatalf("letter = %#v", letter)
			}
			return nil
		}).MinTimes(1)

	scheduler := NewScheduler(nil, store, quietLogger(), func() time.Time { return testNow })
	scheduler.baseBackoff = time.Millisecond

	var executions atomic.Int32
	job := Job{
		Name:        "flaky",
		Interval:    time.Millisecond,
		Timeout:     time.Second,
		MaxAttempts: 2,
		Run: func(context.Context) error {
			executions.Add(1)
			return errors.New("boom")
		},
	}
	scheduler.jobs = []Job{job}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := scheduler.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if executions.Load() < 2 {
		t.Fatalf("executions = %d, want at least 2", executions.Load())
	}
}

func TestHealthyJobNeverDeadLetters(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockDeadLetterStore(ctrl)
	// No Record expectation: any call fails the test.

	scheduler := NewScheduler(nil, store, quietLogger(), time.Now)
	var executions atomic.Int32
	scheduler.jobs = []Job{{
		Name:     "healthy",
		Interval: time.Millisecond,
		Timeout:  time.Second,
		Run: func(context.Context) error {
			executions.Add(1)
			return nil
		},
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := scheduler.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if executions.Load() == 0 {
		t.Fatal("healthy job never executed")
	}
}
