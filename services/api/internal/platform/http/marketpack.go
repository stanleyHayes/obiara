package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/application"
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

// MarketPacks is the inbound port for pack governance (E16-S06).
type MarketPacks interface {
	Draft(ctx context.Context, market domain.Market, terminologyRef string, features map[string]bool, proposerID string) (domain.MarketPack, error)
	Publish(ctx context.Context, packID, approverID string) (domain.MarketPack, error)
	Retire(ctx context.Context, packID, actorID string) (domain.MarketPack, error)
	All(ctx context.Context, limit int) ([]domain.MarketPack, error)
	Published(ctx context.Context) ([]domain.MarketPack, error)
}

// RegisterMarketPackRoutes adds the market-pack routes.
func RegisterMarketPackRoutes(mux *http.ServeMux, packs MarketPacks, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/market-packs", listAdminPacksHandler(packs, resolve))
	mux.Handle("POST /v1/admin/market-packs", draftPackHandler(packs, resolve))
	mux.Handle("POST /v1/admin/market-packs/{id}/publish", publishPackHandler(packs, resolve))
	mux.Handle("POST /v1/admin/market-packs/{id}/retire", retirePackHandler(packs, resolve))
	mux.Handle("GET /v1/market-packs/published", listPublishedHandler(packs))
}

func packJSONGuard(w http.ResponseWriter, r *http.Request) bool {
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
		writeError(w, r, http.StatusUnsupportedMediaType, APIError{
			Code:    "unsupported_media_type",
			Message: "Content-Type must be application/json.",
		})
		return false
	}
	return true
}

type packResponse struct {
	PackID         string          `json:"packId"`
	Market         string          `json:"market"`
	TerminologyRef string          `json:"terminologyRef"`
	Features       map[string]bool `json:"features"`
	Status         string          `json:"status"`
	Version        int64           `json:"version"`
	CreatedAt      time.Time       `json:"createdAt"`
	PublishedAt    *time.Time      `json:"publishedAt,omitempty"`
	ProposedByMe   bool            `json:"proposedByMe,omitempty"`
	ApprovedByMe   bool            `json:"approvedByMe,omitempty"`
}

func toPackResponse(pack domain.MarketPack, actorID string) packResponse {
	return packResponse{
		PackID: pack.ID(), Market: string(pack.Market()), TerminologyRef: pack.TerminologyRef(),
		Features: pack.Features(), Status: string(pack.Status()), Version: pack.Version(),
		CreatedAt: pack.CreatedAt(), PublishedAt: pack.PublishedAt(),
		ProposedByMe: actorID != "" && pack.ProposedBy() == actorID,
		ApprovedByMe: actorID != "" && pack.ApprovedBy() == actorID,
	}
}

func listAdminPacksHandler(packs MarketPacks, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		found, err := packs.All(r.Context(), 100)
		if err != nil {
			writePackError(w, r, err)
			return
		}
		response := make([]packResponse, 0, len(found))
		for _, pack := range found {
			response = append(response, toPackResponse(pack, principal.ActorID))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Packs []packResponse `json:"packs"`
		}{Packs: response})
	})
}

type draftPackRequest struct {
	Market         string          `json:"market"`
	TerminologyRef string          `json:"terminologyRef"`
	Features       map[string]bool `json:"features"`
}

func draftPackHandler(packs MarketPacks, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !requireMarketPackStepUp(w, r, principal) {
			return
		}
		if !packJSONGuard(w, r) {
			return
		}
		var body draftPackRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if strings.TrimSpace(body.TerminologyRef) == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "terminologyRef", Reason: "is required"}},
			})
			return
		}
		pack, err := packs.Draft(r.Context(), domain.Market(body.Market), body.TerminologyRef, body.Features, principal.ActorID)
		if err != nil {
			writePackError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toPackResponse(pack, principal.ActorID))
	})
}

func publishPackHandler(packs MarketPacks, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !requireMarketPackStepUp(w, r, principal) {
			return
		}
		if !packJSONGuard(w, r) {
			return
		}
		var body struct{}
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		pack, err := packs.Publish(r.Context(), r.PathValue("id"), principal.ActorID)
		if err != nil {
			writePackError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPackResponse(pack, principal.ActorID))
	})
}

func retirePackHandler(packs MarketPacks, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := operationsPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !requireMarketPackStepUp(w, r, principal) {
			return
		}
		if !packJSONGuard(w, r) {
			return
		}
		var body struct{}
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		pack, err := packs.Retire(r.Context(), r.PathValue("id"), principal.ActorID)
		if err != nil {
			writePackError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPackResponse(pack, principal.ActorID))
	})
}

func operationsPrincipal(w http.ResponseWriter, r *http.Request, resolve AdminPrincipalResolver) (admin.Principal, bool) {
	principal, ok := resolveAdminPrincipal(w, r, resolve)
	if !ok {
		return admin.Principal{}, false
	}
	if !principal.Has(adminOperationsScope) {
		writeAdminVerificationError(w, r, errAdminOperationsForbidden)
		return admin.Principal{}, false
	}
	return principal, true
}

func requireMarketPackStepUp(w http.ResponseWriter, r *http.Request, principal admin.Principal) bool {
	if principal.MFAVerified {
		return true
	}
	writeError(w, r, http.StatusForbidden, APIError{
		Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before changing market configuration.",
	})
	return false
}

func listPublishedHandler(packs MarketPacks) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		market := strings.TrimSpace(r.URL.Query().Get("market"))
		if market != "" {
			switch domain.Market(market) {
			case domain.MarketGhanaEnglish, domain.MarketGhanaTwi, domain.MarketGhanaPidgin, domain.MarketGhanaGa:
			default:
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code:    "validation_failed",
					Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "market", Reason: "must be gh_en, gh_tw, gh_pidgin or gh_ga"}},
				})
				return
			}
		}
		published, err := packs.Published(r.Context())
		if err != nil {
			writePackError(w, r, err)
			return
		}
		response := make([]packResponse, 0, len(published))
		for _, pack := range published {
			if market != "" && string(pack.Market()) != market {
				continue
			}
			response = append(response, toPackResponse(pack, ""))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Packs []packResponse `json:"packs"`
		}{Packs: response})
	})
}

func writePackError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrSelfApproval):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "self_approval",
			Message: "Publishing requires a different approver than the proposer.",
		})
	case errors.Is(err, domain.ErrPackNotDraft), errors.Is(err, domain.ErrPackNotPublished):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "pack_state",
			Message: "The pack is not in the required state for this action.",
		})
	case errors.Is(err, application.ErrPackNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "pack_not_found",
			Message: "No such market pack.",
		})
	case errors.Is(err, domain.ErrInvalidMarket), errors.Is(err, domain.ErrTerminologyRequired), errors.Is(err, domain.ErrActorRequired):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "One or more fields are invalid.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
