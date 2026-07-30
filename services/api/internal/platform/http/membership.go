package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
)

type Memberships interface {
	Current(context.Context, string) (domain.Pass, error)
	Cancel(context.Context, string, string) (domain.Pass, error)
	RequestRefund(context.Context, string, string) (domain.Pass, error)
}

type MembershipKeyer interface {
	MemberKey(string) (string, error)
}

func RegisterMembershipRoutes(
	mux *http.ServeMux,
	memberships Memberships,
	keyer MembershipKeyer,
	sessions SessionAuthenticator,
) {
	mux.Handle("GET /v1/membership", getMembershipHandler(memberships, keyer, sessions))
	mux.Handle("POST /v1/membership/cancel", mutateMembershipHandler(memberships, keyer, sessions, "cancel"))
	mux.Handle("POST /v1/membership/refunds", mutateMembershipHandler(memberships, keyer, sessions, "refund"))
}

type membershipResponse struct {
	PassID              string `json:"passId"`
	PassName            string `json:"passName"`
	Status              string `json:"status"`
	PaidThrough         string `json:"paidThrough"`
	GraceUntil          string `json:"graceUntil"`
	RenewsAutomatically bool   `json:"renewsAutomatically"`
	ReceiptRef          string `json:"receiptRef"`
	RefundRequestRef    string `json:"refundRequestRef,omitempty"`
	Revision            uint64 `json:"revision"`
}

func membershipView(pass domain.Pass, now time.Time) membershipResponse {
	state := pass.State()
	return membershipResponse{
		PassID: state.ID, PassName: state.PassID, Status: string(pass.Status(now)),
		PaidThrough:         state.PaidThrough.Format(time.RFC3339),
		GraceUntil:          state.GraceUntil.Format(time.RFC3339),
		RenewsAutomatically: state.CancelledAt.IsZero(),
		ReceiptRef:          state.ReceiptRef, RefundRequestRef: state.RefundRequestRef,
		Revision: state.Revision,
	}
}

func membershipMember(
	w http.ResponseWriter,
	r *http.Request,
	keyer MembershipKeyer,
	sessions SessionAuthenticator,
) (string, bool) {
	memberID, ok := subanSubject(w, r, sessions)
	if !ok {
		return "", false
	}
	if keyer == nil {
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "membership_unavailable", Message: "Membership is temporarily unavailable."})
		return "", false
	}
	key, err := keyer.MemberKey(memberID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "membership_unavailable", Message: "Membership is temporarily unavailable."})
		return "", false
	}
	return key, true
}

func getMembershipHandler(memberships Memberships, keyer MembershipKeyer, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberKey, ok := membershipMember(w, r, keyer, sessions)
		if !ok {
			return
		}
		pass, err := memberships.Current(r.Context(), memberKey)
		if errors.Is(err, application.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, APIError{Code: "membership_not_found", Message: "No membership pass is active for this account."})
			return
		}
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "membership_unavailable", Message: "Membership is temporarily unavailable."})
			return
		}
		writeSuccess(w, r, http.StatusOK, membershipView(pass, time.Now().UTC()))
	})
}

func mutateMembershipHandler(
	memberships Memberships,
	keyer MembershipKeyer,
	sessions SessionAuthenticator,
	action string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberKey, ok := membershipMember(w, r, keyer, sessions)
		if !ok {
			return
		}
		commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if !validIdentifier(commandID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "A valid Idempotency-Key is required."})
			return
		}
		current, err := memberships.Current(r.Context(), memberKey)
		if errors.Is(err, application.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, APIError{Code: "membership_not_found", Message: "No membership pass is active for this account."})
			return
		}
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "membership_unavailable", Message: "Membership is temporarily unavailable."})
			return
		}
		var next domain.Pass
		if action == "cancel" {
			next, err = memberships.Cancel(r.Context(), current.ID(), commandID)
		} else {
			next, err = memberships.RequestRefund(r.Context(), current.ID(), commandID)
		}
		if errors.Is(err, application.ErrApplied) {
			next, err = memberships.Current(r.Context(), memberKey)
		}
		if errors.Is(err, application.ErrInvalid) {
			writeError(w, r, http.StatusConflict, APIError{Code: "membership_transition_invalid", Message: "That membership action is not available in the current state."})
			return
		}
		if err != nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "membership_unavailable", Message: "The membership action could not be completed."})
			return
		}
		writeSuccess(w, r, http.StatusOK, membershipView(next, time.Now().UTC()))
	})
}
