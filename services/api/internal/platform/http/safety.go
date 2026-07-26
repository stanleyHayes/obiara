package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

// Safety is the inbound port for report/block intake (E12-S01).
type Safety interface {
	File(ctx context.Context, reporterID, subjectID string, category domain.Category, surface domain.Surface, contextRef, reason string) (string, domain.Tier, error)
	Block(ctx context.Context, blockerID, blockedID string) error
	Unblock(ctx context.Context, blockerID, blockedID string) error
}

// RegisterSafetyRoutes adds report and block routes.
func RegisterSafetyRoutes(mux *http.ServeMux, safety Safety) {
	mux.Handle("POST /v1/reports", fileReportHandler(safety))
	mux.Handle("POST /v1/blocks", blockHandler(safety))
	mux.Handle("DELETE /v1/blocks/{blockerId}/{blockedId}", unblockHandler(safety))
}

type fileReportRequest struct {
	ReporterID string `json:"reporterId"`
	SubjectID  string `json:"subjectId"`
	Category   string `json:"category"`
	Surface    string `json:"surface"`
	ContextRef string `json:"contextRef,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type reportAckResponse struct {
	ReportID  string    `json:"reportId"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"createdAt"`
}

func fileReportHandler(safety Safety) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body fileReportRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.ReporterID = strings.TrimSpace(body.ReporterID)
		body.SubjectID = strings.TrimSpace(body.SubjectID)
		var details []FieldError
		if !validOpaqueID(body.ReporterID) {
			details = append(details, FieldError{Field: "reporterId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if !validOpaqueID(body.SubjectID) {
			details = append(details, FieldError{Field: "subjectId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		id, tier, err := safety.File(r.Context(), body.ReporterID, body.SubjectID,
			domain.Category(body.Category), domain.Surface(body.Surface), body.ContextRef, body.Reason)
		if err != nil {
			writeSafetyError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, reportAckResponse{ReportID: id, Tier: string(tier), CreatedAt: time.Now().UTC()})
	})
}

type blockRequest struct {
	BlockerID string `json:"blockerId"`
	BlockedID string `json:"blockedId"`
}

func blockHandler(safety Safety) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body blockRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.BlockerID = strings.TrimSpace(body.BlockerID)
		body.BlockedID = strings.TrimSpace(body.BlockedID)
		if !validOpaqueID(body.BlockerID) || !validOpaqueID(body.BlockedID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "blockerId/blockedId", Reason: "must be opaque identifiers"}},
			})
			return
		}

		if err := safety.Block(r.Context(), body.BlockerID, body.BlockedID); err != nil {
			writeSafetyError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, struct {
			Blocked bool `json:"blocked"`
		}{Blocked: true})
	})
}

func unblockHandler(safety Safety) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := safety.Unblock(r.Context(), r.PathValue("blockerId"), r.PathValue("blockedId")); err != nil {
			writeSafetyError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Blocked bool `json:"blocked"`
		}{Blocked: false})
	})
}

func writeSafetyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrSelfReport), errors.Is(err, domain.ErrInvalidCategory),
		errors.Is(err, domain.ErrInvalidSurface), errors.Is(err, domain.ErrReasonTooLong),
		errors.Is(err, domain.ErrReporterRequired), errors.Is(err, domain.ErrSubjectRequired):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "One or more fields are invalid.",
		})
	case errors.Is(err, application.ErrBlockExists):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "block_exists",
			Message: "That member is already blocked.",
		})
	case errors.Is(err, application.ErrBlockNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "block_not_found",
			Message: "No block to remove.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
