package apihttp

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/platform/flags"
)

type FeatureFlagEvaluator interface {
	Evaluate(flags.Flag) flags.Decision
}

// FeatureFlags enforces release and kill-switch decisions at the HTTP
// composition boundary. Administrative, health and authentication routes are
// deliberately outside product capability flags.
func FeatureFlags(next http.Handler, evaluator FeatureFlagEvaluator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flag, gated := capabilityForPath(r.URL.Path)
		if !gated {
			next.ServeHTTP(w, r)
			return
		}
		decision := evaluator.Evaluate(flag)
		if !decision.Known || !decision.Enabled || decision.Killed {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "feature_unavailable", Message: "This capability is not currently available.",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func capabilityForPath(path string) (flags.Flag, bool) {
	switch {
	case strings.HasPrefix(path, "/v1/fires"), strings.HasPrefix(path, "/v1/embers"):
		return flags.FlagFires, true
	case strings.HasPrefix(path, "/v1/listening"), strings.HasPrefix(path, "/v1/garden"):
		return flags.FlagSow, true
	case strings.HasPrefix(path, "/v1/matchmakers"), strings.HasPrefix(path, "/v1/memberships"), strings.HasPrefix(path, "/v1/escrows"):
		return flags.FlagPayments, true
	case strings.HasPrefix(path, "/v1/doorway-question"), strings.HasPrefix(path, "/v1/photo-vault"):
		return flags.FlagGate, true
	default:
		return "", false
	}
}
