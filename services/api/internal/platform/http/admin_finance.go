package apihttp

import (
	"context"
	"net/http"
	"time"

	reconciliationapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/application"
	reconciliationdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type AdminFinanceOverview interface {
	Overview(context.Context, int) (reconciliationapp.Overview, error)
}

func RegisterAdminFinanceRoutes(mux *http.ServeMux, queries AdminFinanceOverview, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/finance/reconciliation", adminFinanceOverviewHandler(queries, resolve))
}

type adminFinanceExceptionResponse struct {
	FactRef      string `json:"factRef"`
	ProviderRef  string `json:"providerRef"`
	StatementRef string `json:"statementRef"`
	Currency     string `json:"currency"`
	Minor        int64  `json:"minor"`
	Exception    string `json:"exception"`
	OccurredAt   string `json:"occurredAt"`
	RecordedAt   string `json:"recordedAt"`
}

type adminFinanceCheckpointResponse struct {
	Day         string `json:"day"`
	Total       int    `json:"total"`
	Reconciled  int    `json:"reconciled"`
	Excepted    int    `json:"excepted"`
	CompletedAt string `json:"completedAt"`
}

func boundedFinanceRef(prefix, value string) string {
	if len(value) > 12 {
		value = value[len(value)-12:]
	}
	return prefix + "···" + value
}

func adminFinanceOverviewHandler(queries AdminFinanceOverview, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := resolveAdminPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !principal.Has(admin.ScopeFinance) {
			writeAdminVerificationError(w, r, admin.ErrForbidden)
			return
		}
		overview, err := queries.Overview(r.Context(), 50)
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "finance_reconciliation_unavailable", Message: "Reconciliation evidence is temporarily unavailable.",
			})
			return
		}
		exceptions := make([]adminFinanceExceptionResponse, 0, len(overview.Exceptions))
		for _, item := range overview.Exceptions {
			exceptions = append(exceptions, adminFinanceExceptionResponse{
				FactRef:      boundedFinanceRef("fact", item.FactID),
				ProviderRef:  boundedFinanceRef("provider", item.ProviderKey),
				StatementRef: boundedFinanceRef("statement", item.ReferenceKey),
				Currency:     string(item.Currency), Minor: item.Minor, Exception: string(item.Exception),
				OccurredAt: item.OccurredAt.UTC().Format(time.RFC3339),
				RecordedAt: item.RecordedAt.UTC().Format(time.RFC3339),
			})
		}
		checkpoints := make([]adminFinanceCheckpointResponse, 0, len(overview.Checkpoints))
		for _, item := range overview.Checkpoints {
			checkpoints = append(checkpoints, adminFinanceCheckpointResponse{
				Day: item.Day(), Total: item.Total(), Reconciled: item.Reconciled(),
				Excepted: item.Excepted(), CompletedAt: item.CompletedAt().UTC().Format(time.RFC3339),
			})
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{
			"exceptions": exceptions, "checkpoints": checkpoints,
			"exceptionCodes": []reconciliationdomain.ExceptionCode{
				reconciliationdomain.ExceptionLedgerMissing, reconciliationdomain.ExceptionReference,
				reconciliationdomain.ExceptionCurrency, reconciliationdomain.ExceptionAmount,
				reconciliationdomain.ExceptionUnbalancedLedger,
			},
		})
	})
}
