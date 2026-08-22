package apihttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
)

// Allowances is the inbound port for the weekly seed allowance.
type Allowances interface {
	CurrentOrIssue(ctx context.Context, subjectID, commandID string) (domain.Ledger, error)
}

// RegisterSeedAllowanceRoutes adds the allowance route. A member reads only
// their own budget; the subject comes from the session, never the path.
func RegisterSeedAllowanceRoutes(mux *http.ServeMux, allowances Allowances, sessions SessionAuthenticator) {
	mux.Handle("GET /v1/seed/allowance", readAllowanceHandler(allowances, sessions))
}

type allowanceResponse struct {
	Balance         int64  `json:"balance"`
	WeeklyAllowance int64  `json:"weeklyAllowance"`
	WeekStart       string `json:"weekStart"`
}

func readAllowanceHandler(allowances Allowances, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		// The read may issue this week's ledger, so it carries a command id
		// derived from the member and the day. Two reads in the same week
		// then replay rather than issuing twice.
		commandID := "allowance-read-" + memberID + "-" + time.Now().UTC().Format("2006-01-02")

		ledger, err := allowances.CurrentOrIssue(r.Context(), memberID, commandID)
		if err != nil {
			writeAllowanceError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, allowanceResponse{
			Balance: ledger.Balance(), WeeklyAllowance: ledger.WeeklyAllowance(),
			WeekStart: ledger.WeekStart().UTC().Format(time.RFC3339),
		})
	})
}

func writeAllowanceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrConcurrentChange):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "allowance_conflict", Message: "Your allowance changed. Try again.",
		})
	case errors.Is(err, domain.ErrInvalidLedger), errors.Is(err, domain.ErrCommandConflict):
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
