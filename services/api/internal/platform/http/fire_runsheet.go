package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/domain"
)

// RunSheets is the inbound port for the fire run sheet: the ordered segments
// a host works through while running a fire.
type RunSheets interface {
	Create(ctx context.Context, command application.Command, version uint64, segments []domain.Segment) (domain.RunSheet, error)
	Start(context.Context, application.Command) (domain.Projection, error)
	Advance(context.Context, application.Command) (domain.Projection, error)
	Skip(context.Context, application.Command) (domain.Projection, error)
	Extend(ctx context.Context, command application.Command, by time.Duration) (domain.Projection, error)
	View(context.Context, application.Command) (domain.Projection, error)
}

// RegisterFireRunSheetRoutes adds the run sheet routes. Every one is
// host-only; the service settles that before it touches the aggregate.
func RegisterFireRunSheetRoutes(mux *http.ServeMux, sheets RunSheets, sessions SessionAuthenticator) {
	mux.Handle("POST /v1/fires/{fireId}/run-sheet", createRunSheetHandler(sheets, sessions))
	mux.Handle("GET /v1/fires/{fireId}/run-sheet/{id}", viewRunSheetHandler(sheets, sessions))
	mux.Handle("POST /v1/fires/{fireId}/run-sheet/{id}/start", runSheetMoveHandler(sheets, sessions, "start"))
	mux.Handle("POST /v1/fires/{fireId}/run-sheet/{id}/advance", runSheetMoveHandler(sheets, sessions, "advance"))
	mux.Handle("POST /v1/fires/{fireId}/run-sheet/{id}/skip", runSheetMoveHandler(sheets, sessions, "skip"))
	mux.Handle("POST /v1/fires/{fireId}/run-sheet/{id}/extend", extendRunSheetHandler(sheets, sessions))
}

type runSheetSegmentInput struct {
	Type           string `json:"type"`
	TitleCode      string `json:"titleCode"`
	PlannedMinutes int    `json:"plannedMinutes"`
	CapabilityRef  string `json:"capabilityRef"`
}

type createRunSheetRequest struct {
	CommandID string                 `json:"commandId"`
	Version   uint64                 `json:"version"`
	Segments  []runSheetSegmentInput `json:"segments"`
}

type runSheetCommandRequest struct {
	CommandID        string `json:"commandId"`
	ExpectedRevision uint64 `json:"expectedRevision"`
	// ByMinutes is used by the extend endpoint.
	ByMinutes int `json:"byMinutes"`
}

type runSheetSegmentResponse struct {
	Type           string `json:"type"`
	TitleCode      string `json:"titleCode"`
	PlannedMinutes int    `json:"plannedMinutes"`
}

type runSheetResponse struct {
	RunSheetID       string                   `json:"runSheetId"`
	Version          uint64                   `json:"version"`
	Status           string                   `json:"status"`
	CurrentIndex     int                      `json:"currentIndex"`
	Current          *runSheetSegmentResponse `json:"current,omitempty"`
	RemainingSeconds int                      `json:"remainingSeconds"`
	ServerTime       string                   `json:"serverTime"`
}

func toRunSheetResponse(projection domain.Projection) runSheetResponse {
	response := runSheetResponse{
		RunSheetID: projection.ID, Version: projection.Version,
		Status: string(projection.Status), CurrentIndex: projection.CurrentIndex,
		RemainingSeconds: int(projection.Remaining / time.Second),
		// The host's device counts down against the server's clock, not its
		// own, so a skewed handset cannot run the room fast or slow.
		ServerTime: projection.ServerTime.UTC().Format(time.RFC3339),
	}
	if projection.Current != nil {
		response.Current = &runSheetSegmentResponse{
			Type: string(projection.Current.Type), TitleCode: projection.Current.TitleCode,
			PlannedMinutes: int(projection.Current.PlannedDuration / time.Minute),
		}
	}
	return response
}

