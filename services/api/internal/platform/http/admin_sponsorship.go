package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	sponsorshipapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/application"
	sponsorshipdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/domain"
)

// AdminSponsorships is the operator surface for organization-funded seats.
type AdminSponsorships interface {
	Deposit(context.Context, sponsorshipapp.DepositCommand) (sponsorshipdomain.Fund, error)
	FindByOrganization(ctx context.Context, organizationID string) (sponsorshipdomain.Fund, error)
	List(ctx context.Context, limit int) ([]sponsorshipdomain.Fund, error)
}

// RegisterAdminSponsorshipRoutes exposes an organization's funded balance.
//
// Deposits are recorded rather than collected: an organization pays by
// whatever means it and Obiara agreed — a bank transfer, an invoice settled
// offline — and somebody records what arrived (agent_plan.md §78). Step-up on
// the write, because recording money that did not arrive would let seats be
// given away.
func RegisterAdminSponsorshipRoutes(
	mux *http.ServeMux, funds AdminSponsorships, keyer OperatorKeyer,
	resolve AdminPrincipalResolver,
) {
	mux.Handle("GET /v1/admin/sponsorships", adminListFundsHandler(funds, resolve))
	mux.Handle("GET /v1/admin/organizations/{id}/sponsorship",
		adminFundHandler(funds, resolve))
	mux.Handle("POST /v1/admin/organizations/{id}/sponsorship/deposits",
		adminDepositHandler(funds, keyer, resolve))
}

type fundResponse struct {
	OrganizationID   string `json:"organizationId"`
	DepositedPesewas int64  `json:"depositedPesewas"`
	DrawnPesewas     int64  `json:"drawnPesewas"`
	BalancePesewas   int64  `json:"balancePesewas"`
	// Seats is how many have been taken. Which members took them is never
	// said: an organization is told a count.
	Seats    int    `json:"seats"`
	Closed   bool   `json:"closed"`
	OpenedAt string `json:"openedAt"`
}

func projectFund(fund sponsorshipdomain.Fund) fundResponse {
	return fundResponse{
		OrganizationID: fund.OrganizationID(), DepositedPesewas: fund.DepositedPesewas(),
		DrawnPesewas: fund.DrawnPesewas(), BalancePesewas: fund.Balance(),
		Seats: fund.Seats(), Closed: fund.Closed(),
		OpenedAt: fund.OpenedAt().UTC().Format(time.RFC3339),
	}
}

func adminListFundsHandler(
	funds AdminSponsorships, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, false); !ok {
			return
		}
		if funds == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		found, err := funds.List(r.Context(), limit)
		if err != nil {
			writeSponsorshipError(w, r, err)
			return
		}
		items := make([]fundResponse, 0, len(found))
		for _, fund := range found {
			items = append(items, projectFund(fund))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Sponsorships []fundResponse `json:"sponsorships"`
		}{items})
	})
}

func adminFundHandler(
	funds AdminSponsorships, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, false); !ok {
			return
		}
		if funds == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		fund, err := funds.FindByOrganization(r.Context(), r.PathValue("id"))
		if err != nil {
			writeSponsorshipError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectFund(fund))
	})
}

type depositInput struct {
	AmountPesewas int64  `json:"amountPesewas"`
	ReasonCode    string `json:"reasonCode"`
}

func adminDepositHandler(
	funds AdminSponsorships, keyer OperatorKeyer, resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireLicensingAdmin(w, r, resolve, true)
		if !ok || !adminJSONGuard(w, r) {
			return
		}
		commandID, ok := organizationCommandID(w, r)
		if !ok {
			return
		}
		var body depositInput
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if funds == nil || keyer == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		operatorKey, err := keyer.Key("sponsorship_operator", actorID)
		if err != nil {
			logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This is not available right now.",
			})
			return
		}
		fund, err := funds.Deposit(r.Context(), sponsorshipapp.DepositCommand{
			CommandID: commandID, OperatorKey: operatorKey,
			ReasonCode:     strings.TrimSpace(body.ReasonCode),
			OrganizationID: r.PathValue("id"), AmountPesewas: body.AmountPesewas,
		})
		if err != nil {
			writeSponsorshipError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectFund(fund))
	})
}

func writeSponsorshipError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, sponsorshipapp.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "sponsorship_not_found", Message: "That organization has no funded balance.",
		})
	case errors.Is(err, sponsorshipapp.ErrIssuerNotIssuing):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "organization_not_issuing",
			Message: "That organization is suspended, so nothing can be funded for it.",
		})
	case errors.Is(err, sponsorshipdomain.ErrAmountOutOfRange):
		// Almost certainly a typo. Saying which field and why beats a generic
		// refusal an operator would re-submit unchanged.
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
			Details: []FieldError{
				{Field: "amountPesewas", Reason: "is larger than any real deposit — check the number of zeros"},
			},
		})
	case errors.Is(err, sponsorshipdomain.ErrFundClosed):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "sponsorship_closed", Message: "That fund is closed.",
		})
	case errors.Is(err, sponsorshipdomain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, sponsorshipapp.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "sponsorship_conflict",
			Message: "That fund changed while you were working. Refresh and try again.",
		})
	case errors.Is(err, sponsorshipdomain.ErrInvalidFund):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	case errors.Is(err, sponsorshipapp.ErrUnavailable):
		logServerError(r.Context(), r, http.StatusServiceUnavailable, "feature_unavailable", err)
		writeError(w, r, http.StatusServiceUnavailable, APIError{
			Code: "feature_unavailable", Message: "This is not available right now.",
		})
	default:
		logServerError(r.Context(), r, http.StatusInternalServerError, "internal_error", err)
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code: "internal_error", Message: "The request could not be completed.",
		})
	}
}
