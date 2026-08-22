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
		"  /v1/waitlist:",
		"  /v1/auth/otp:",
		"  /v1/auth/otp/verify:",
		"  /v1/verifications/ghana-card:",
		"  /v1/privacy/exports:",
		"  /v1/privacy/deletions:",
		"  /v1/privacy/requests/{id}:",
		"  /v1/members/{memberId}/trust-paths:",
		"  /v1/listening/heartbeats:",
		"  /v1/listening/eligibility/{assetId}:",
		"  /v1/fires:",
		"  /v1/fires/{id}/rsvps:",
		"  /v1/fires/{id}/rsvps/{memberId}:",
		"  /v1/fires/{id}/embers:",
		"  /v1/embers/{id}/redeem:",
		"  /v1/notification-preferences/{memberId}:",
		"  /v1/reports:",
		"  /v1/blocks:",
		"  /v1/blocks/{blockerId}/{blockedId}:",
		"  /webhooks/resend:",
		"  /v1/suban/marks/{memberId}:",
		"  /v1/suban/events/{memberId}:",
		"  /v1/admin/principals:",
		"  /v1/admin/login/start:",
		"  /v1/admin/login/complete:",
		"  /v1/admin/sessions/{id}/step-up/start:",
		"  /v1/admin/sessions/{id}/step-up/complete:",
		"  /v1/fires/{id}/close:",
		"  /v1/rooms/{roomId}/calls:",
		"  /v1/calls/{id}/end:",
		"  /v1/metrics/funnel:",
		"  /v1/scam-arc/signals:",
		"  /v1/metrics/deliveries:",
		"  /v1/consent/{memberId}:",
		"  /v1/consent/{memberId}/{purpose}:",
		"  /v1/admin/market-packs:",
		"  /v1/admin/market-packs/{id}/publish:",
		"  /v1/admin/market-packs/{id}/retire:",
		"  /v1/market-packs/published:",
		"  /v1/admin/verifications:",
		"  /v1/admin/verifications/{id}:",
		"  /v1/admin/verifications/{id}/evidence-access:",
		"  /v1/admin/verifications/{id}/decisions:",
		"  /v1/admin/safety/cases:",
		"  /v1/admin/safety/cases/{id}/assignment:",
		"  /v1/admin/safety/cases/{id}/evidence-access:",
		"  /v1/admin/care/cases:",
		"  /v1/admin/care/cases/{id}/engagement:",
		"  /v1/admin/care/cases/{id}/resolution:",
		"  /v1/admin/controls:",
		"  /v1/admin/members:",
		"  /v1/admin/waitlist:",
		"  /v1/admin/finance/reconciliation:",
		"  /v1/admin/account:",
		"  /v1/admin/controls/{id}/approval:",
		"  /v1/admin/controls/{id}/application:",
		"  /v1/doorway-question:",
		"  /v1/doorway-question/{memberId}:",
		"  /v1/photo-vault/items:",
		"  /v1/photo-vault/{ownerId}:",
		"  /v1/circles/{circleId}/oware:",
		"  /v1/circles/{circleId}/oware/{gameId}:",
		"  /v1/circles/{circleId}/oware/{gameId}/moves:",
		"  /v1/circles/{circleId}/stories:",
		"  /v1/circles/{circleId}/stories/{storyId}:",
		"  /v1/circles/{circleId}/stories/{storyId}/passages:",
		"  /v1/circles/{circleId}/stories/{storyId}/passages/{passageId}:",
		"  /v1/circles/{circleId}/stories/{storyId}/publication-grants:",
		"  /v1/circles/{circleId}/stories/{storyId}/publish:",
		"operationId: registerMember",
		"operationId: requestOtp",
		"operationId: verifyOtp",
		"operationId: submitGhanaCard",
		"operationId: requestExport",
		"operationId: requestDeletion",
		"operationId: privacyRequestStatus",
		"operationId: listAdminVerificationQueue",
		"operationId: getAdminVerificationCase",
		"operationId: accessAdminVerificationEvidence",
		"operationId: decideAdminVerificationCase",
		"operationId: listAdminSafetyCases",
		"operationId: assignAdminSafetyCase",
		"operationId: accessAdminSafetyEvidence",
		"operationId: listAdminCareCases",
		"operationId: engageAdminCareCase",
		"operationId: resolveAdminCareCase",
		"operationId: listAdminRuntimeControls",
		"operationId: proposeAdminRuntimeControl",
		"operationId: approveAdminRuntimeControl",
		"operationId: applyAdminRuntimeControl",
		"operationId: setDoorwayQuestion",
		"operationId: createOwareGame",
		"operationId: getOwareGame",
		"operationId: moveOwareGame",
		"operationId: createAnansesemStory",
		"operationId: getAnansesemStory",
		"operationId: addAnansesemPassage",
		"operationId: editAnansesemPassage",
		"operationId: grantAnansesemPublication",
		"operationId: publishAnansesemStory",
		"operationId: getOwnDoorwayQuestion",
		"operationId: getDoorwayQuestion",
		"operationId: addVaultItem",
		"operationId: viewVault",
		"operationId: getMemberTrustPaths",
		"operationId: recordListeningHeartbeats",
		"operationId: getListeningEligibility",
		"operationId: scheduleFire",
		"operationId: listUpcomingFires",
		"operationId: rsvpFire",
		"operationId: cancelFireRsvp",
		"operationId: issueEmber",
		"operationId: redeemEmber",
		"operationId: getNotificationPreferences",
		"operationId: configureNotificationPreferences",
		"operationId: getOwnNotificationPreferences",
		"operationId: configureOwnNotificationPreferences",
		"operationId: fileReport",
		"operationId: blockMember",
		"operationId: unblockMember",
		"operationId: resendDeliveryWebhook",
		"operationId: getSubanMarks",
		"operationId: getSubanEvents",
		"operationId: enrollAdminPrincipal",
		"operationId: startAdminLogin",
		"operationId: completeAdminLogin",
		"operationId: startAdminStepUp",
		"operationId: completeAdminStepUp",
		"operationId: closeFireToEmbers",
		"operationId: initiateCall",
		"operationId: endCall",
		"operationId: getFunnelMetrics",
		"operationId: observeScamArcSignal",
		"operationId: getDeliveryStats",
		"operationId: getConsentSwitchboard",
		"operationId: setConsentPurpose",
		"operationId: draftMarketPack",
		"operationId: listAdminMarketPacks",
		"operationId: getAdminFinanceReconciliation",
		"operationId: getAdminAccount",
		"operationId: publishMarketPack",
		"operationId: retireMarketPack",
		"operationId: listPublishedMarketPacks",
		"  /v1/nominations:",
		"  /v1/nominations/{id}/consent:",
		"  /v1/nominations/{id}/decline:",
		"operationId: nominateKin",
		"operationId: listNominations",
		"operationId: consentNomination",
		"operationId: declineNomination",
		"name: Idempotency-Key",
		"name: X-Correlation-ID",
		"additionalProperties: false",
		"invalid_json",
		"validation_failed",
		"otp_rate_limited",
		"verification_rejected",
		"legal_hold_active",
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
	if len(seen) != 167 {
		t.Errorf("operationId count = %d, want 167", len(seen))
	}
}
