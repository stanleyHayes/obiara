package apihttp

import (
	"context"
	"net"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/waitlist"
)

const waitlistConsentVersion = "launch-availability-v1"

type WaitlistStore interface {
	Join(context.Context, string, string, string) (waitlist.Entry, bool, error)
	List(context.Context, int) ([]waitlist.Entry, error)
}

func RegisterWaitlistRoutes(mux *http.ServeMux, store WaitlistStore, resolve AdminPrincipalResolver) {
	mux.Handle("POST /v1/waitlist", joinWaitlistHandler(store, newWaitlistJoinThrottle()))
	mux.Handle("GET /v1/admin/waitlist", adminWaitlistHandler(store, resolve))
}

// waitlistJoinWindow and waitlistJoinLimit bound how often one client IP may
// hit the unauthenticated join endpoint, mirroring the per-phone resend
// throttling of the OTP handler (in-memory, per process).
const (
	waitlistJoinWindow = time.Hour
	waitlistJoinLimit  = 5
)

type waitlistJoinThrottle struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	now      func() time.Time
}

func newWaitlistJoinThrottle() *waitlistJoinThrottle {
	return &waitlistJoinThrottle{attempts: map[string][]time.Time{}, now: time.Now}
}

// allow reports whether the client IP may submit another join request.
func (throttle *waitlistJoinThrottle) allow(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	now := throttle.now().UTC()
	throttle.mu.Lock()
	defer throttle.mu.Unlock()
	recent := throttle.attempts[host][:0]
	for _, attempt := range throttle.attempts[host] {
		if now.Sub(attempt) < waitlistJoinWindow {
			recent = append(recent, attempt)
		}
	}
	if len(recent) >= waitlistJoinLimit {
		throttle.attempts[host] = recent
		return false
	}
	throttle.attempts[host] = append(recent, now)
	return true
}

type joinWaitlistRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Consent bool   `json:"consent"`
}

func joinWaitlistHandler(store WaitlistStore, throttle *waitlistJoinThrottle) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !throttle.allow(r.RemoteAddr) {
			writeError(w, r, http.StatusTooManyRequests, APIError{Code: "waitlist_rate_limited", Message: "Too many join attempts. Please wait and try again."})
			return
		}
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
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "waitlist_unavailable", Message: "The waitlist could not be loaded."})
			return
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"entries": entries})
	})
}
