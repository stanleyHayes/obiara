package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	consentdomain "github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/application"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

type Profiles interface {
	View(context.Context, string, domain.Audience) (application.View, error)
	Upsert(context.Context, application.UpsertCommand) (application.UpsertResult, error)
}

type ProfileConsent interface {
	StateFor(context.Context, string, consentdomain.Purpose) (bool, error)
	Set(context.Context, string, consentdomain.Purpose, bool) (bool, error)
}

func RegisterProfileRoutes(
	mux *http.ServeMux,
	profiles Profiles,
	consents ProfileConsent,
	sessions SessionAuthenticator,
) {
	mux.Handle("GET /v1/profile", getOwnProfileHandler(profiles, sessions))
	mux.Handle("PUT /v1/profile", putOwnProfileHandler(profiles, consents, sessions))
}

type profileResponse struct {
	MemberID               string  `json:"memberId"`
	DisplayName            *string `json:"displayName"`
	Introduction           *string `json:"introduction"`
	DisplayNameVisibility  string  `json:"displayNameVisibility"`
	IntroductionVisibility string  `json:"introductionVisibility"`
	Revision               uint64  `json:"revision"`
	UpdatedAt              string  `json:"updatedAt"`
	Replayed               bool    `json:"replayed"`
}

func ownProfileResponse(view application.View, replayed bool) profileResponse {
	return profileResponse{
		MemberID: view.MemberID, DisplayName: view.DisplayName,
		Introduction:           view.Introduction,
		DisplayNameVisibility:  string(view.DisplayNameVisibility),
		IntroductionVisibility: string(view.IntroductionVisibility),
		Revision:               view.Revision, UpdatedAt: view.UpdatedAt.Format(time.RFC3339),
		Replayed: replayed,
	}
}

func getOwnProfileHandler(profiles Profiles, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		view, err := profiles.View(r.Context(), memberID, domain.AudienceSelf)
		if errors.Is(err, application.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, APIError{Code: "profile_not_found", Message: "Create your profile to continue."})
			return
		}
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "profile_unavailable", Message: "Your profile is temporarily unavailable."})
			return
		}
		writeSuccess(w, r, http.StatusOK, ownProfileResponse(view, false))
	})
}

type profileRequest struct {
	DisplayName            string `json:"displayName"`
	Introduction           string `json:"introduction"`
	DisplayNameVisibility  string `json:"displayNameVisibility"`
	IntroductionVisibility string `json:"introductionVisibility"`
	ExpectedRevision       uint64 `json:"expectedRevision"`
}

func putOwnProfileHandler(profiles Profiles, consents ProfileConsent, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{Code: "unsupported_media_type", Message: "Content-Type must be application/json."})
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		var body profileRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		nameVisibility := domain.Visibility(body.DisplayNameVisibility)
		introductionVisibility := domain.Visibility(body.IntroductionVisibility)
		if !validIdentifier(commandID) || !validProfileVisibility(nameVisibility) || !validProfileVisibility(introductionVisibility) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Profile values, visibility and Idempotency-Key must be valid."})
			return
		}
		community := nameVisibility == domain.VisibilityCommunity ||
			introductionVisibility == domain.VisibilityCommunity
		if consents == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "consent_unavailable", Message: "Profile visibility consent is unavailable."})
			return
		}
		current, err := consents.StateFor(r.Context(), memberID, consentdomain.PurposeProfileVisibility)
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "consent_unavailable", Message: "Profile visibility consent is unavailable."})
			return
		}
		if current != community {
			if _, err := consents.Set(r.Context(), memberID, consentdomain.PurposeProfileVisibility, community); err != nil {
				writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "consent_unavailable", Message: "Profile visibility consent could not be recorded."})
				return
			}
		}
		consentRef := func(visibility domain.Visibility) string {
			if visibility == domain.VisibilityCommunity {
				return "cons_profile_visibility"
			}
			return ""
		}
		result, err := profiles.Upsert(r.Context(), application.UpsertCommand{
			CommandID: commandID, MemberID: memberID, ExpectedRevision: body.ExpectedRevision,
			DisplayName: application.FieldInput{
				Value: body.DisplayName, Visibility: nameVisibility,
				ConsentRef: consentRef(nameVisibility),
			},
			Introduction: application.FieldInput{
				Value: body.Introduction, Visibility: introductionVisibility,
				ConsentRef: consentRef(introductionVisibility),
			},
		})
		if errors.Is(err, domain.ErrStaleRevision) || errors.Is(err, domain.ErrCommandMismatch) {
			writeError(w, r, http.StatusConflict, APIError{Code: "profile_conflict", Message: "Your profile changed in another session. Reload and try again."})
			return
		}
		if errors.Is(err, domain.ErrInvalidProfile) || errors.Is(err, domain.ErrUnsafeProfile) ||
			errors.Is(err, domain.ErrConsentRequired) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Profile fields are invalid or contain contact details."})
			return
		}
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "profile_unavailable", Message: "Your profile could not be saved."})
			return
		}
		view, err := profiles.View(r.Context(), memberID, domain.AudienceSelf)
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "profile_unavailable", Message: "Your profile was saved but could not be reloaded."})
			return
		}
		status := http.StatusOK
		if body.ExpectedRevision == 0 && !result.Replayed {
			status = http.StatusCreated
		}
		writeSuccess(w, r, status, ownProfileResponse(view, result.Replayed))
	})
}

func validProfileVisibility(value domain.Visibility) bool {
	return value == domain.VisibilityPrivate || value == domain.VisibilityCircles ||
		value == domain.VisibilityCommunity
}
