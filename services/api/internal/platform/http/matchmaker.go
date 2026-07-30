package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
)

type MatchmakerEngagements interface {
	Marketplace(context.Context) ([]domain.LicensedProfile, error)
	ForMember(context.Context, string) ([]domain.Engagement, error)
	FindForMember(context.Context, string, string) (domain.Engagement, error)
	Book(context.Context, string, string, domain.Terms, string) (domain.Engagement, error)
	Mutate(context.Context, string, string, func(domain.Engagement) (domain.Engagement, error)) (domain.Engagement, error)
}

type MatchmakerMemberKeyer interface {
	MemberKey(string) (string, error)
}

func RegisterMatchmakerRoutes(mux *http.ServeMux, engagements MatchmakerEngagements, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) {
	mux.Handle("GET /v1/matchmakers", listMatchmakersHandler(engagements, sessions))
	mux.Handle("GET /v1/matchmaker-engagements", listMatchmakerEngagementsHandler(engagements, keyer, sessions))
	mux.Handle("POST /v1/matchmaker-engagements", bookMatchmakerHandler(engagements, keyer, sessions))
	mux.Handle("GET /v1/matchmaker-engagements/{id}", getMatchmakerEngagementHandler(engagements, keyer, sessions))
	mux.Handle("POST /v1/matchmaker-engagements/{id}/member-consent", consentMatchmakerProposalHandler(engagements, keyer, sessions))
}

type matchmakerProfileResponse struct {
	MatchmakerID         string   `json:"matchmakerId"`
	DisplayName          string   `json:"displayName"`
	LicenseID            string   `json:"licenseId"`
	Jurisdiction         string   `json:"jurisdiction"`
	LicenseVersion       uint64   `json:"licenseVersion"`
	LicenseValidUntil    string   `json:"licenseValidUntil"`
	MinimumFeePesewas    uint64   `json:"minimumFeePesewas"`
	MaximumFeePesewas    uint64   `json:"maximumFeePesewas"`
	Languages            []string `json:"languages"`
	Specialties          []string `json:"specialties"`
	CompletedEngagements uint64   `json:"completedEngagements"`
	RatingBasisPoints    uint16   `json:"ratingBasisPoints"`
}

func listMatchmakersHandler(engagements MatchmakerEngagements, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := subanSubject(w, r, sessions); !ok {
			return
		}
		profiles, err := engagements.Marketplace(r.Context())
		if err != nil {
			writeMatchmakerError(w, r, err)
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

type engagementResponse struct {
	EngagementID       string `json:"engagementId"`
	MatchmakerID       string `json:"matchmakerId"`
	TermsID            string `json:"termsId"`
	TermsVersion       uint64 `json:"termsVersion"`
	TotalFeePesewas    uint64 `json:"totalFeePesewas"`
	BookedAt           string `json:"bookedAt"`
	MemberConsented    bool   `json:"memberConsented"`
	CandidateConsented bool   `json:"candidateConsented"`
	ProposalExposed    bool   `json:"proposalExposed"`
	Completed          bool   `json:"completed"`
	ProposalRef        string `json:"proposalRef,omitempty"`
	Revision           uint64 `json:"revision"`
}

func engagementView(engagement domain.Engagement) engagementResponse {
	state := engagement.State()
	response := engagementResponse{
		EngagementID: state.ID, MatchmakerID: state.MatchmakerKey,
		TermsID: state.Terms.ID, TermsVersion: state.Terms.Version,
		TotalFeePesewas: state.Terms.TotalFeePesewas, BookedAt: state.BookedAt.UTC().Format(time.RFC3339),
		MemberConsented: state.MemberConsented, CandidateConsented: state.CandidateConsented,
		ProposalExposed: state.Exposed, Completed: state.Completed, Revision: state.Revision,
	}
	if proposal, ok := engagement.ProposalRef(); ok {
		response.ProposalRef = proposal
	}
	return response
}

func matchmakerMember(w http.ResponseWriter, r *http.Request, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) (string, bool) {
	memberID, ok := subanSubject(w, r, sessions)
	if !ok {
		return "", false
	}
	if keyer == nil {
		writeMatchmakerError(w, r, application.ErrUnavailable)
		return "", false
	}
	key, err := keyer.MemberKey(memberID)
	if err != nil {
		writeMatchmakerError(w, r, application.ErrUnavailable)
		return "", false
	}
	return key, true
}

func listMatchmakerEngagementsHandler(engagements MatchmakerEngagements, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberKey, ok := matchmakerMember(w, r, keyer, sessions)
		if !ok {
			return
		}
		found, err := engagements.ForMember(r.Context(), memberKey)
		if err != nil {
			writeMatchmakerError(w, r, err)
			return
		}
		items := make([]engagementResponse, 0, len(found))
		for _, engagement := range found {
			items = append(items, engagementView(engagement))
		}
		writeSuccess(w, r, http.StatusOK, map[string]any{"items": items})
	})
}

