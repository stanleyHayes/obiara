package visibility

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type HTTPHandler struct {
	service   Service
	requester RequesterResolver
}

func NewHTTPHandler(service Service, requester RequesterResolver) HTTPHandler {
	return HTTPHandler{service: service, requester: requester}
}

type response struct {
	Paths []Explanation `json:"paths"`
}

func (handler HTTPHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", http.MethodGet)
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rootID, ok := trustPathRoot(request.URL.Path)
	if !ok || handler.requester == nil {
		http.NotFound(writer, request)
		return
	}
	requesterID, err := handler.requester.RequesterID(request.Context())
	if err != nil || strings.TrimSpace(requesterID) == "" {
		http.NotFound(writer, request)
		return
	}
	depth, validDepth := optionalPositiveInt(request.URL.Query().Get("depth"), DefaultDepth)
	nodes, validNodes := optionalPositiveInt(request.URL.Query().Get("nodes"), DefaultNodes)
	if !validDepth || !validNodes {
		http.Error(writer, "invalid bounds", http.StatusBadRequest)
		return
	}
	paths, err := handler.service.Explain(request.Context(), Request{
		RequesterID: requesterID, RootID: rootID, MaxDepth: depth, MaxNodes: nodes,
	})
	if errors.Is(err, ErrInvalidBounds) {
		http.Error(writer, "invalid bounds", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(response{Paths: paths})
}

func trustPathRoot(path string) (string, bool) {
	const prefix = "/v1/members/"
	const suffix = "/trust-paths"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return "", false
	}
	root := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	if root == "" || strings.Contains(root, "/") {
		return "", false
	}
	return root, true
}

func optionalPositiveInt(raw string, fallback int) (int, bool) {
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.Atoi(raw)
	return value, err == nil && value > 0
}