func runSheetCommand(w http.ResponseWriter, r *http.Request, sessions SessionAuthenticator) (application.Command, runSheetCommandRequest, bool) {
	var body runSheetCommandRequest
	if !proposalJSONGuard(w, r) {
		return application.Command{}, body, false
	}
	actorID, ok := authenticatedMember(w, r, sessions)
	if !ok {
		return application.Command{}, body, false
	}
	fireID, runSheetID := r.PathValue("fireId"), r.PathValue("id")
	if !validOpaqueID(fireID) || !validOpaqueID(runSheetID) {
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
		return application.Command{}, body, false
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, http.StatusBadRequest, APIError{
			Code: "invalid_json", Message: "The request body must be one valid JSON object.",
		})
		return application.Command{}, body, false
	}
	body.CommandID = strings.TrimSpace(body.CommandID)
	if !validOpaqueID(body.CommandID) {
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
			Details: []FieldError{{Field: "commandId", Reason: "must be an opaque identifier"}},
		})
		return application.Command{}, body, false
	}
	return application.Command{
		ID: body.CommandID, RunSheetID: runSheetID, FireID: fireID,
		ActorID: actorID, ExpectedRevision: body.ExpectedRevision,
	}, body, true
}

func createRunSheetHandler(sheets RunSheets, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !proposalJSONGuard(w, r) {
			return
		}
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		fireID := r.PathValue("fireId")
		if !validOpaqueID(fireID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		var body createRunSheetRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		if !validOpaqueID(body.CommandID) || len(body.Segments) == 0 || len(body.Segments) > 50 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "segments", Reason: "must be 1-50 segments"}},
			})
			return
		}

		segments := make([]domain.Segment, 0, len(body.Segments))
		for _, input := range body.Segments {
			segmentType := domain.SegmentType(strings.ToLower(strings.TrimSpace(input.Type)))
			switch segmentType {
			case domain.TypeTalk, domain.TypeBreak, domain.TypeGame, domain.TypeClose:
			default:
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "segments.type", Reason: "must be talk, break, game or close"}},
				})
				return
			}
			if input.PlannedMinutes < 1 || input.PlannedMinutes > 240 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "segments.plannedMinutes", Reason: "must be 1-240"}},
				})
				return
			}
			segments = append(segments, domain.Segment{
				Type: segmentType, TitleCode: strings.TrimSpace(input.TitleCode),
				PlannedDuration: time.Duration(input.PlannedMinutes) * time.Minute,
				CapabilityRef:   strings.TrimSpace(input.CapabilityRef),
			})
		}

		sheet, err := sheets.Create(r.Context(), application.Command{
			ID: body.CommandID, FireID: fireID, ActorID: actorID,
		}, body.Version, segments)
		if err != nil {
			writeRunSheetError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toRunSheetResponse(sheet.Project(time.Now().UTC())))
	})
}

func runSheetMoveHandler(sheets RunSheets, sessions SessionAuthenticator, action string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command, _, ok := runSheetCommand(w, r, sessions)
		if !ok {
			return
		}
		var projection domain.Projection
		var err error
		switch action {
		case "start":
			projection, err = sheets.Start(r.Context(), command)
		case "advance":
			projection, err = sheets.Advance(r.Context(), command)
		default:
			projection, err = sheets.Skip(r.Context(), command)
		}
		if err != nil {
			writeRunSheetError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toRunSheetResponse(projection))
	})
}

func extendRunSheetHandler(sheets RunSheets, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command, body, ok := runSheetCommand(w, r, sessions)
		if !ok {
			return
		}
		if body.ByMinutes < 1 || body.ByMinutes > 60 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "byMinutes", Reason: "must be 1-60"}},
			})
			return
		}
		projection, err := sheets.Extend(r.Context(), command, time.Duration(body.ByMinutes)*time.Minute)
		if err != nil {
			writeRunSheetError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toRunSheetResponse(projection))
	})
}

func viewRunSheetHandler(sheets RunSheets, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		fireID, runSheetID := r.PathValue("fireId"), r.PathValue("id")
		if !validOpaqueID(fireID) || !validOpaqueID(runSheetID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		projection, err := sheets.View(r.Context(), application.Command{
			RunSheetID: runSheetID, FireID: fireID, ActorID: actorID,
		})
		if err != nil {
			writeRunSheetError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toRunSheetResponse(projection))
	})
}

// writeRunSheetError maps the slice's failures.
//
// A run sheet the caller does not host and one that does not exist answer
// identically: the service collapses both into ErrNotAvailable before they
// reach here, so a member cannot probe which fires have a run sheet.
func writeRunSheetError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotAvailable), errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "run_sheet_not_found", Message: "That run sheet is not available.",
		})
	case errors.Is(err, application.ErrConflict), errors.Is(err, domain.ErrTransition):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "run_sheet_conflict", Message: "That run sheet changed. Refresh and try again.",
		})
	case errors.Is(err, application.ErrApplied):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, domain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