type bookMatchmakerRequest struct {
	MatchmakerID string `json:"matchmakerId"`
	TermsID      string `json:"termsId"`
	TermsVersion uint64 `json:"termsVersion"`
	Milestones   []struct {
		ID           string `json:"id"`
		FeePesewas   uint64 `json:"feePesewas"`
		DueAfterDays uint16 `json:"dueAfterDays"`
	} `json:"milestones"`
}

func bookMatchmakerHandler(engagements MatchmakerEngagements, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberKey, ok := matchmakerMember(w, r, keyer, sessions)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		command := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		var body bookMatchmakerRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
			return
		}
		milestones := make([]domain.Milestone, 0, len(body.Milestones))
		var total uint64
		for _, milestone := range body.Milestones {
			if ^uint64(0)-total < milestone.FeePesewas {
				writeMatchmakerError(w, r, application.ErrInvalid)
				return
			}
			total += milestone.FeePesewas
			milestones = append(milestones, domain.Milestone{
				ID: milestone.ID, FeePesewas: milestone.FeePesewas,
				DueAfter: time.Duration(milestone.DueAfterDays) * 24 * time.Hour,
			})
		}
		engagement, err := engagements.Book(r.Context(), memberKey, body.MatchmakerID, domain.Terms{
			ID: body.TermsID, Version: body.TermsVersion,
			TotalFeePesewas: total, Milestones: milestones,
		}, command)
		if err != nil {
			writeMatchmakerError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, engagementView(engagement))
	})
}

func getMatchmakerEngagementHandler(engagements MatchmakerEngagements, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberKey, ok := matchmakerMember(w, r, keyer, sessions)
		if !ok {
			return
		}
		engagement, err := engagements.FindForMember(r.Context(), r.PathValue("id"), memberKey)
		if err != nil {
			writeMatchmakerError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, engagementView(engagement))
	})
}

func consentMatchmakerProposalHandler(engagements MatchmakerEngagements, keyer MatchmakerMemberKeyer, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberKey, ok := matchmakerMember(w, r, keyer, sessions)
		if !ok {
			return
		}
		if !adminJSONGuard(w, r) {
			return
		}
		command := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		current, err := engagements.FindForMember(r.Context(), r.PathValue("id"), memberKey)
		if err != nil {
			writeMatchmakerError(w, r, err)
			return
		}
		next, err := engagements.Mutate(r.Context(), current.ID(), command, func(value domain.Engagement) (domain.Engagement, error) {
			if value.State().MemberKey != memberKey {
				return domain.Engagement{}, application.ErrNotFound
			}
			return value.Consent(domain.ConsentMember, command, time.Now().UTC())
		})
		if errors.Is(err, application.ErrApplied) {
			next, err = engagements.FindForMember(r.Context(), current.ID(), memberKey)
		}
		if err != nil {
			writeMatchmakerError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, engagementView(next))
	})
}

func writeMatchmakerError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound):
		writeError(w, r, http.StatusNotFound, APIError{Code: "matchmaker_engagement_not_found", Message: "The matchmaker engagement was not found."})
	case errors.Is(err, application.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "matchmaker_request_invalid", Message: "The matchmaker request is not valid for the current terms."})
	case errors.Is(err, application.ErrConflict):
		writeError(w, r, http.StatusConflict, APIError{Code: "matchmaker_conflict", Message: "The engagement changed. Refresh and try again."})
	case errors.Is(err, application.ErrUnavailable):
		writeError(w, r, http.StatusServiceUnavailable, APIError{Code: "matchmaker_unavailable", Message: "Licensed matchmaker services are temporarily unavailable."})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The matchmaker request could not be completed."})
	}
}
