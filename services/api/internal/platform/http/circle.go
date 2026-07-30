package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
)

type Circles interface {
	Create(context.Context, application.Command, domain.Type) (application.Result, error)
	Request(context.Context, application.Command) (application.Result, error)
	Approve(context.Context, application.Command, string) (application.Result, error)
	PromoteHost(context.Context, application.Command, string) (application.Result, error)
	Leave(context.Context, application.Command) (application.Result, error)
	Expel(context.Context, application.Command, string) (application.Result, error)
	SetVisibility(context.Context, application.Command, domain.Visibility) (application.Result, error)
	Get(context.Context, string, string) (domain.Circle, error)
	List(context.Context, string, string, int) ([]domain.Circle, error)
}

func RegisterCircleRoutes(mux *http.ServeMux, circles Circles, sessions SessionAuthenticator) {
	mux.Handle("GET /v1/circles", listCirclesHandler(circles, sessions))
	mux.Handle("POST /v1/circles", createCircleHandler(circles, sessions))
	mux.Handle("GET /v1/circles/{id}", getCircleHandler(circles, sessions))
	mux.Handle("POST /v1/circles/{id}/requests", mutateCircleHandler(circles, sessions, "request"))
	mux.Handle("POST /v1/circles/{id}/leave", mutateCircleHandler(circles, sessions, "leave"))
	mux.Handle("PUT /v1/circles/{id}/visibility", mutateCircleHandler(circles, sessions, "visibility"))
	mux.Handle("POST /v1/circles/{id}/members/{memberId}/approve", mutateCircleHandler(circles, sessions, "approve"))
	mux.Handle("POST /v1/circles/{id}/members/{memberId}/promote", mutateCircleHandler(circles, sessions, "promote"))
	mux.Handle("POST /v1/circles/{id}/members/{memberId}/expel", mutateCircleHandler(circles, sessions, "expel"))
}

