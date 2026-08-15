package apihttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/platform/flags"
)

type flagEvaluatorStub struct{ decisions map[flags.Flag]flags.Decision }

func (stub flagEvaluatorStub) Evaluate(flag flags.Flag) flags.Decision {
	return stub.decisions[flag]
}

func TestFeatureFlagsFailClosedAndLeaveControlPlaneOpen(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	evaluator := flagEvaluatorStub{decisions: map[flags.Flag]flags.Decision{
		flags.FlagFires: {Flag: flags.FlagFires, Known: true, Enabled: true},
		flags.FlagGate:  {Flag: flags.FlagGate, Known: true, Enabled: true, Killed: true},
	}}
	handler := FeatureFlags(next, evaluator)
	for _, testCase := range []struct {
		path string
		want int
	}{
		{"/v1/fires", http.StatusNoContent},
		{"/v1/doorway-question", http.StatusServiceUnavailable},
		{"/v1/escrows", http.StatusServiceUnavailable},
		{"/v1/admin/controls", http.StatusNoContent},
		{"/live", http.StatusNoContent},
	} {
		request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != testCase.want {
			t.Fatalf("%s status=%d want=%d body=%s", testCase.path, response.Code, testCase.want, response.Body.String())
		}
	}
}

func TestPaymentsKillSwitchCoversRegisteredPaymentRoutes(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	evaluator := flagEvaluatorStub{decisions: map[flags.Flag]flags.Decision{
		flags.FlagPayments: {Flag: flags.FlagPayments, Known: true, Enabled: true, Killed: true},
	}}
	handler := FeatureFlags(next, evaluator)
	for _, path := range []string{
		// Marketplace listing (gated by the pre-existing prefix).
		"/v1/matchmakers",
		// Registered fee-bearing engagement routes (matchmaker.go).
		"/v1/matchmaker-engagements",
		"/v1/matchmaker-engagements/eng_1",
		"/v1/matchmaker-engagements/eng_1/member-consent",
		// Registered membership routes (membership.go).
		"/v1/membership",
		"/v1/membership/cancel",
		"/v1/membership/refunds",
		// Escrow routes.
		"/v1/escrows",
		"/v1/escrows/esc_1/disputes",
	} {
		request := httptest.NewRequest(http.MethodPost, path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d want=%d: payments kill switch must gate this route", path, response.Code, http.StatusServiceUnavailable)
		}
	}
}
