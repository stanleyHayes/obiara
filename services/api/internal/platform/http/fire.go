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
func RegisterFireRoutes(mux *http.ServeMux, fires Fires) {
	mux.Handle("POST /v1/fires", scheduleFireHandler(fires))
	mux.Handle("GET /v1/fires", listFiresHandler(fires))
	mux.Handle("POST /v1/fires/{id}/rsvps", rsvpHandler(fires))
	mux.Handle("DELETE /v1/fires/{id}/rsvps/{memberId}", cancelRsvpHandler(fires))
	mux.Handle("POST /v1/fires/{id}/close", closeFireHandler(fires))
}

type scheduleFireRequest struct {
	HostID   string `json:"hostId"`
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

func scheduleFireHandler(fires Fires) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		body.HostID = strings.TrimSpace(body.HostID)
		body.Title = strings.TrimSpace(body.Title)
		startsAt, startErr := time.Parse(time.RFC3339, strings.TrimSpace(body.StartsAt))
		var details []FieldError
		if !validOpaqueID(body.HostID) {
			details = append(details, FieldError{Field: "hostId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
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

		fire, err := fires.Schedule(r.Context(), body.HostID, body.CircleID, body.Title, startsAt, body.Capacity)
		if err != nil {
			writeFireError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toFireResponse(fire))
	})
}

func listFiresHandler(fires Fires) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	MemberID string `json:"memberId"`
	Tier     int    `json:"tier"`
}

type rsvpResponse struct {
	Status   string `json:"status"`
	Position int    `json:"position,omitempty"`
}

func rsvpHandler(fires Fires) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		body.MemberID = strings.TrimSpace(body.MemberID)
		if !validOpaqueID(body.MemberID) || body.Tier < 0 || body.Tier > 2 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId/tier", Reason: "memberId must be an opaque id and tier 0-2"}},
			})
			return
		}

		rsvp, err := fires.RSVP(r.Context(), r.PathValue("id"), body.MemberID, body.Tier)
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

func cancelRsvpHandler(fires Fires) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promoted, err := fires.Cancel(r.Context(), r.PathValue("id"), r.PathValue("memberId"))
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
	ActorID string `json:"actorId"`
}

type closeFireResponse struct {
	Status    string   `json:"status"`
	Attendees []string `json:"attendees"`
}

func closeFireHandler(fires Fires) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		body.ActorID = strings.TrimSpace(body.ActorID)
		if !validOpaqueID(body.ActorID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "actorId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		attendees, err := fires.CloseToEmbers(r.Context(), r.PathValue("id"), body.ActorID)
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
