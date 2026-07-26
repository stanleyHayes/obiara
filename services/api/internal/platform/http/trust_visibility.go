package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	trustdomain "github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/visibility"
)

type TrustVisibility interface {
	Explain(context.Context, visibility.Request) ([]visibility.Explanation, error)
}

type SessionAuthenticator interface {
	Authenticate(context.Context, string) (identitydomain.Session, error)
}

func RegisterTrustVisibilityRoutes(mux *http.ServeMux, service TrustVisibility, sessions SessionAuthenticator) {
	mux.Handle("GET /v1/members/{memberId}/trust-paths", trustVisibilityHandler(service, sessions))
}

type trustPathStepResponse struct {
	SourceID string `json:"sourceId"`
	TargetID string `json:"targetId"`
	Reason   string `json:"reason"`
}

type trustPathResponse struct {
	TargetID string                  `json:"targetId"`
	Hops     int                     `json:"hops"`
	Steps    []trustPathStepResponse `json:"steps"`
}

type trustPathsResponse struct {
	Paths []trustPathResponse `json:"paths"`
}

func trustVisibilityHandler(service TrustVisibility, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		memberID := strings.TrimSpace(request.PathValue("memberId"))
		token, ok := bearerToken(request.Header.Get("Authorization"))
		if !ok || service == nil || sessions == nil || !validIdentifier(memberID) {
			writeTrustNotFound(writer, request)
			return
		}
		session, err := sessions.Authenticate(request.Context(), token)
		if err != nil || session.MemberID() != memberID {
			writeTrustNotFound(writer, request)
			return
		}
		depth, ok := boundedQueryInt(request.URL.Query().Get("depth"), visibility.DefaultDepth, 1, trustdomain.MaxProjectionDepth)
		if !ok {
			writeTrustBoundsError(writer, request)
			return
		}
		nodes, ok := boundedQueryInt(request.URL.Query().Get("nodes"), visibility.DefaultNodes, 2, trustdomain.MaxProjectionNodes)
		if !ok {
			writeTrustBoundsError(writer, request)
			return
		}
		paths, err := service.Explain(request.Context(), visibility.Request{
			RequesterID: session.MemberID(), RootID: memberID, MaxDepth: depth, MaxNodes: nodes,
		})
		if err != nil {
			if errors.Is(err, visibility.ErrInvalidBounds) {
				writeTrustBoundsError(writer, request)
				return
			}
			writeTrustNotFound(writer, request)
			return
		}
		response := trustPathsResponse{Paths: make([]trustPathResponse, 0, len(paths))}
		for _, path := range paths {
			item := trustPathResponse{
				TargetID: path.TargetID, Hops: path.Hops,
				Steps: make([]trustPathStepResponse, 0, len(path.Steps)),
			}
			for _, step := range path.Steps {
				item.Steps = append(item.Steps, trustPathStepResponse{
					SourceID: step.SourceID, TargetID: step.TargetID, Reason: string(step.Reason),
				})
			}
			response.Paths = append(response.Paths, item)
		}
		writeSuccess(writer, request, http.StatusOK, response)
	})
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	returnToken := ""
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		returnToken = strings.TrimSpace(parts[1])
	}
	return returnToken, returnToken != ""
}

func validIdentifier(value string) bool {
	return value != "" && len(value) <= maxIdentifierLength && identifierPattern.MatchString(value)
}

func boundedQueryInt(raw string, fallback, minimum, maximum int) (int, bool) {
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.Atoi(raw)
	return value, err == nil && value >= minimum && value <= maximum
}

func writeTrustNotFound(writer http.ResponseWriter, request *http.Request) {
	writeError(writer, request, http.StatusNotFound, APIError{
		Code:    "trust_paths_not_found",
		Message: "Trust paths are not available.",
	})
}

func writeTrustBoundsError(writer http.ResponseWriter, request *http.Request) {
	writeError(writer, request, http.StatusBadRequest, APIError{
		Code:    "invalid_trust_path_bounds",
		Message: "Trust path limits are invalid.",
	})
}
