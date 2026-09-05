package secrets

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func env(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}
func TestValidateRuntime(t *testing.T) {
	now := time.Date(2026, 7, 27, 6, 0, 0, 0, time.UTC)
	values := map[string]string{"MONGODB_URI": "not-inspected", "MONGODB_URI_ROTATED_AT": now.Add(-30 * 24 * time.Hour).Format(time.RFC3339), "RESEND_WEBHOOK_SECRET": "not-inspected", "RESEND_WEBHOOK_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "LIVENESS_HMAC_SECRET": "not-inspected", "LIVENESS_HMAC_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "COMMERCE_HMAC_SECRET": "not-inspected", "COMMERCE_HMAC_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "ADMIN_HMAC_SECRET": "not-inspected", "ADMIN_HMAC_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "NNOBOA_INVITE_SECRET": "not-inspected", "NNOBOA_INVITE_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "SEED_HMAC_SECRET": "not-inspected", "SEED_HMAC_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "SAFEGUARDING_HMAC_SECRET": "not-inspected", "SAFEGUARDING_HMAC_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "CIRCLE_HMAC_SECRET": "not-inspected", "CIRCLE_HMAC_SECRET_ROTATED_AT": now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)}
	if e := ValidateRuntime(API, "staging", env(values), now); e != nil {
		t.Fatal(e)
	}
	if e := ValidateRuntime(Worker, "staging", env(values), now); e != nil {
		t.Fatal(e)
	}
}
func TestValidateRuntimeFailsClosedWithoutLeakingValues(t *testing.T) {
	now := time.Now().UTC()
	secretValue := "unique-do-not-print-value"
	cases := []struct {
		name   string
		values map[string]string
		want   error
	}{
		{"missing", map[string]string{}, ErrMissing},
		{"bad metadata", map[string]string{"MONGODB_URI": secretValue, "MONGODB_URI_ROTATED_AT": "yesterday"}, ErrRotationMetadata},
		{"stale", map[string]string{"MONGODB_URI": secretValue, "MONGODB_URI_ROTATED_AT": now.Add(-91 * 24 * time.Hour).Format(time.RFC3339)}, ErrStale},
		{"future", map[string]string{"MONGODB_URI": secretValue, "MONGODB_URI_ROTATED_AT": now.Add(time.Hour).Format(time.RFC3339)}, ErrRotationMetadata},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := ValidateRuntime(Worker, "staging", env(tc.values), now)
			if !errors.Is(e, tc.want) {
				t.Fatalf("got %v", e)
			}
			if strings.Contains(e.Error(), secretValue) {
				t.Fatal("secret value leaked in error")
			}
		})
	}
}
func TestDevelopmentDoesNotRequireManagedSecrets(t *testing.T) {
	if e := ValidateRuntime(API, "development", env(nil), time.Now()); e != nil {
		t.Fatal(e)
	}
}
