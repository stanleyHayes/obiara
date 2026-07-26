package openapi_test

import (
	"os"
	"strings"
	"testing"
)

// This lightweight guard intentionally uses only the standard library. The
// generated-client job performs full OpenAPI parsing; this test catches
// accidental removal or renaming of contract elements implemented by Go.
func TestContractContainsImplementedSurface(t *testing.T) {
	document, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	source := string(document)
	required := []string{
		"openapi: 3.1.0",
		"  /live:",
		"  /ready:",
		"  /v1/members:",
		"  /v1/auth/otp:",
		"  /v1/auth/otp/verify:",
		"  /v1/verifications/ghana-card:",
		"  /v1/privacy/exports:",
		"  /v1/privacy/deletions:",
		"  /v1/privacy/requests/{id}:",
		"  /v1/members/{memberId}/trust-paths:",
		"  /v1/admin/verifications:",
		"  /v1/admin/verifications/{id}:",
		"  /v1/admin/verifications/{id}/evidence-access:",
		"  /v1/admin/verifications/{id}/decisions:",
		"operationId: registerMember",
		"operationId: requestOtp",
		"operationId: verifyOtp",
		"operationId: submitGhanaCard",
		"operationId: requestExport",
		"operationId: requestDeletion",
		"operationId: privacyRequestStatus",
		"operationId: getMemberTrustPaths",
		"operationId: listAdminVerificationQueue",
		"operationId: getAdminVerificationCase",
		"operationId: accessAdminVerificationEvidence",
		"operationId: decideAdminVerificationCase",
		"name: Idempotency-Key",
		"name: X-Correlation-ID",
		"additionalProperties: false",
		"invalid_json",
		"validation_failed",
		"otp_rate_limited",
		"verification_rejected",
		"legal_hold_active",
		"invalid_trust_path_bounds",
		"trust_paths_not_found",
		"MemberBearer:",
		"correlationId:",
	}
	for _, token := range required {
		if !strings.Contains(source, token) {
			t.Errorf("contract missing %q", token)
		}
	}
}

func TestOperationIDsAreUnique(t *testing.T) {
	document, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	seen := make(map[string]struct{})
	for _, line := range strings.Split(string(document), "\n") {
		line = strings.TrimSpace(line)
		const prefix = "operationId:"
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		id := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if _, exists := seen[id]; exists {
			t.Errorf("duplicate operationId %q", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != 14 {
		t.Errorf("operationId count = %d, want 14", len(seen))
	}
}
