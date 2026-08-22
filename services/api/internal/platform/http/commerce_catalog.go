package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
)

// Catalog is the inbound port for the commerce catalog: what an operator
// curates and a member can read once it is published.
type Catalog interface {
	Create(context.Context, application.CreateCommand) (domain.SKU, error)
	Publish(context.Context, application.Change) (domain.SKU, error)
	Retire(context.Context, application.Change) (domain.SKU, error)
	ReadPublished(ctx context.Context, sku string, version uint64) (domain.SKU, error)
}

// RegisterCatalogRoutes adds the catalog routes.
//
// The write side is an operator surface authenticated by the admin session;
// the read side serves members and returns published SKUs only, so a draft
// or retired price can never be quoted to anybody.
func RegisterCatalogRoutes(mux *http.ServeMux, catalog Catalog, sessions SessionAuthenticator) {
	mux.Handle("POST /v1/admin/catalog/skus", createSKUHandler(catalog))
	mux.Handle("POST /v1/admin/catalog/skus/{id}/publish", changeSKUHandler(catalog, true))
	mux.Handle("POST /v1/admin/catalog/skus/{id}/retire", changeSKUHandler(catalog, false))
	mux.Handle("GET /v1/catalog/skus/{skuKey}", readSKUHandler(catalog, sessions))
}

// adminSessionID reads the operator's bearer session without resolving it.
// The catalog's authority bridge does the resolution, so the role check
// lives in one place rather than being repeated per handler.
func adminSessionID(w http.ResponseWriter, r *http.Request) (string, bool) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code: "admin_authentication_required", Message: "A valid admin session is required.",
		})
		return "", false
	}
	session := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if session == "" {
		writeError(w, r, http.StatusUnauthorized, APIError{
			Code: "admin_authentication_required", Message: "A valid admin session is required.",
		})
		return "", false
	}
	return session, true
}

type createSKURequest struct {
	CommandID   string `json:"commandId"`
	SKUKey      string `json:"skuKey"`
	Title       string `json:"title"`
	Kind        string `json:"kind"`
	Currency    string `json:"currency"`
	AmountMinor int64  `json:"amountMinor"`
}

type changeSKURequest struct {
	CommandID        string `json:"commandId"`
	Version          uint64 `json:"version"`
	ExpectedRevision uint64 `json:"expectedRevision"`
}

type skuResponse struct {
	SKUID       string `json:"skuId"`
	SKUKey      string `json:"skuKey"`
	Version     uint64 `json:"version"`
	Kind        string `json:"kind"`
	Currency    string `json:"currency"`
	AmountMinor int64  `json:"amountMinor"`
	Status      string `json:"status"`
	Revision    uint64 `json:"revision"`
}

func toSKUResponse(sku domain.SKU) skuResponse {
	return skuResponse{
		SKUID: sku.ID(), SKUKey: sku.SKUKey(), Version: sku.Version(),
		Kind: string(sku.Kind()), Currency: string(sku.Price().Currency),
		// Money is carried in minor units. A float here would eventually
		// quote a member a price that does not exist.
		AmountMinor: sku.Price().Minor,
		Status:      string(sku.Status()), Revision: sku.Revision(),
	}
}

