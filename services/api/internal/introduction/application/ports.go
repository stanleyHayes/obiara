package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

var (
	// ErrImplausibleRecording refuses a claimed length that could not have
	// come from the bytes being uploaded. Nothing here can decode audio, so
	// the length is the client's word; the twenty seconds that arm a sow are
	// counted against it, so the word is bounded rather than taken.
	ErrImplausibleRecording = errors.New("that recording's length does not fit its size")
	// ErrUploadNotArrived reports a confirmation for bytes that are not in
	// storage. It is separate from "not found" because the introduction does
	// exist; what is missing is the recording, and only that is something a
	// member can put right by uploading again.
	ErrUploadNotArrived      = errors.New("that recording has not arrived in storage")
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
	// PromptsRecorded lists the distinct prompts this member has a usable
	// recording for. Distinct, because re-recording one question three times
	// is three aggregates and must not read as a finished introduction.
	PromptsRecorded(context.Context, string) ([]domain.Prompt, error)
}

// Ladder is told when a member's Voice of Introduction becomes complete.
//
// It is a port so this context never reaches into identity; the composition
// root bridges it (agent_plan.md §7.2). Implementations must be idempotent:
// a retried confirmation, or a fourth recording, calls it again.
type Ladder interface {
	SowingEarned(ctx context.Context, memberID string) error
}

type ConsentGate interface {
	Effective(context.Context, string, string, uint64) (bool, error)
}

type UploadAccess struct {
	URL       string
	ExpiresAt time.Time
}

type MediaManager interface {
	// AuthorizeUpload registers the recording and returns a grant to send its
	// bytes. It takes the MediaRef whole rather than three strings: the
	// previous signature was (subjectID, assetID, contentType) and was called
	// with (ownerID, introductionID, assetID), so it looked the asset up by
	// the introduction's id and threw the real one away as a content type.
	AuthorizeUpload(context.Context, string, domain.MediaRef) (UploadAccess, error)
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
