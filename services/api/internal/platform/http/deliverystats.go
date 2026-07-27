package apihttp

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/obiara/internal/notifications/deliverystats/application"
)

// DeliveryStats is the inbound port for delivery observability (E13-S08).
type DeliveryStats interface {
	Stats(ctx context.Context, windowDays int) (application.StatsReport, error)
}

// RegisterDeliveryStatsRoutes adds the delivery statistics route.
func RegisterDeliveryStatsRoutes(mux *http.ServeMux, stats DeliveryStats) {
	mux.Handle("GET /v1/metrics/deliveries", deliveryStatsHandler(stats))
}

func deliveryStatsHandler(stats DeliveryStats) http.Handler {
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
		report, err := stats.Stats(r.Context(), days)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
			return
		}
		writeSuccess(w, r, http.StatusOK, report)
	})
}
