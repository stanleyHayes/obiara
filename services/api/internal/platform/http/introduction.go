package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

// VoiceIntroduction is the inbound port for the Voice of Introduction
// (E04-S02). A member records; the bytes go straight to object storage under
// a signed grant and never through this service.
type VoiceIntroduction interface {
	BeginUpload(context.Context, application.BeginUploadRequest) (application.BeginUploadResult, error)
	ConfirmUpload(ctx context.Context, introductionID, commandID string) (domain.Introduction, error)
	Transcribe(ctx context.Context, introductionID, commandID string) (application.TranscribeResult, error)
	Revoke(ctx context.Context, introductionID, commandID string) (domain.Introduction, error)
}

// IntroductionReader reads one back for its owner.
type IntroductionReader interface {
	FindByID(context.Context, string) (domain.Introduction, error)
}

// IntroductionPlayback issues a short-lived grant to hear one recording.
// Whether this subject may hear this asset is the media context's decision,
// not this transport's.
type IntroductionPlayback interface {
	AuthorizePlayback(ctx context.Context, subjectID, assetID string) (application.UploadAccess, error)
}

// voiceIntroductionPurpose must match introduction.ConsentPurposeID. It is
// restated rather than imported so the transport does not depend on the
// composition root.
const voiceIntroductionPurpose = "voice.introduction"

// voiceRetention is the Build Pack's RET-01 window for member voice.
const voiceRetention = 180 * 24 * time.Hour

func RegisterIntroductionRoutes(
	mux *http.ServeMux,
	service VoiceIntroduction,
	reader IntroductionReader,
	playback IntroductionPlayback,
	sessions SessionAuthenticator,
	gate MemberGate,
) {
	// Recording a Voice of Introduction and hearing one are romantic
	// surfaces (FR-101a). Reading your own back and taking it down are not:
	// erasure is a right, and a member demoted for safety must still be able
	// to remove their own recording.
	mux.Handle("POST /v1/introductions", gate.guard(sessions, "introductions.view", "introduction", beginIntroductionHandler(service, sessions)))
	mux.Handle("POST /v1/introductions/{id}/uploaded", gate.guard(sessions, "introductions.view", "introduction", confirmIntroductionHandler(service, sessions, reader)))
	mux.Handle("GET /v1/introductions/{id}", readIntroductionHandler(reader, sessions))
	mux.Handle("GET /v1/introductions/{id}/audio", gate.guard(sessions, "introductions.view", "introduction", playIntroductionHandler(playback, reader, sessions)))
	mux.Handle("DELETE /v1/introductions/{id}", revokeIntroductionHandler(service, sessions, reader))
}

type playbackResponse struct {
	AssetID    string `json:"assetId"`
	URL        string `json:"url"`
	ExpiresAt  string `json:"expiresAt"`
	DurationMs int64  `json:"durationMs"`
}

// playIntroductionHandler hands back a URL the client plays the audio from.
//
// Without this the recording could not be heard by anybody, including the
// member who made it — and the twenty seconds of verified listening that arms
// Sow (FR-202) could never be accumulated, because there was nothing to
// listen to. The audio is streamed from storage, not proxied: a play request
// per member per replay is not traffic this service should carry.
func playIntroductionHandler(
	playback IntroductionPlayback,
	reader IntroductionReader,
	sessions SessionAuthenticator,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if playback == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "introduction_unavailable", Message: "Voice introductions are unavailable.",
			})
			return
		}
		introduction, found := loadOwnedIntroduction(w, r, reader, r.PathValue("id"), memberID)
		if !found {
			return
		}
		// Withdrawal stops playback now, not when erasure catches up. Revoke
		// already moves the data to purge_pending, so the data check alone
		// would cover it; the status check is stated too because "may this be
		// heard" should not depend on a reader tracing which transition
		// happens to move which field.
		//
		// Saying "not found" also beats a signed URL that 404s at the bucket,
		// which reads to a member as a broken player rather than their own
		// decision being honoured.
		if introduction.Status() == domain.StatusRevoked ||
			introduction.Status() == domain.StatusCancelled ||
			introduction.DataStatus() != domain.DataRetained {
			writeError(w, r, http.StatusNotFound, APIError{
				Code: "introduction_not_found", Message: "That voice introduction was not found.",
			})
			return
		}

		access, err := playback.AuthorizePlayback(r.Context(), memberID, introduction.Media().AssetID())
		if err != nil {
			writeIntroductionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, playbackResponse{
			AssetID:    introduction.Media().AssetID(),
			URL:        access.URL,
			ExpiresAt:  access.ExpiresAt.Format(time.RFC3339),
			DurationMs: introduction.Media().Duration().Milliseconds(),
		})
	})
}

type beginIntroductionRequest struct {
	ContentType string `json:"contentType"`
}

type introductionResponse struct {
	IntroductionID string `json:"introductionId"`
	Status         string `json:"status"`
	DataStatus     string `json:"dataStatus"`
	AssetID        string `json:"assetId"`
	ContentType    string `json:"contentType"`
	DurationMs     int64  `json:"durationMs"`
	SizeBytes      int64  `json:"sizeBytes"`
	TranscriptID   string `json:"transcriptId,omitempty"`
	UploadURL      string `json:"uploadUrl,omitempty"`
	UploadExpires  string `json:"uploadExpiresAt,omitempty"`
}

