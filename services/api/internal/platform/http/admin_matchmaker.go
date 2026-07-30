package apihttp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
)

type AdminMatchmakerCatalog interface {
	ListAll(context.Context) ([]domain.LicensedProfile, error)
	Put(context.Context, domain.LicensedProfile, uint64, string) error
}

func RegisterAdminMatchmakerRoutes(mux *http.ServeMux, catalog AdminMatchmakerCatalog, resolve AdminPrincipalResolver) {
	mux.Handle("GET /v1/admin/matchmakers", adminListMatchmakersHandler(catalog, resolve))
	mux.Handle("POST /v1/admin/matchmakers", adminPutMatchmakerHandler(catalog, resolve, true))
	mux.Handle("PUT /v1/admin/matchmakers/{id}", adminPutMatchmakerHandler(catalog, resolve, false))
}

func requireLicensingAdmin(w http.ResponseWriter, r *http.Request, resolve AdminPrincipalResolver, stepUp bool) (string, bool) {
	principal, ok := resolveAdminPrincipal(w, r, resolve)
	if !ok {
		return "", false
	}
	if !principal.Has(adminOperationsScope) {
		writeAdminVerificationError(w, r, errAdminOperationsForbidden)
		return "", false
	}
	if stepUp && !principal.MFAVerified {
		writeError(w, r, http.StatusForbidden, APIError{Code: "admin_step_up_required", Message: "Complete a fresh MFA step-up before changing a matchmaker licence."})
		return "", false
	}
	return principal.ActorID, true
}

func adminListMatchmakersHandler(catalog AdminMatchmakerCatalog, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireLicensingAdmin(w, r, resolve, false); !ok {
			return
		}
		profiles, err := catalog.ListAll(r.Context())
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The licensing register could not be loaded."})
			return
		}
		items := make([]matchmakerProfileResponse, 0, len(profiles))
		for _, profile := range profiles {
			items = append(items, matchmakerProfileResponse{
				MatchmakerID: profile.License.MatchmakerKey, DisplayName: profile.DisplayName,
				LicenseID: profile.License.ID, Jurisdiction: profile.License.Jurisdiction,
				LicenseVersion:       profile.License.Version,
				LicenseValidUntil:    profile.License.ValidUntil.UTC().Format(time.RFC3339),
				MinimumFeePesewas:    profile.License.MinimumFeePesewas,
				MaximumFeePesewas:    profile.License.MaximumFeePesewas,
				Languages:            append([]string(nil), profile.Languages...),
				Specialties:          append([]string(nil), profile.Specialties...),
				CompletedEngagements: profile.CompletedEngagements,
				RatingBasisPoints:    profile.RatingBasisPoints,
			})
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"items": items})
	})
}

type adminMatchmakerInput struct {
	LicenseID            string   `json:"licenseId"`
	Jurisdiction         string   `json:"jurisdiction"`
	ExpectedVersion      uint64   `json:"expectedVersion"`
	ValidFrom            string   `json:"validFrom"`
	ValidUntil           string   `json:"validUntil"`
	MinimumFeePesewas    uint64   `json:"minimumFeePesewas"`
	MaximumFeePesewas    uint64   `json:"maximumFeePesewas"`
	DisplayName          string   `json:"displayName"`
	Languages            []string `json:"languages"`
	Specialties          []string `json:"specialties"`
	CompletedEngagements uint64   `json:"completedEngagements"`
	RatingBasisPoints    uint16   `json:"ratingBasisPoints"`
}

func adminPutMatchmakerHandler(catalog AdminMatchmakerCatalog, resolve AdminPrincipalResolver, create bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := requireLicensingAdmin(w, r, resolve, true)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		var body adminMatchmakerInput
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		if create && body.ExpectedVersion != 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "A new licence must start at version one."})
			return
		}
		validFrom, fromErr := time.Parse(time.RFC3339, strings.TrimSpace(body.ValidFrom))
		validUntil, untilErr := time.Parse(time.RFC3339, strings.TrimSpace(body.ValidUntil))
		if fromErr != nil || untilErr != nil {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Licence validity must use RFC 3339 timestamps."})
			return
		}
		matchmakerID := strings.TrimSpace(r.PathValue("id"))
		if create {
			matchmakerID = newOpaqueMatchmakerID()
		}
		profile := domain.LicensedProfile{
			License: domain.License{
				ID: strings.TrimSpace(body.LicenseID), MatchmakerKey: matchmakerID,
				Jurisdiction: strings.TrimSpace(body.Jurisdiction), Version: body.ExpectedVersion + 1,
				ValidFrom: validFrom.UTC(), ValidUntil: validUntil.UTC(),
				MinimumFeePesewas: body.MinimumFeePesewas, MaximumFeePesewas: body.MaximumFeePesewas,
			},
			DisplayName: strings.TrimSpace(body.DisplayName),
			Languages:   append([]string(nil), body.Languages...), Specialties: append([]string(nil), body.Specialties...),
			CompletedEngagements: body.CompletedEngagements, RatingBasisPoints: body.RatingBasisPoints,
		}
		if err := catalog.Put(r.Context(), profile, body.ExpectedVersion, actorID); err != nil {
			switch {
			case errors.Is(err, application.ErrInvalid):
				writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "The matchmaker profile or licence is invalid."})
			case errors.Is(err, application.ErrConflict):
				writeError(w, r, http.StatusConflict, APIError{Code: "matchmaker_license_conflict", Message: "The licence changed. Refresh and try again."})
			default:
				writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The licence could not be retained."})
			}
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeSuccess(w, r, status, matchmakerProfileResponse{
			MatchmakerID: matchmakerID, DisplayName: profile.DisplayName,
			LicenseID: profile.License.ID, Jurisdiction: profile.License.Jurisdiction,
			LicenseVersion:    profile.License.Version,
			LicenseValidUntil: profile.License.ValidUntil.Format(time.RFC3339),
			MinimumFeePesewas: profile.License.MinimumFeePesewas,
			MaximumFeePesewas: profile.License.MaximumFeePesewas,
			Languages:         profile.Languages, Specialties: profile.Specialties,
			CompletedEngagements: profile.CompletedEngagements, RatingBasisPoints: profile.RatingBasisPoints,
		})
	})
}

func newOpaqueMatchmakerID() string {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}
