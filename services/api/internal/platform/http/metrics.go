package apihttp

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/analytics/application"
)

// Metrics is the inbound port for P0 funnel metrics (E15-S07).
type Metrics interface {
	Funnel(ctx context.Context, windowDays int) (application.FunnelReport, error)
}

// RegisterMetricsRoutes adds the metrics routes.
func RegisterMetricsRoutes(mux *http.ServeMux, metrics Metrics) {
	mux.Handle("GET /v1/metrics/funnel", funnelHandler(metrics))
}

func funnelHandler(metrics Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		days := 30
		if raw := strings.TrimSpace(r.URL.Query().Get("days")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 90 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code:    "validation_failed",
					Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "days", Reason: "must be 1-90"}},
				})
				return
			}
			days = parsed
		}
		report, err := metrics.Funnel(r.Context(), days)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
			return
		}
		writeSuccess(w, r, http.StatusOK, report)
	})
}
