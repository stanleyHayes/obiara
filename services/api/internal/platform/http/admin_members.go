package apihttp

import (
	"context"
	"net/http"
	"strconv"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

// AdminMemberAccounts is the narrow read port used by the operations
// directory. Enforcement intentionally stays in the case-bound safety
// action ladder and is not exposed here.
type AdminMemberAccounts interface {
	List(context.Context, int) ([]identitydomain.Account, error)
}

type AdminMemberKeyer interface {
	Key(namespace, value string) (string, error)
}

func RegisterAdminMemberRoutes(
	mux *http.ServeMux,
	accounts AdminMemberAccounts,
	keyer AdminMemberKeyer,
	resolve AdminPrincipalResolver,
) {
	mux.Handle("GET /v1/admin/members", adminMemberDirectoryHandler(accounts, keyer, resolve))
}

type adminMemberResponse struct {
	Ref            string  `json:"ref"`
	Tier           int     `json:"tier"`
	Status         string  `json:"status"`
	SuspendedUntil *string `json:"suspendedUntil,omitempty"`
	JoinedAt       string  `json:"joinedAt"`
}

func adminMemberDirectoryHandler(
	accounts AdminMemberAccounts,
	keyer AdminMemberKeyer,
	resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := resolveAdminPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !principal.Has(admin.ScopeOperations) && !principal.Has(admin.ScopeSafety) {
			writeAdminVerificationError(w, r, admin.ErrForbidden)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 200 {
			limit = 100
		}
		found, err := accounts.List(r.Context(), limit)
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "member_directory_unavailable", Message: "The member directory is temporarily unavailable.",
			})
			return
		}
		items := make([]adminMemberResponse, 0, len(found))
		for _, account := range found {
			ref, keyErr := keyer.Key("admin-member", account.ID())
			if keyErr != nil {
				writeError(w, r, http.StatusServiceUnavailable, APIError{
					Code: "member_directory_unavailable", Message: "The member directory is temporarily unavailable.",
				})
				return
			}
			item := adminMemberResponse{
				Ref: ref, Tier: int(account.Tier()), Status: string(account.Status()),
				JoinedAt: account.CreatedAt().UTC().Format(time.RFC3339),
			}
			if until := account.SuspendedUntil(); until != nil {
				formatted := until.UTC().Format(time.RFC3339)
				item.SuspendedUntil = &formatted
			}
			items = append(items, item)
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"members": items})
	})
}
