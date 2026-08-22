package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

// Ledger is the inbound port for the double-entry commerce ledger.
type Ledger interface {
	Post(context.Context, application.PostCommand) (domain.Posting, error)
	Balance(context.Context, application.BalanceRequest) (int64, error)
}

// RegisterLedgerRoutes adds the ledger routes. Both are finance-desk
// surfaces: a ledger records what the platform owes and is owed, and no
// member-facing route reads or writes it.
func RegisterLedgerRoutes(mux *http.ServeMux, ledger Ledger) {
	mux.Handle("POST /v1/admin/ledger/postings", postLedgerHandler(ledger))
	mux.Handle("GET /v1/admin/ledger/balances/{accountId}", ledgerBalanceHandler(ledger))
}

type ledgerLineInput struct {
	AccountID string `json:"accountId"`
	Class     string `json:"class"`
	Side      string `json:"side"`
	Minor     int64  `json:"minor"`
}

type postLedgerRequest struct {
	CommandID   string            `json:"commandId"`
	ReferenceID string            `json:"referenceId"`
	Purpose     string            `json:"purpose"`
	Currency    string            `json:"currency"`
	Lines       []ledgerLineInput `json:"lines"`
}

type postingResponse struct {
	PostingID string `json:"postingId"`
	Purpose   string `json:"purpose"`
	Currency  string `json:"currency"`
}

type balanceResponse struct {
	AccountID string `json:"accountId"`
	Class     string `json:"class"`
	Currency  string `json:"currency"`
	Minor     int64  `json:"minor"`
}

func validLedgerClass(class domain.AccountClass) bool {
	switch class {
	case domain.ClassAsset, domain.ClassLiability, domain.ClassEquity,
		domain.ClassRevenue, domain.ClassExpense:
		return true
	}
	return false
}

func postLedgerHandler(ledger Ledger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		var body postLedgerRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.ReferenceID = strings.TrimSpace(body.ReferenceID)

		var details []FieldError
		if !validOpaqueID(body.CommandID) {
			details = append(details, FieldError{Field: "commandId", Reason: "must be an opaque identifier"})
		}
		if !validOpaqueID(body.ReferenceID) {
			details = append(details, FieldError{Field: "referenceId", Reason: "must be an opaque identifier"})
		}
		// A posting with fewer than two lines cannot balance, so it is
		// rejected here rather than reaching the domain as an unbalanced
		// entry.
		if len(body.Lines) < 2 || len(body.Lines) > 50 {
			details = append(details, FieldError{Field: "lines", Reason: "must be 2-50 lines"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.", Details: details,
			})
			return
		}

		lines := make([]application.Line, 0, len(body.Lines))
		for _, input := range body.Lines {
			class := domain.AccountClass(strings.ToLower(strings.TrimSpace(input.Class)))
			side := domain.Side(strings.ToLower(strings.TrimSpace(input.Side)))
			if !validLedgerClass(class) || (side != domain.SideDebit && side != domain.SideCredit) ||
				!validOpaqueID(strings.TrimSpace(input.AccountID)) || input.Minor <= 0 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "lines", Reason: "each line needs an account, a valid class and side, and a positive minor amount"}},
				})
				return
			}
			lines = append(lines, application.Line{
				AccountID: strings.TrimSpace(input.AccountID), Class: class,
				Side: side, Minor: input.Minor,
			})
		}

		posting, err := ledger.Post(r.Context(), application.PostCommand{
			Actor: session, CommandID: body.CommandID, ReferenceID: body.ReferenceID,
			Purpose:  domain.Purpose(strings.ToLower(strings.TrimSpace(body.Purpose))),
			Currency: domain.Currency(strings.ToUpper(strings.TrimSpace(body.Currency))),
			Lines:    lines,
		})
		if err != nil {
			writeLedgerError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, postingResponse{
			PostingID: posting.ID(), Purpose: string(posting.Purpose()),
			Currency: string(posting.Currency()),
		})
	})
}

func ledgerBalanceHandler(ledger Ledger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		accountID := r.PathValue("accountId")
		class := domain.AccountClass(strings.ToLower(strings.TrimSpace(r.URL.Query().Get("class"))))
		currency := domain.Currency(strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("currency"))))
		if !validOpaqueID(accountID) || !validLedgerClass(class) || currency == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "class", Reason: "class and currency are required"}},
			})
			return
		}

		minor, err := ledger.Balance(r.Context(), application.BalanceRequest{
			Actor: session, AccountID: accountID, Class: class, Currency: currency,
		})
		if err != nil {
			writeLedgerError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, balanceResponse{
			AccountID: accountID, Class: string(class),
			Currency: string(currency), Minor: minor,
		})
	})
}

// writeLedgerError maps the slice's failures. An unbalanced posting is a
// caller error worth naming precisely: it is the one mistake a double-entry
// ledger exists to refuse.
func writeLedgerError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUnbalanced):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "posting_unbalanced", Message: "Debits and credits must balance.",
		})
	case errors.Is(err, domain.ErrOverflow):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "amount_overflow", Message: "That amount is out of range.",
		})
	case errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "posting_not_found", Message: "That posting is not available.",
		})
	case errors.Is(err, application.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "ledger_conflict", Message: "That command id was already used for a different posting.",
		})
	case errors.Is(err, application.ErrInvalid), errors.Is(err, domain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	default:
		logServerError(r.Context(), r, http.StatusForbidden, "finance_role_required", err)
		writeError(w, r, http.StatusForbidden, APIError{
			Code: "finance_role_required", Message: "The ledger requires the finance or admin role.",
		})
	}
}
