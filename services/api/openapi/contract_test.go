package openapi_test

import (
	"os"
	"regexp"
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
	// The count is deliberate, not incidental: adding a route to the
	// contract is a decision, and this guard makes an accidental one fail
	// loudly. 194 adds getOnboardingStatus, which lets a member resume the
	// walk instead of paying for it again after a refresh; adminLogout,
	// without which signing out of the console left the session id live
	// upstream until it expired on its own; submitGhanaCardDocuments, which
	// moved identity out of signing up so an outage at the card provider can
	// no longer stop anyone creating an account; and the four Voice of
	// Introduction routes, which is the first time that context has been
	// reachable at all — its domain and application layers have been
	// complete and untouchable since S2-021. 195 adds playVoiceIntroduction,
	// without which a recording could not be heard by anyone and the twenty
	// seconds of verified listening that arms Sow could never accumulate. 198
	// adds the three introduction-source routes — the first time a member can
	// ask to be introduced through a circle they belong to. 201 writes down the
	// three seed-stage routes — sprout, doorway exchange and decline — which
	// had been served and undocumented; TestEveryServedRouteIsInTheContract
	// now makes that impossible to repeat, in both directions. 202 adds
	// applyAdminSafetyAction: the T&S ladder had been enforced inside a
	// service registered on no route, so a case could be queued, assigned and
	// read, and then nothing could happen to it.
	// 203 adds sendSow: the atomic gesture itself, which had a complete
	// aggregate, an atomic allowance spend and a screening chain, and no
	// route at all.
	// 205 adds the screening review queue and its decision — the surface
	// that makes "a person reads every sow before delivery" something
	// somebody can do, rather than a policy with nowhere to happen.
	// 207 adds the pod: placing one and opening one. It is the last step of
	// the sow — until it existed, releasing a sow marked it delivered and
	// nobody received anything.
	// 208 adds the house front. Without it a member could only open a pod
	// whose id they already knew, and nothing told them — a house front with
	// no door.
	// 209 adds hearing somebody else. Every playback route before it was
	// scoped to the caller's own recording, so the twenty seconds that arm a
	// sow could never be accumulated against anybody and no sow was possible.
	// 210 adds opening a sow's recording. The only upload path in the
	// product made a Voice of Introduction, so a sow needed a recording that
	// could not be made.
	// 215 adds the organizations: list, register, suspend, restore, rename.
	// A discount code needs an issuer to hang off, and an audit trail needs
	// somebody to name.
	// 217 adds buying a membership and the provider's callback. Before them
	// membership.Service.Grant had no callers at all, so no pass could exist
	// and nothing in the product could be bought.
	if len(seen) != 217 {
		t.Errorf("operationId count = %d, want 217", len(seen))
	}
}

// TestEveryComponentRefResolves closes a blind spot that caught me twice.
//
// The Go contract tests read the document as YAML and never follow a $ref, so
// a reference to a response or schema that does not exist passes here and
// fails only in the TypeScript generator — which fails silently from Go's
// side, leaving the generated client quietly stale. Both times it was an
// invented name that reads perfectly plausibly: "Forbidden" and
// "FeatureUnavailable", where the real ones are "AdminRoleRequired" and
// "ServiceUnavailable".
func TestEveryComponentRefResolves(t *testing.T) {
	raw, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)

	defined := map[string]bool{}
	for _, section := range []string{"responses", "schemas", "parameters"} {
		for _, name := range definedComponents(text, section) {
			defined[section+"/"+name] = true
		}
	}

	referenced := regexp.MustCompile(`#/components/(responses|schemas|parameters)/([A-Za-z0-9_]+)`)
	seen := map[string]bool{}
	for _, match := range referenced.FindAllStringSubmatch(text, -1) {
		key := match[1] + "/" + match[2]
		if seen[key] {
			continue
		}
		seen[key] = true
		if !defined[key] {
			t.Errorf("$ref to #/components/%s does not resolve; the generator refuses the whole document", key)
		}
	}
	if len(seen) == 0 {
		t.Fatal("no component references found, so this test is checking nothing")
	}
}

// definedComponents lists the keys directly under components.<section>.
func definedComponents(text, section string) []string {
	start := strings.Index(text, "\n  "+section+":\n")
	if start < 0 {
		return nil
	}
	rest := text[start+len("\n  "+section+":\n"):]
	var names []string
	for _, line := range strings.Split(rest, "\n") {
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.TrimSpace(line) != "" {
			break // the next section at the same indent
		}
		if match := regexp.MustCompile(`^    ([A-Za-z0-9_]+):$`).FindStringSubmatch(line); match != nil {
			names = append(names, match[1])
		}
	}
	return names
}
