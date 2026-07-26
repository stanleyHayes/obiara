// Package simulator provides a deterministic liveness provider for local
// development and provider-contract tests. It never examines biometric bytes.
package simulator

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
)

var ErrUnavailable = errors.New("simulated liveness provider unavailable")

type Provider struct {
	mu       sync.Mutex
	requests []application.ProviderRequest
}

func NewProvider() *Provider {
	return &Provider{}
}

func (provider *Provider) Assess(_ context.Context, request application.ProviderRequest) (application.ProviderResult, error) {
	provider.mu.Lock()
	provider.requests = append(provider.requests, request)
	provider.mu.Unlock()
	if strings.TrimSpace(request.CommandID) == "" || strings.TrimSpace(request.AttemptID) == "" ||
		strings.TrimSpace(request.VoiceArtifactRef) == "" || strings.TrimSpace(request.FaceArtifactRef) == "" {
		return application.ProviderResult{}, errors.New("invalid simulated liveness request")
	}
	switch {
	case strings.HasSuffix(request.VoiceArtifactRef, ":outage"):
		return application.ProviderResult{}, ErrUnavailable
	case strings.HasSuffix(request.VoiceArtifactRef, ":uncertain"):
		return application.ProviderResult{
			Outcome:     application.OutcomeUncertain,
			ProviderRef: "sim:" + request.AttemptID,
		}, nil
	case strings.HasSuffix(request.VoiceArtifactRef, ":fail"):
		return application.ProviderResult{
			Outcome:     application.OutcomeNotLive,
			ProviderRef: "sim:" + request.AttemptID,
		}, nil
	default:
		return application.ProviderResult{
			Outcome:     application.OutcomeLive,
			ProviderRef: "sim:" + request.AttemptID,
		}, nil
	}
}

func (provider *Provider) Requests() []application.ProviderRequest {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return append([]application.ProviderRequest(nil), provider.requests...)
}