func createSKUHandler(catalog Catalog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		// The operator's session id is the actor; the authority bridge
		// resolves it to a principal and its roles.
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		var body createSKURequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		body.SKUKey = strings.TrimSpace(body.SKUKey)
		body.Title = strings.TrimSpace(body.Title)

		var details []FieldError
		if !validOpaqueID(body.CommandID) {
			details = append(details, FieldError{Field: "commandId", Reason: "must be an opaque identifier"})
		}
		if !validOpaqueID(body.SKUKey) {
			details = append(details, FieldError{Field: "skuKey", Reason: "must be an opaque identifier"})
		}
		if body.Title == "" || len(body.Title) > 200 {
			details = append(details, FieldError{Field: "title", Reason: "must be 1-200 characters"})
		}
		kind := domain.Kind(strings.ToLower(strings.TrimSpace(body.Kind)))
		switch kind {
		case domain.KindPhysicalGood, domain.KindEventTicket, domain.KindDigitalService:
		default:
			details = append(details, FieldError{Field: "kind", Reason: "must be physical_good, event_ticket or digital_service"})
		}
		currency := domain.Currency(strings.ToUpper(strings.TrimSpace(body.Currency)))
		switch currency {
		case domain.CurrencyGHS, domain.CurrencyUSD:
		default:
			details = append(details, FieldError{Field: "currency", Reason: "must be GHS or USD"})
		}
		if body.AmountMinor < 0 {
			details = append(details, FieldError{Field: "amountMinor", Reason: "must not be negative"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.", Details: details,
			})
			return
		}

		sku, err := catalog.Create(r.Context(), application.CreateCommand{
			Actor: session, SKUKey: body.SKUKey, Title: body.Title, CommandID: body.CommandID,
			Kind: kind, Price: domain.Price{Currency: currency, Minor: body.AmountMinor},
		})
		if err != nil {
			writeCatalogError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toSKUResponse(sku))
	})
}

func changeSKUHandler(catalog Catalog, publish bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !adminJSONGuard(w, r) {
			return
		}
		session, ok := adminSessionID(w, r)
		if !ok {
			return
		}
		id := r.PathValue("id")
		if !validOpaqueID(id) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		var body changeSKURequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.CommandID = strings.TrimSpace(body.CommandID)
		if !validOpaqueID(body.CommandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "commandId", Reason: "must be an opaque identifier"}},
			})
			return
		}

		change := application.Change{
			Actor: session, ID: id, CommandID: body.CommandID,
			Version: body.Version, ExpectedRevision: body.ExpectedRevision,
		}
		var sku domain.SKU
		var err error
		if publish {
			sku, err = catalog.Publish(r.Context(), change)
		} else {
			sku, err = catalog.Retire(r.Context(), change)
		}
		if err != nil {
			writeCatalogError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSKUResponse(sku))
	})
}

// readSKUHandler serves members. Only published SKUs are returned, so a
// draft price or a retired one can never be quoted.
func readSKUHandler(catalog Catalog, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := authenticatedMember(w, r, sessions); !ok {
			return
		}
		skuKey := r.PathValue("skuKey")
		if !validOpaqueID(skuKey) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
			})
			return
		}
		version := uint64(0)
		if raw := strings.TrimSpace(r.URL.Query().Get("version")); raw != "" {
			parsed, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "version", Reason: "must be a version number"}},
				})
				return
			}
			version = parsed
		}

		sku, err := catalog.ReadPublished(r.Context(), skuKey, version)
		if err != nil {
			writeCatalogError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toSKUResponse(sku))
	})
}

// writeCatalogError maps the slice's failures. An operator who may not
// curate and a SKU that does not exist are told the same thing, so the
// catalog cannot be used to enumerate unpublished products.
func writeCatalogError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound), errors.Is(err, application.ErrUnavailable):
		writeError(w, r, http.StatusNotFound, APIError{
			Code: "sku_not_found", Message: "That item is not available.",
		})
	case errors.Is(err, application.ErrConflict), errors.Is(err, domain.ErrStale):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "catalog_conflict", Message: "That item changed. Refresh and try again.",
		})
	case errors.Is(err, domain.ErrMismatch):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "command_mismatch", Message: "That command id was already used for a different request.",
		})
	case errors.Is(err, application.ErrInvalid), errors.Is(err, domain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "One or more fields are invalid.",
		})
	default:
		// An authority refusal lands here. It answers 403 rather than 404
		// because the operator is authenticated: they know the console
		// exists, and telling them they lack the role is actionable.
		logServerError(r.Context(), r, http.StatusForbidden, "catalog_editor_required", err)
		writeError(w, r, http.StatusForbidden, APIError{
			Code: "catalog_editor_required", Message: "Editing the catalog requires the finance or admin role.",
		})
	}
}
