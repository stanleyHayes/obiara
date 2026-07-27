package health

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/internal/quality/performance"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLiveBoundedLoad(t *testing.T) {
	handler := Live()
	profile := performance.Profile{Name: "api_live", Requests: 2000, Concurrency: 16, MaxP95: 100 * time.Millisecond, MaxErrorRate: 0}
	result, e := performance.Run(context.Background(), profile, func(ctx context.Context, index int) error {
		request := httptest.NewRequest(http.MethodGet, "/live", nil).WithContext(ctx)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Body.String() != "ok" {
			return fmt.Errorf("unexpected response")
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if !result.Within(profile) {
		t.Fatalf("load budget failed: %#v", result)
	}
	if raw, e := result.JSON(); e == nil {
		t.Logf("performance_evidence=%s", raw)
	}
}
