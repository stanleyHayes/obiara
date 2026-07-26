package application

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestPrivacyJobDelegatesBoundedBatchAndPropagatesFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	processor := NewMockPrivacyProcessor(ctrl)
	expected := errors.New("archive unavailable")
	processor.EXPECT().RunBatch(gomock.Any(), PrivacyBatchSize).Return(expected)

	job := NewPrivacyJob(processor)
	if job.Name != PrivacyJobName || job.Interval != PrivacyJobInterval || job.Timeout != PrivacyJobTimeout || job.MaxAttempts != 5 {
		t.Fatalf("job = %#v", job)
	}
	if err := job.Run(context.Background()); !errors.Is(err, expected) {
		t.Fatalf("error = %v, want %v", err, expected)
	}
}
