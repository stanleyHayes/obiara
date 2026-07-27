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
)

// MarketPacks is the inbound port for pack governance (E16-S06).
type MarketPacks interface {
	Draft(ctx context.Context, market domain.Market, terminologyRef string, features map[string]bool, proposerID string) (domain.MarketPack, error)
	Publish(ctx context.Context, packID, approverID string) (domain.MarketPack, error)
	Retire(ctx context.Context, packID, actorID string) (domain.MarketPack, error)
	Published(ctx context.Context) ([]domain.MarketPack, error)
}

// RegisterMarketPackRoutes adds the market-pack routes.
func RegisterMarketPackRoutes(mux *http.ServeMux, packs MarketPacks) {
	mux.Handle("POST /v1/admin/market-packs", draftPackHandler(packs))
	mux.Handle("POST /v1/admin/market-packs/{id}/publish", publishPackHandler(packs))
	mux.Handle("POST /v1/admin/market-packs/{id}/retire", retirePackHandler(packs))
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
	PackID     string    `json:"packId"`
	Market     string    `json:"market"`
	Status     string    `json:"status"`
	ProposedBy string    `json:"proposedBy"`
	ApprovedBy string    `json:"approvedBy,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

func toPackResponse(pack domain.MarketPack) packResponse {
	return packResponse{
		PackID: pack.ID(), Market: string(pack.Market()), Status: string(pack.Status()),
		ProposedBy: pack.ProposedBy(), ApprovedBy: pack.ApprovedBy(), CreatedAt: pack.CreatedAt(),
	}
}

type draftPackRequest struct {
	Market         string          `json:"market"`
	TerminologyRef string          `json:"terminologyRef"`
	Features       map[string]bool `json:"features"`
	ProposerID     string          `json:"proposerId"`
}

func draftPackHandler(packs MarketPacks) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !packJSONGuard(w, r) {
			return
		}
		var body draftPackRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if !validOpaqueID(body.ProposerID) || strings.TrimSpace(body.TerminologyRef) == "" {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "proposerId/terminologyRef", Reason: "proposerId must be an opaque id and terminologyRef is required"}},
			})
			return
		}
		pack, err := packs.Draft(r.Context(), domain.Market(body.Market), body.TerminologyRef, body.Features, body.ProposerID)
		if err != nil {
			writePackError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toPackResponse(pack))
	})
}

type actorRequest struct {
	ActorID string `json:"actorId"`
}

func publishPackHandler(packs MarketPacks) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !packJSONGuard(w, r) {
			return
		}
		var body actorRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if !validOpaqueID(body.ActorID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "actorId", Reason: "must be an opaque identifier"}},
			})
			return
		}
		pack, err := packs.Publish(r.Context(), r.PathValue("id"), body.ActorID)
		if err != nil {
			writePackError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPackResponse(pack))
	})
}

func retirePackHandler(packs MarketPacks) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !packJSONGuard(w, r) {
			return
		}
		var body actorRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		pack, err := packs.Retire(r.Context(), r.PathValue("id"), strings.TrimSpace(body.ActorID))
		if err != nil {
			writePackError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toPackResponse(pack))
	})
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
			response = append(response, toPackResponse(pack))
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
