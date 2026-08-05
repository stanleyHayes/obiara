package apihttp

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/waitlist"
)

const waitlistConsentVersion = "launch-availability-v1"

type WaitlistStore interface {
	Join(context.Context, string, string, string) (waitlist.Entry, bool, error)
	List(context.Context, int) ([]waitlist.Entry, error)
}

func RegisterWaitlistRoutes(mux *http.ServeMux, store WaitlistStore, resolve AdminPrincipalResolver) {
	mux.Handle("POST /v1/waitlist", joinWaitlistHandler(store))
	mux.Handle("GET /v1/admin/waitlist", adminWaitlistHandler(store, resolve))
}

type joinWaitlistRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Consent bool   `json:"consent"`
}

func joinWaitlistHandler(store WaitlistStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request joinWaitlistRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		request.Name = strings.TrimSpace(request.Name)
		request.Email = strings.ToLower(strings.TrimSpace(request.Email))
		address, emailErr := mail.ParseAddress(request.Email)
		if request.Name == "" || len(request.Name) > 100 || emailErr != nil || address.Address != request.Email || len(request.Email) > 254 || !request.Consent {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "Enter your name and a valid email, then agree to receive the launch email."})
			return
		}
		entry, created, err := store.Join(r.Context(), request.Name, request.Email, waitlistConsentVersion)
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "waitlist_unavailable", Message: "The waitlist is temporarily unavailable. Please try again."})
			return
		}
		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}
		writeSuccess(w, r, status, map[string]any{"email": entry.Email, "alreadyJoined": !created})
	})
}

func adminWaitlistHandler(store WaitlistStore, resolve AdminPrincipalResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := resolveAdminPrincipal(w, r, resolve)
		if !ok {
			return
		}
		if !principal.Has(application.ScopeOperations) {
			writeAdminVerificationError(w, r, application.ErrForbidden)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		entries, err := store.List(r.Context(), limit)
		if err != nil && !errors.Is(err, context.Canceled) {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "waitlist_unavailable", Message: "The waitlist could not be loaded."})
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"entries": entries})
	})
}
