package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// Fires is the inbound port for fire scheduling and attendance (E09-S01).
type Fires interface {
	Schedule(ctx context.Context, hostID, circleID, title string, startsAt time.Time, capacity int) (domain.Fire, error)
	RSVP(ctx context.Context, fireID, memberID string, tier int) (domain.RSVP, error)
	Cancel(ctx context.Context, fireID, memberID string) (*domain.RSVP, error)
	Upcoming(ctx context.Context, limit int) ([]domain.Fire, error)
	CloseToEmbers(ctx context.Context, fireID, actorID string) ([]domain.RSVP, error)
}

// RegisterFireRoutes adds fire scheduling and RSVP routes.
type TierReader interface {
	Tier(context.Context, string) (identitydomain.Tier, error)
}

func RegisterFireRoutes(mux *http.ServeMux, fires Fires, sessions SessionAuthenticator, tiers TierReader, gate MemberGate) {
	// Attending already requires Tier 1 inside the aggregate; scheduling one
	// is the same kind of participation and was not gated at all.
	mux.Handle("POST /v1/fires", gate.guard(sessions, "fires.attend", "fire", scheduleFireHandler(fires, sessions)))
	mux.Handle("GET /v1/fires", listFiresHandler(fires, sessions))
	mux.Handle("POST /v1/fires/{id}/rsvps", rsvpHandler(fires, sessions, tiers))
	mux.Handle("DELETE /v1/fires/{id}/rsvps/{memberId}", cancelRsvpHandler(fires, sessions))
	mux.Handle("POST /v1/fires/{id}/close", closeFireHandler(fires, sessions))
}

type scheduleFireRequest struct {
	CircleID string `json:"circleId,omitempty"`
	Title    string `json:"title"`
	StartsAt string `json:"startsAt"`
	Capacity int    `json:"capacity"`
}

type fireResponse struct {
	FireID     string    `json:"fireId"`
	HostID     string    `json:"hostId"`
	CircleID   string    `json:"circleId,omitempty"`
	Title      string    `json:"title"`
	StartsAt   time.Time `json:"startsAt"`
	Capacity   int       `json:"capacity"`
	GoingCount int       `json:"goingCount"`
	Status     string    `json:"status"`
}

func toFireResponse(fire domain.Fire) fireResponse {
	return fireResponse{
		FireID:     fire.ID(),
		HostID:     fire.HostID(),
		CircleID:   fire.CircleID(),
		Title:      fire.Title(),
		StartsAt:   fire.StartsAt(),
		Capacity:   fire.Capacity(),
		GoingCount: fire.GoingCount(),
		Status:     string(fire.Status()),
	}
}

func fireMember(w http.ResponseWriter, r *http.Request, sessions SessionAuthenticator) (string, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok || sessions == nil {
		writeError(w, r, http.StatusUnauthorized, APIError{Code: "authentication_required", Message: "A valid member session is required."})
		return "", false
	}
	session, err := sessions.Authenticate(r.Context(), token)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, APIError{Code: "authentication_required", Message: "A valid member session is required."})
		return "", false
	}
	return session.MemberID(), true
}

