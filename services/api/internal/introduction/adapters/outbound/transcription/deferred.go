// Package transcription provides the transcriber used when no speech vendor
// is contracted.
//
// This mirrors the decision already taken for liveness in
// internal/verification/liveness/adapters/outbound/manual: rather than let a
// missing vendor block the feature, or invent a transcript nobody said, every
// request is reported as uncertain. The introduction aggregate already routes
// that to its own uncertain state, which keeps the recording playable and
// leaves the transcript absent until a real provider or a human supplies one.
//
// A transcript is not what the product is for — a member is introduced by
// their voice. The transcript exists for search, accessibility and safety
// review, and none of those are worth blocking a recording on.
package transcription

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
)

// Deferred answers every request as uncertain and never errors: a deliberate
// policy must not be recorded as a provider outage, because an outage is
// retried and a policy is not.
type Deferred struct{}

func NewDeferred() Deferred { return Deferred{} }

func (Deferred) Transcribe(
	_ context.Context,
	_ application.TranscriptionRequest,
) (application.TranscriptionResult, error) {
	return application.TranscriptionResult{Outcome: application.TranscriptionUncertain}, nil
}

// Cancel and Delete are no-ops with nothing to undo: no work was ever handed
// to a vendor, so there is no job to stop and no remote copy to erase. They
// return nil rather than an error so revoke and purge still complete.
func (Deferred) Cancel(context.Context, string) error { return nil }
func (Deferred) Delete(context.Context, string) error { return nil }