func toIntroductionResponse(introduction domain.Introduction) introductionResponse {
	return introductionResponse{
		IntroductionID: introduction.ID(),
		Status:         string(introduction.Status()),
		DataStatus:     string(introduction.DataStatus()),
		AssetID:        introduction.Media().AssetID(),
		ContentType:    introduction.Media().ContentType(),
		DurationMs:     introduction.Media().Duration().Milliseconds(),
		SizeBytes:      introduction.Media().Size(),
		TranscriptID:   introduction.Transcript().ID(),
	}
}

// beginIntroductionHandler opens a recording and hands back a write grant.
//
// The audio is never posted here. The member's browser sends it straight to
// object storage with the signed URL this returns, which keeps a two-minute
// clip off the API's request path entirely and means a failed upload costs
// the member a retry rather than the service a timeout.
func beginIntroductionHandler(service VoiceIntroduction, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code: "unsupported_media_type", Message: "Content-Type must be application/json.",
			})
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if !validIdentifier(commandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "A valid Idempotency-Key is required.",
			})
			return
		}
		var body beginIntroductionRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}

		result, err := service.BeginUpload(r.Context(), application.BeginUploadRequest{
			CommandID:      commandID,
			OwnerID:        memberID,
			PurposeID:      voiceIntroductionPurpose,
			PurposeVersion: 1,
			ContentType:    body.ContentType,
			RetentionUntil: time.Now().UTC().Add(voiceRetention),
		})
		if err != nil {
			writeIntroductionError(w, r, err)
			return
		}

		response := toIntroductionResponse(result.Introduction)
		response.UploadURL = result.Access.URL
		if !result.Access.ExpiresAt.IsZero() {
			response.UploadExpires = result.Access.ExpiresAt.Format(time.RFC3339)
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, response)
	})
}

// confirmIntroductionHandler is called once the bytes have landed. It reads
// back what storage actually accepted rather than trusting the client, then
// starts transcription in the same request so a member is not left holding a
// recording that never progresses.
func confirmIntroductionHandler(
	service VoiceIntroduction,
	sessions SessionAuthenticator,
	reader IntroductionReader,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if !validIdentifier(commandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "A valid Idempotency-Key is required.",
			})
			return
		}
		id := r.PathValue("id")
		if !ownsIntroduction(w, r, reader, id, memberID) {
			return
		}

		if _, err := service.ConfirmUpload(r.Context(), id, commandID); err != nil {
			writeIntroductionError(w, r, err)
			return
		}
		// Transcription is deferred by policy in this deployment, so this
		// returns "uncertain" rather than a transcript. The recording is
		// complete and playable either way.
		result, err := service.Transcribe(r.Context(), id, commandID+":transcribe")
		if err != nil {
			writeIntroductionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toIntroductionResponse(result.Introduction))
	})
}

func readIntroductionHandler(reader IntroductionReader, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		introduction, found := loadOwnedIntroduction(w, r, reader, r.PathValue("id"), memberID)
		if !found {
			return
		}
		writeSuccess(w, r, http.StatusOK, toIntroductionResponse(introduction))
	})
}

// revokeIntroductionHandler withdraws a recording. Revoke is not delete: the
// aggregate stops the transcription, marks the data for erasure and keeps the
// audit trail, which is what makes the erasure provable later.
func revokeIntroductionHandler(
	service VoiceIntroduction,
	sessions SessionAuthenticator,
	reader IntroductionReader,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if !validIdentifier(commandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "A valid Idempotency-Key is required.",
			})
			return
		}
		id := r.PathValue("id")
		if !ownsIntroduction(w, r, reader, id, memberID) {
			return
		}
		introduction, err := service.Revoke(r.Context(), id, commandID)
		if err != nil {
			writeIntroductionError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toIntroductionResponse(introduction))
	})
}

// loadOwnedIntroduction refuses another member's recording as absent rather
// than forbidden. Answering 403 would confirm the id exists, which is a
// membership disclosure on a surface built to avoid exactly that.
func loadOwnedIntroduction(
	w http.ResponseWriter,
	r *http.Request,
	reader IntroductionReader,
	id, memberID string,
) (domain.Introduction, bool) {
	if reader == nil {
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "introduction_unavailable", Message: "Voice introductions are unavailable.",
		})
		return domain.Introduction{}, false
	}
	introduction, err := reader.FindByID(r.Context(), strings.TrimSpace(id))
	if err != nil || introduction.OwnerID() != memberID {
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "introduction_not_found", Message: "That voice introduction was not found.",
		})
		return domain.Introduction{}, false
	}
	return introduction, true
}

func ownsIntroduction(
	w http.ResponseWriter,
	r *http.Request,
	reader IntroductionReader,
	id, memberID string,
) bool {
	_, ok := loadOwnedIntroduction(w, r, reader, id, memberID)
	return ok
}

func writeIntroductionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrConsentRequired):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "consent_required",
			Message: "Recording a voice introduction needs your consent for that purpose.",
		})
	case errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "introduction_not_found", Message: "That voice introduction was not found.",
		})
	case errors.Is(err, application.ErrOptimisticConflict),
		errors.Is(err, application.ErrCommandAlreadyUsed):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "introduction_conflict",
			Message: "That recording changed while this request was in flight. Try again.",
		})
	case errors.Is(err, domain.ErrInvalidTransition), errors.Is(err, domain.ErrInvalidIntroduction):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "That recording cannot change in the way requested.",
		})
	case errors.Is(err, application.ErrDependencyUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "introduction_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "introduction_unavailable", Message: "Voice introductions are temporarily unavailable.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
