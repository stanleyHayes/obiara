package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/domain"
)

type CircleRooms interface {
	Voice(context.Context, application.Create) (domain.Entry, error)
	Event(context.Context, application.Create) (domain.Entry, error)
	Notice(context.Context, application.Create) (domain.Entry, error)
	List(context.Context, string, string, int) ([]domain.Entry, error)
	Delete(context.Context, string, string, string) (domain.Entry, error)
}

func RegisterCircleRoomRoutes(mux *http.ServeMux, rooms CircleRooms, sessions SessionAuthenticator, gate MemberGate) {
	mux.Handle("GET /v1/circles/{circleId}/room", listCircleRoomHandler(rooms, sessions))
	mux.Handle("POST /v1/circles/{circleId}/room", gate.guard(sessions, "circles.participate", "circle", createCircleRoomEntryHandler(rooms, sessions)))
	mux.Handle("POST /v1/circle-room-entries/{id}/delete", deleteCircleRoomEntryHandler(rooms, sessions))
}

type circleRoomEntryResponse struct {
	ID           string     `json:"id"`
	CircleID     string     `json:"circleId"`
	Kind         string     `json:"kind"`
	ContentRef   string     `json:"contentRef,omitempty"`
	AssetID      string     `json:"assetId,omitempty"`
	TranscriptID string     `json:"transcriptId,omitempty"`
	ContentType  string     `json:"contentType,omitempty"`
	DurationMs   int64      `json:"durationMs,omitempty"`
	StartsAt     *time.Time `json:"startsAt,omitempty"`
	EndsAt       *time.Time `json:"endsAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	Revision     uint64     `json:"revision"`
}

func projectCircleRoomEntry(entry domain.Entry) circleRoomEntryResponse {
	response := circleRoomEntryResponse{
		ID: entry.ID(), CircleID: entry.CircleID(), Kind: string(entry.Kind()),
		ContentRef: entry.ContentRef(), AssetID: entry.Media().AssetID(),
		TranscriptID: entry.Media().TranscriptID(), ContentType: entry.Media().ContentType(),
		DurationMs: entry.Media().Duration().Milliseconds(), CreatedAt: entry.CreatedAt(),
		ExpiresAt: entry.ExpiresAt(), Revision: entry.Revision(),
	}
	if !entry.StartsAt().IsZero() {
		value := entry.StartsAt()
		response.StartsAt = &value
	}
	if !entry.EndsAt().IsZero() {
		value := entry.EndsAt()
		response.EndsAt = &value
	}
	return response
}

func listCircleRoomHandler(rooms CircleRooms, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		limit := 50
		if raw := r.URL.Query().Get("limit"); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value < 1 || value > 100 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Limit must be between 1 and 100."})
				return
			}
			limit = value
		}
		entries, err := rooms.List(r.Context(), r.PathValue("circleId"), memberID, limit)
		if err != nil {
			writeCircleRoomError(w, r, err)
			return
		}
		items := make([]circleRoomEntryResponse, 0, len(entries))
		for _, entry := range entries {
			items = append(items, projectCircleRoomEntry(entry))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Items []circleRoomEntryResponse `json:"items"`
		}{Items: items})
	})
}

type createCircleRoomEntryRequest struct {
	Kind          string    `json:"kind"`
	ContentRef    string    `json:"contentRef,omitempty"`
	AssetID       string    `json:"assetId,omitempty"`
	TranscriptID  string    `json:"transcriptId,omitempty"`
	ContentType   string    `json:"contentType,omitempty"`
	DurationMs    int64     `json:"durationMs,omitempty"`
	StartsAt      time.Time `json:"startsAt,omitempty"`
	EndsAt        time.Time `json:"endsAt,omitempty"`
	RetentionDays int       `json:"retentionDays"`
}

func createCircleRoomEntryHandler(rooms CircleRooms, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body createCircleRoomEntryRequest
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		create := application.Create{
			CommandID: commandID, CircleID: r.PathValue("circleId"), ActorID: memberID,
			ContentRef: strings.TrimSpace(body.ContentRef), StartsAt: body.StartsAt,
			EndsAt: body.EndsAt, Retention: time.Duration(body.RetentionDays) * 24 * time.Hour,
		}
		var (
			entry domain.Entry
			err   error
		)
		switch domain.Kind(body.Kind) {
		case domain.KindVoice:
			create.Media, err = domain.NewMediaRef(
				body.AssetID, body.TranscriptID, body.ContentType,
				time.Duration(body.DurationMs)*time.Millisecond,
			)
			if err == nil {
				entry, err = rooms.Voice(r.Context(), create)
			}
		case domain.KindEvent:
			entry, err = rooms.Event(r.Context(), create)
		case domain.KindNotice:
			entry, err = rooms.Notice(r.Context(), create)
		default:
			err = domain.ErrInvalidEntry
		}
		if err != nil {
			writeCircleRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectCircleRoomEntry(entry))
	})
}

func deleteCircleRoomEntryHandler(rooms CircleRooms, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct{}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		entry, err := rooms.Delete(r.Context(), r.PathValue("id"), memberID, commandID)
		if err != nil {
			writeCircleRoomError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectCircleRoomEntry(entry))
	})
}

func writeCircleRoomError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrDenied), errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "circle_room_not_found", Message: "That room or entry is not available."})
	case errors.Is(err, application.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{Code: "circle_room_conflict", Message: "The room entry changed. Refresh and try again."})
	case errors.Is(err, domain.ErrInvalidEntry), errors.Is(err, domain.ErrRetention):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "The room entry is invalid."})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
