package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	podapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/application"
	poddomain "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
)

// Pods is the inbound port for the pod: a recording resting at a member's
// house front for named people to open.
type Pods interface {
	Create(context.Context, podapplication.Command, podapplication.Proposal) (podapplication.Result, error)
	Playback(context.Context, podapplication.Command) (podapplication.Result, error)
	Resting(ctx context.Context, memberID string, limit int) ([]poddomain.Pod, error)
}

// podTTL is how long a pod rests before it closes. The aggregate refuses
// anything past a week, and a pod nobody opened in that time has been
// answered by silence.
const podTTL = 7 * 24 * time.Hour

// RegisterPodRoutes exposes placing a pod and opening one.
//
// There is no route to list somebody else's pods and none to read a pod's
// contents directly: opening one is a transition on the aggregate that
// records who listened, and a read that skipped it would be a way to hear
// somebody without them ever knowing they had been heard.
func RegisterPodRoutes(mux *http.ServeMux, pods Pods, sessions SessionAuthenticator, gate MemberGate) {
	mux.Handle("GET /v1/seed/pods", gate.guard(sessions, "seed.pod.playback", "pod", restingPodsHandler(pods, sessions)))
	mux.Handle("POST /v1/seed/pods", gate.guard(sessions, "seed.pod.create", "pod", createPodHandler(pods, sessions)))
	mux.Handle("POST /v1/seed/pods/{id}/playback", gate.guard(sessions, "seed.pod.playback", "pod", playPodHandler(pods, sessions)))
}

type createPodRequest struct {
	MediaRef     string   `json:"mediaRef"`
	RecipientIDs []string `json:"recipientIds"`
}

type podResponse struct {
	PodID    string `json:"podId"`
	Status   string `json:"status"`
	Replayed bool   `json:"replayed"`
	// PlaybackURL is present only on an open, and only briefly.
	PlaybackURL string `json:"playbackUrl,omitempty"`
}

func createPodHandler(pods Pods, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if commandID == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "Idempotency-Key", Reason: "is required"}},
			})
			return
		}
		var body createPodRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if pods == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		result, err := pods.Create(r.Context(),
			podapplication.Command{ID: commandID, ActorID: actorID},
			podapplication.Proposal{
				// The owner is the session. A member cannot place a pod at
				// somebody else's house front.
				OwnerID:      actorID,
				MediaRef:     strings.TrimSpace(body.MediaRef),
				RecipientIDs: body.RecipientIDs,
				TTL:          podTTL,
			})
		if err != nil {
			writePodError(w, r, err)
			return
		}
		status := http.StatusCreated
		if result.Replayed {
			status = http.StatusOK
		}
		writeSuccess(w, r, status, podResponse{
			PodID: result.Pod.ID(), Status: string(result.Pod.Status()), Replayed: result.Replayed,
		})
	})
}

func playPodHandler(pods Pods, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if commandID == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "Idempotency-Key", Reason: "is required"}},
			})
			return
		}
		if pods == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		result, err := pods.Playback(r.Context(), podapplication.Command{
			ID: commandID, ActorID: actorID, PodID: r.PathValue("id"),
		})
		if err != nil {
			writePodError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, podResponse{
			PodID: result.Pod.ID(), Status: string(result.Pod.Status()),
			Replayed: result.Replayed, PlaybackURL: result.PlaybackToken,
		})
	})
}

func writePodError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, podapplication.ErrNotFound), errors.Is(err, podapplication.ErrNotAvailable):
		// One answer for "no such pod", "not for you", "expired", "revoked"
		// and "one of you blocked the other". Telling them apart would say
		// which, and every one of those is somebody's business but the
		// caller's.
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "pod_not_available", Message: "That pod is not available.",
		})
	case errors.Is(err, podapplication.ErrOptimisticConflict):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "pod_conflict", Message: "That pod changed. Refresh and try again.",
		})
	case errors.Is(err, poddomain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, poddomain.ErrInvalidPod), errors.Is(err, poddomain.ErrInvalidTransition):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, podapplication.ErrUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "feature_unavailable", Message: "This is not available right now.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}

type restingPodResponse struct {
	PodID string `json:"podId"`
	// ClosesAt is the only urgency a member is given. There is deliberately
	// nothing here about who left it: the pod keys its owner, and a member
	// hears who it is from by opening it, which is the whole shape of the
	// gesture.
	ClosesAt string `json:"closesAt"`
	Opened   bool   `json:"opened"`
}

type restingPodsResponse struct {
	Pods []restingPodResponse `json:"pods"`
}

func restingPodsHandler(pods Pods, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if pods == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		resting, err := pods.Resting(r.Context(), memberID, limit)
		if err != nil {
			writePodError(w, r, err)
			return
		}
		response := restingPodsResponse{Pods: make([]restingPodResponse, 0, len(resting))}
		for _, item := range resting {
			opened := false
			for _, event := range item.Events() {
				if event.Action == poddomain.ActionPlayed {
					opened = true
					break
				}
			}
			response.Pods = append(response.Pods, restingPodResponse{
				PodID:    item.ID(),
				ClosesAt: item.ExpiresAt().UTC().Format(time.RFC3339),
				Opened:   opened,
			})
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}
