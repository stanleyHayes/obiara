package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	sowmedia "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/adapters/outbound/media"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

// SowRecordings opens a recording for a sow and returns the grant to upload it.
type SowRecordings interface {
	Open(
		ctx context.Context,
		ownerID, contentType string,
		sizeBytes int64,
		checksum string,
		duration time.Duration,
	) (sowmedia.Recording, error)
}

// RegisterSowRecordingRoutes exposes making the recording a sow carries.
//
// Until this existed there was nowhere to put one. The only upload path in the
// product created a Voice of Introduction, which is a different thing: it
// answers one of three fixed questions and is offered to anyone who may hear
// the member. So sending a sow required a recording that could not be made
// (agent_plan.md §68).
//
// Behind the sowing rung, because a member who cannot sow has nothing to
// record for.
func RegisterSowRecordingRoutes(
	mux *http.ServeMux,
	recordings SowRecordings,
	sessions SessionAuthenticator,
	gate MemberGate,
) {
	mux.Handle("POST /v1/seed/sows/recordings", gate.guard(
		sessions, "seeds.sow", "seed", openSowRecordingHandler(recordings, sessions),
	))
}

type sowRecordingRequest struct {
	ContentType string `json:"contentType"`
	// SizeBytes, Checksum and DurationMs describe the recording before its
	// bytes are sent, because the grant is signed over the length and the
	// digest — the store then refuses anything else.
	SizeBytes  int64  `json:"sizeBytes"`
	Checksum   string `json:"checksum"`
	DurationMs int64  `json:"durationMs"`
}

type sowRecordingResponse struct {
	// MediaRef is what the sow carries. It is the asset id under another
	// name, because that is what POST /v1/seed/sows calls it.
	MediaRef      string `json:"mediaRef"`
	UploadURL     string `json:"uploadUrl"`
	UploadExpires string `json:"uploadExpires"`
}

func openSowRecordingHandler(recordings SowRecordings, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		ownerID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if recordings == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		var body sowRecordingRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		recording, err := recordings.Open(
			r.Context(), ownerID, strings.TrimSpace(body.ContentType), body.SizeBytes,
			strings.ToLower(strings.TrimSpace(body.Checksum)),
			time.Duration(body.DurationMs)*time.Millisecond,
		)
		switch {
		case errors.Is(err, sowmedia.ErrNotRecordable):
			// Nothing here decodes audio, so the length is the client's word
			// and is bounded against the byte count rather than believed.
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{
					{Field: "durationMs", Reason: "does not match the size of the recording"},
				},
			})
			return
		case errors.Is(err, sowapplication.ErrUnavailable):
			logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		case err != nil:
			logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
			writeError(w, r, http.StatusInternalServerError, APIError{
				Code: "internal_error", Message: "The request could not be completed.",
			})
			return
		}
		writeSuccess(w, r, http.StatusCreated, sowRecordingResponse{
			MediaRef:      recording.AssetID,
			UploadURL:     recording.URL,
			UploadExpires: recording.ExpiresAt.UTC().Format(time.RFC3339),
		})
	})
}