func scheduleFireHandler(fires Fires, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostID, ok := fireMember(w, r, sessions)
		if !ok {
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body scheduleFireRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.Title = strings.TrimSpace(body.Title)
		startsAt, startErr := time.Parse(time.RFC3339, strings.TrimSpace(body.StartsAt))
		var details []FieldError
		if body.CircleID != "" && !validOpaqueID(body.CircleID) {
			details = append(details, FieldError{Field: "circleId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if body.Title == "" || len(body.Title) > 120 {
			details = append(details, FieldError{Field: "title", Reason: "must be 1-120 characters"})
		}
		if startErr != nil {
			details = append(details, FieldError{Field: "startsAt", Reason: "must be RFC 3339"})
		}
		if body.Capacity < 1 || body.Capacity > domain.MaxCapacity {
			details = append(details, FieldError{Field: "capacity", Reason: "must be 1-500"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		fire, err := fires.Schedule(r.Context(), hostID, body.CircleID, body.Title, startsAt, body.Capacity)
		if err != nil {
			writeFireError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toFireResponse(fire))
	})
}

func listFiresHandler(fires Fires, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := fireMember(w, r, sessions); !ok {
			return
		}
		circleID := strings.TrimSpace(r.URL.Query().Get("circleId"))
		if circleID != "" && !validOpaqueID(circleID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "circleId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		upcoming, err := fires.Upcoming(r.Context(), 20)
		if err != nil {
			writeFireError(w, r, err)
			return
		}
		response := make([]fireResponse, 0, len(upcoming))
		for _, fire := range upcoming {
			if circleID != "" && fire.CircleID() != circleID {
				continue
			}
			response = append(response, toFireResponse(fire))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Fires []fireResponse `json:"fires"`
		}{Fires: response})
	})
}

type rsvpRequest struct {
}

type rsvpResponse struct {
	Status   string `json:"status"`
	Position int    `json:"position,omitempty"`
}

func rsvpHandler(fires Fires, sessions SessionAuthenticator, tiers TierReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := fireMember(w, r, sessions)
		if !ok {
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body rsvpRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		if tiers == nil {
			writeFireError(w, r, errors.New("tier service unavailable"))
			return
		}
		tier, err := tiers.Tier(r.Context(), memberID)
		if err != nil {
			writeFireError(w, r, err)
			return
		}
		rsvp, err := fires.RSVP(r.Context(), r.PathValue("id"), memberID, int(tier))
		if err != nil {
			writeFireError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, rsvpResponse{Status: string(rsvp.Status()), Position: rsvp.Position()})
	})
}

type cancelRsvpResponse struct {
	Cancelled        bool   `json:"cancelled"`
	PromotedMemberID string `json:"promotedMemberId,omitempty"`
}

func cancelRsvpHandler(fires Fires, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := fireMember(w, r, sessions)
		if !ok {
			return
		}
		if memberID != r.PathValue("memberId") {
			writeError(w, r, http.StatusForbidden, APIError{Code: "access_denied", Message: "That RSVP belongs to another member."})
			return
		}
		promoted, err := fires.Cancel(r.Context(), r.PathValue("id"), memberID)
		if err != nil {
			writeFireError(w, r, err)
			return
		}
		response := cancelRsvpResponse{Cancelled: true}
		if promoted != nil {
			response.PromotedMemberID = promoted.MemberID()
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

type closeFireRequest struct {
}

type closeFireResponse struct {
	Status    string   `json:"status"`
	Attendees []string `json:"attendees"`
}

func closeFireHandler(fires Fires, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := fireMember(w, r, sessions)
		if !ok {
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}
		var body closeFireRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}
		attendees, err := fires.CloseToEmbers(r.Context(), r.PathValue("id"), actorID)
		if err != nil {
			writeFireError(w, r, err)
			return
		}
		response := closeFireResponse{Status: string(domain.StatusEmbers), Attendees: make([]string, 0, len(attendees))}
		for _, attendee := range attendees {
			response.Attendees = append(response.Attendees, attendee.MemberID())
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

func writeFireError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotHost):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "not_host",
			Message: "Only the host can close a fire.",
		})
	case errors.Is(err, domain.ErrFireNotClosable):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "fire_not_closable",
			Message: "This fire cannot dim to embers right now.",
		})
	case errors.Is(err, domain.ErrTierTooLow):
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "tier_too_low",
			Message: "Fire entry requires full verification.",
		})
	case errors.Is(err, domain.ErrFireNotOpen):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "fire_not_open",
			Message: "This fire is not open for RSVP.",
		})
	case errors.Is(err, application.ErrAlreadyRSVPed):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "rsvp_exists",
			Message: "You already have a place at this fire.",
		})
	case errors.Is(err, application.ErrFireNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "fire_not_found",
			Message: "No such fire.",
		})
	case errors.Is(err, application.ErrRSVPNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "rsvp_not_found",
			Message: "No RSVP to cancel.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
