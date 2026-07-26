// Package application holds the worker scheduling kernel (E02-S09):
// interval jobs with per-execution timeouts, exponential-backoff retries,
// dead letters and structured observability. Domain stores (outbox,
// inbox) live in the shared internal/platform packages; adapters implement
// the ports declared here.
package application

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job is one scheduled unit of durable work.
type Job struct {
	Name     string
	Interval time.Duration
	// Timeout bounds a single execution.
	Timeout time.Duration
	// MaxAttempts is the number of consecutive failures tolerated before a
	// dead letter is recorded and the counter resets (the job keeps its
	// schedule; operators triage dead letters).
	MaxAttempts int
	Run         func(ctx context.Context) error
}

// DeadLetter captures a job that exhausted its consecutive-failure budget.
type DeadLetter struct {
	JobName    string
	Reason     string
	Failures   int
	OccurredAt time.Time
}

// DeadLetterStore is the outbound port for dead-letter persistence.
type DeadLetterStore interface {
	Record(context.Context, DeadLetter) error
}

type Scheduler struct {
	jobs        []Job
	store       DeadLetterStore
	logger      *slog.Logger
	clock       func() time.Time
	baseBackoff time.Duration
}

func NewScheduler(jobs []Job, store DeadLetterStore, logger *slog.Logger, clock func() time.Time) *Scheduler {
	return &Scheduler{
		jobs:        jobs,
		store:       store,
		logger:      logger,
		clock:       clock,
		baseBackoff: 500 * time.Millisecond,
	}
}

// Start runs every job until ctx is cancelled, then waits for in-flight
// executions to finish.
func (scheduler *Scheduler) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, job := range scheduler.jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scheduler.loop(ctx, job)
		}()
	}
	wg.Wait()
	return nil
}

func (scheduler *Scheduler) loop(ctx context.Context, job Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	failures := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if err := scheduler.runOnce(ctx, job); err != nil {
			failures++
			backoff := scheduler.backoff(failures)
			scheduler.logger.WarnContext(ctx, "job execution failed",
				slog.String("job", job.Name),
				slog.Int("consecutiveFailures", failures),
				slog.Duration("backoff", backoff),
				slog.String("error", err.Error()))
			if job.MaxAttempts > 0 && failures >= job.MaxAttempts {
				scheduler.recordDeadLetter(ctx, job, err, failures)
				failures = 0
			}
			if !sleep(ctx, backoff) {
				return
			}
			continue
		}
		failures = 0
	}
}

// runOnce executes the job with its per-execution timeout.
func (scheduler *Scheduler) runOnce(ctx context.Context, job Job) error {
	execCtx, cancel := context.WithTimeout(ctx, job.Timeout)
	defer cancel()
	return job.Run(execCtx)
}

// backoff grows exponentially from baseBackoff, capped at one minute.
func (scheduler *Scheduler) backoff(failures int) time.Duration {
	backoff := scheduler.baseBackoff
	for i := 1; i < failures; i++ {
		backoff *= 2
		if backoff >= time.Minute {
			return time.Minute
		}
	}
	return backoff
}

func (scheduler *Scheduler) recordDeadLetter(ctx context.Context, job Job, cause error, failures int) {
	letter := DeadLetter{
		JobName:    job.Name,
		Reason:     cause.Error(),
		Failures:   failures,
		OccurredAt: scheduler.clock().UTC(),
	}
	if err := scheduler.store.Record(ctx, letter); err != nil {
		scheduler.logger.ErrorContext(ctx, "record dead letter failed",
			slog.String("job", job.Name),
			slog.String("error", err.Error()))
		return
	}
	scheduler.logger.ErrorContext(ctx, "job dead-lettered",
		slog.String("job", job.Name),
		slog.Int("failures", failures))
}

func sleep(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