type circleResponse struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Visibility  string                 `json:"visibility"`
	Membership  string                 `json:"membership"`
	MemberCount int                    `json:"memberCount"`
	Revision    uint64                 `json:"revision"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Members     []circleMemberResponse `json:"members,omitempty"`
}

type circleMemberResponse struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

func projectCircle(circle domain.Circle, actorID string) circleResponse {
	membership := "none"
	memberCount := 0
	memberships := circle.Memberships()
	for _, item := range memberships {
		switch item.State() {
		case domain.StateMember, domain.StateHost, domain.StateOwner:
			memberCount++
		}
		if item.MemberID() == actorID {
			membership = string(item.State())
		}
	}
	response := circleResponse{
		ID: circle.ID(), Type: string(circle.Type()), Visibility: string(circle.Visibility()),
		Membership: membership, MemberCount: memberCount, Revision: circle.Revision(), UpdatedAt: circle.UpdatedAt(),
	}
	if membership == string(domain.StateOwner) || membership == string(domain.StateHost) {
		for _, item := range memberships {
			switch item.State() {
			case domain.StateRequested, domain.StateMember, domain.StateHost, domain.StateOwner:
				response.Members = append(response.Members, circleMemberResponse{ID: item.MemberID(), State: string(item.State())})
			}
		}
		sort.Slice(response.Members, func(i, j int) bool { return response.Members[i].ID < response.Members[j].ID })
	}
	return response
}

func listCirclesHandler(circles Circles, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		limit := 50
		if value := r.URL.Query().Get("limit"); value != "" {
			parsed, err := strconv.Atoi(value)
			if err != nil || parsed < 1 || parsed > 100 {
				writeError(w, r, http.StatusUnprocessableEntity, APIError{
					Code: "validation_failed", Message: "One or more fields are invalid.",
					Details: []FieldError{{Field: "limit", Reason: "must be 1-100"}},
				})
				return
			}
			limit = parsed
		}
		items, err := circles.List(r.Context(), memberID, r.URL.Query().Get("view"), limit)
		if err != nil {
			writeCircleError(w, r, err)
			return
		}
		projected := make([]circleResponse, 0, len(items))
		for _, circle := range items {
			projected = append(projected, projectCircle(circle, memberID))
		}
		writeSuccess(w, r, http.StatusOK, struct {
			Items []circleResponse `json:"items"`
		}{Items: projected})
	})
}

func getCircleHandler(circles Circles, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		circle, err := circles.Get(r.Context(), r.PathValue("id"), memberID)
		if err != nil {
			writeCircleError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectCircle(circle, memberID))
	})
}

type createCircleRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func createCircleHandler(circles Circles, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body createCircleRequest
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		body.ID, body.Type = strings.TrimSpace(body.ID), strings.TrimSpace(body.Type)
		if !validOpaqueID(body.ID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "id", Reason: "must be an opaque 1-128 character identifier"}},
			})
			return
		}
		result, err := circles.Create(r.Context(), application.Command{
			ID: commandID, CircleID: body.ID, ActorID: memberID,
		}, domain.Type(body.Type))
		if err != nil {
			writeCircleError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, projectCircle(result.Circle, memberID))
	})
}

type circleMutationRequest struct {
	ExpectedRevision uint64 `json:"expectedRevision"`
	Visibility       string `json:"visibility,omitempty"`
}

func mutateCircleHandler(circles Circles, sessions SessionAuthenticator, action string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := subanSubject(w, r, sessions)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body circleMutationRequest
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		command := application.Command{
			ID: commandID, CircleID: r.PathValue("id"), ActorID: actorID,
			ExpectedRevision: body.ExpectedRevision,
		}
		var (
			result application.Result
			err    error
		)
		switch action {
		case "request":
			result, err = circles.Request(r.Context(), command)
		case "leave":
			result, err = circles.Leave(r.Context(), command)
		case "visibility":
			result, err = circles.SetVisibility(r.Context(), command, domain.Visibility(body.Visibility))
		case "approve":
			result, err = circles.Approve(r.Context(), command, r.PathValue("memberId"))
		case "promote":
			result, err = circles.PromoteHost(r.Context(), command, r.PathValue("memberId"))
		case "expel":
			result, err = circles.Expel(r.Context(), command, r.PathValue("memberId"))
		}
		if err != nil {
			writeCircleError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, projectCircle(result.Circle, actorID))
	})
}

func circleCommandID(w http.ResponseWriter, r *http.Request) (string, bool) {
	commandID := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if !validOpaqueID(commandID) {
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "A valid Idempotency-Key is required.",
			Details: []FieldError{{Field: "Idempotency-Key", Reason: "must be an opaque 1-128 character identifier"}},
		})
		return "", false
	}
	return commandID, true
}

func decodeCircleJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
		writeError(w, r, http.StatusUnsupportedMediaType, APIError{Code: "unsupported_media_type", Message: "Content-Type must be application/json."})
		return false
	}
	if err := decodeJSON(w, r, target); err != nil {
		writeError(w, r, http.StatusBadRequest, APIError{Code: "invalid_json", Message: "The request body must be one valid JSON object."})
		return false
	}
	return true
}

func writeCircleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound), errors.Is(err, domain.ErrAccessDenied):
		writeError(w, r, http.StatusNotFound, APIError{Code: "circle_not_found", Message: "No such circle is available."})
	case errors.Is(err, domain.ErrStaleRevision):
		writeError(w, r, http.StatusConflict, APIError{Code: "circle_revision_conflict", Message: "The circle changed. Refresh and try again."})
	case errors.Is(err, domain.ErrInvalidTransition), errors.Is(err, domain.ErrCommandMismatch):
		writeError(w, r, http.StatusConflict, APIError{Code: "circle_transition_conflict", Message: "That membership change is no longer available."})
	case errors.Is(err, domain.ErrInvalidCircle):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: "The circle request is invalid."})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{Code: "internal_error", Message: "The request could not be completed."})
	}
}
