package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

var (
	ErrNotFound              = errors.New("voice introduction not found")
	ErrOptimisticConflict    = errors.New("voice introduction changed")
	ErrCommandAlreadyUsed    = errors.New("voice introduction command already used")
	ErrDependencyUnavailable = errors.New("voice introduction dependency unavailable")
	ErrConsentRequired       = errors.New("effective voice introduction consent required")
)

type Store interface {
	// Create is idempotent by the aggregate's creation command.
	Create(context.Context, domain.Introduction) (domain.Introduction, bool, error)
	FindByID(context.Context, string) (domain.Introduction, error)
	Update(context.Context, domain.Introduction, uint64, string) error
}

type ConsentGate interface {
	Effective(context.Context, string, string, uint64) (bool, error)
}

type UploadAccess struct {
	URL       string
	ExpiresAt time.Time
}

type MediaManager interface {
	AuthorizeUpload(context.Context, string, string, string) (UploadAccess, error)
	Inspect(context.Context, string) (domain.MediaRef, error)
	Delete(context.Context, string) error
}

type TranscriptionOutcome string

const (
	TranscriptionCompleted TranscriptionOutcome = "completed"
	TranscriptionUncertain TranscriptionOutcome = "uncertain"
	TranscriptionFailed    TranscriptionOutcome = "failed"
)

type TranscriptionRequest struct {
	CommandID      string
	IntroductionID string
	AssetID        string
	ConsentPurpose string
	ConsentVersion uint64
}

type TranscriptionResult struct {
	Outcome    TranscriptionOutcome
	Transcript domain.TranscriptRef
}

type Transcriber interface {
	Transcribe(context.Context, TranscriptionRequest) (TranscriptionResult, error)
	Cancel(context.Context, string) error
	Delete(context.Context, string) error
}

type Keyer interface {
	Key(string) (string, error)
}

type IDSource interface {
	NewID(string) string
}
