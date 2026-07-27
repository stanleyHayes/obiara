package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/application"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/domain"
)

// Nominations is the inbound port for Nnoboa kin nominations (E13-S06).
type Nominations interface {
	Nominate(ctx context.Context, in application.NominateInput) (domain.Nomination, error)
	ListForMember(ctx context.Context, memberID string) ([]domain.Nomination, error)
	Consent(ctx context.Context, id string) (domain.Nomination, error)
	Decline(ctx context.Context, id string) (domain.Nomination, error)
}

// RegisterNominationRoutes adds Nnoboa nomination routes.
func RegisterNominationRoutes(mux *http.ServeMux, nominations Nominations) {
	mux.Handle("POST /v1/nominations", nominateHandler(nominations))
	mux.Handle("GET /v1/nominations", listNominationsHandler(nominations))
	mux.Handle("POST /v1/nominations/{id}/consent", consentNominationHandler(nominations))
	mux.Handle("POST /v1/nominations/{id}/decline", declineNominationHandler(nominations))
}

type nominateRequest struct {
	MemberID     string `json:"memberId"`
	KinName      string `json:"kinName"`
	KinPhone     string `json:"kinPhone"`
	Relationship string `json:"relationship"`
}

type nominationResponse struct {
	ID           string     `json:"id"`
	MemberID     string     `json:"memberId"`
	KinName      string     `json:"kinName"`
	Relationship string     `json:"relationship"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	RespondedAt  *time.Time `json:"respondedAt,omitempty"`
}

type nominationListResponse struct {
	Nominations []nominationResponse `json:"nominations"`
}

func toNominationResponse(n domain.Nomination) nominationResponse {
	return nominationResponse{
		ID:           n.ID,
		MemberID:     n.MemberID,
		KinName:      n.KinName,
		Relationship: string(n.Relationship),
		Status:       string(n.Status),
		CreatedAt:    n.CreatedAt,
		RespondedAt:  n.RespondedAt,
	}
}

func nominateHandler(nominations Nominations) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body nominateRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.MemberID = strings.TrimSpace(body.MemberID)
		body.KinName = strings.TrimSpace(body.KinName)
		body.KinPhone = strings.TrimSpace(body.KinPhone)
		body.Relationship = strings.TrimSpace(body.Relationship)
		var details []FieldError
		if !validOpaqueID(body.MemberID) {
			details = append(details, FieldError{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if body.KinName == "" || len(body.KinName) > 120 {
			details = append(details, FieldError{Field: "kinName", Reason: "must be 1-120 characters"})
		}
		if !e164(body.KinPhone) {
			details = append(details, FieldError{Field: "kinPhone", Reason: "must be an E.164 phone number"})
		}
		switch domain.Relationship(body.Relationship) {
		case domain.Aunt, domain.Uncle, domain.Mother, domain.Father, domain.Elder:
		default:
			details = append(details, FieldError{Field: "relationship", Reason: "must be aunt, uncle, mother, father, or elder"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		n, err := nominations.Nominate(r.Context(), application.NominateInput{
			MemberID:     body.MemberID,
			KinName:      body.KinName,
			KinPhone:     body.KinPhone,
			Relationship: body.Relationship,
		})
		if err != nil {
			writeNominationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, toNominationResponse(n))
	})
}

func listNominationsHandler(nominations Nominations) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID := strings.TrimSpace(r.URL.Query().Get("memberId"))
		if !validOpaqueID(memberID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}

		ns, err := nominations.ListForMember(r.Context(), memberID)
		if err != nil {
			writeNominationError(w, r, err)
			return
		}
		out := make([]nominationResponse, 0, len(ns))
		for _, n := range ns {
			out = append(out, toNominationResponse(n))
		}
		writeSuccess(w, r, http.StatusOK, nominationListResponse{Nominations: out})
	})
}

func consentNominationHandler(nominations Nominations) http.Handler {
	return nominationTransitionHandler(nominations.Consent)
}

func declineNominationHandler(nominations Nominations) http.Handler {
	return nominationTransitionHandler(nominations.Decline)
}

func nominationTransitionHandler(transition func(context.Context, string) (domain.Nomination, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := transition(r.Context(), r.PathValue("id"))
		if err != nil {
			writeNominationError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, toNominationResponse(n))
	})
}

func writeNominationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNominationNotFound):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "nomination_not_found",
			Message: "No such nomination.",
		})
	case errors.Is(err, application.ErrDuplicateNomination):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "duplicate_nomination",
			Message: "A pending nomination already exists for this kin.",
		})
	case errors.Is(err, domain.ErrNotPending):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "nomination_not_pending",
			Message: "This nomination has already been answered.",
		})
	case errors.Is(err, domain.ErrInvalidNomination):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "One or more fields are invalid.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
