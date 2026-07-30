package apihttp

import (
	"context"
	"net/http"
	"time"

	gardenapp "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/application"
	gardendomain "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
)

type Garden interface {
	Summary(context.Context, string) (gardendomain.Summary, error)
}

func RegisterGardenRoutes(mux *http.ServeMux, garden Garden, sessions SessionAuthenticator) {
	mux.Handle("GET /v1/garden", gardenSummaryHandler(garden, sessions))
}

type gardenSummaryResponse struct {
	AsOf          time.Time `json:"asOf"`
	MovingQuietly int       `json:"movingQuietly"`
	Sprouts       int       `json:"sprouts"`
	Message       string    `json:"message"`
}

func gardenSummaryHandler(garden Garden, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		summary, err := garden.Summary(r.Context(), memberID)
		if err != nil {
			if gardenapp.IsNotFound(err) {
				writeError(w, r, http.StatusNotFound, APIError{Code: "garden_not_found", Message: "Your garden is not available."})
				return
			}
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "Your garden could not be loaded."})
			return
		}
		writeSuccess(w, r, http.StatusOK, gardenSummaryResponse{
			AsOf: summary.AsOf, MovingQuietly: summary.MovingQuietly,
			Sprouts: summary.Sprouts, Message: summary.Message,
		})
	})
}
