package application

import (
	"context"
	"time"
)

const (
	PrivacyJobName     = "privacy-requests"
	PrivacyBatchSize   = 25
	PrivacyJobInterval = 30 * time.Second
	PrivacyJobTimeout  = 20 * time.Second
)

// PrivacyProcessor is the worker-facing port. The privacy bounded context
// owns request policy and execution; the scheduler only supplies durability.
type PrivacyProcessor interface {
	RunBatch(context.Context, int) error
}

func NewPrivacyJob(processor PrivacyProcessor) Job {
	return Job{
		Name: PrivacyJobName, Interval: PrivacyJobInterval, Timeout: PrivacyJobTimeout,
		MaxAttempts: 5,
		Run: func(ctx context.Context) error {
			return processor.RunBatch(ctx, PrivacyBatchSize)
		},
	}
}
