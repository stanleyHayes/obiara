package apihttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// tierStub stands a member on a chosen rung of the ladder.
type tierStub struct {
	tier identitydomain.Tier
	err  error
}

func (s tierStub) Tier(context.Context, string) (identitydomain.Tier, error) {
	return s.tier, s.err
}

// gateAt builds a gate over a member standing on the given rung.
func gateAt(tier identitydomain.Tier) MemberGate { return NewMemberGate(tierStub{tier: tier}) }

// verifiedGate stands the test's member on Tier 1, which is what the route
// tables under test now require of a romantic surface.
func verifiedGate() MemberGate { return gateAt(identitydomain.TierVerified) }

func gatedRequest(t *testing.T, gate MemberGate, action, resourceType string) *httptest.ResponseRecorder {
	t.Helper()
	reached := false
	handler := gate.guard(sessionStub{memberID: "member_1"}, action, resourceType,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			reached = true
			w.WriteHeader(http.StatusOK)
		}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Authorization", "Bearer token")
	handler.ServeHTTP(recorder, request)
	if reached && recorder.Code != http.StatusOK {
		t.Fatal("the handler ran but the refusal was still written")
	}
	return recorder
}

func TestAnUnverifiedMemberCannotReachARomanticSurface(t *testing.T) {
	// This is the whole point of FR-101a, and until now every one of these
	// surfaces was open to a Tier-0 account.
	for _, surface := range []struct{ action, resource string }{
		{"introductions.view", "introduction"},
		{"rooms.participate", "room"},
	} {
		recorder := gatedRequest(t, gateAt(identitydomain.TierUnverified), surface.action, surface.resource)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("%s admitted an unverified member: %d", surface.action, recorder.Code)
		}
		// A bare refusal leaves the member with nothing to do about it.
		if !strings.Contains(recorder.Body.String(), "tier_1_required") {
			t.Fatalf("%s refused without naming the rung: %s", surface.action, recorder.Body.String())
		}
	}
}

func TestVerificationOpensTheRomanticSurfaces(t *testing.T) {
	// The opposite failure — gating a surface a member has already earned —
	// is the one members actually feel, so it is worth its own test.
	for _, action := range []string{"introductions.view", "rooms.participate"} {
		if code := gatedRequest(t, gateAt(identitydomain.TierVerified), action, resourceFor(action)).Code; code != http.StatusOK {
			t.Fatalf("%s refused a verified member: %d", action, code)
		}
	}
}

func TestSowingNeedsTheSowingRung(t *testing.T) {
	// FR-101b. Tier 1 is not enough to reach toward someone.
	refused := gatedRequest(t, gateAt(identitydomain.TierVerified), "seeds.sow", "seed")
	if refused.Code != http.StatusForbidden {
		t.Fatalf("a Tier-1 member sowed: %d", refused.Code)
	}
	if !strings.Contains(refused.Body.String(), "tier_2_required") {
		t.Fatalf("sowing refused without naming the rung: %s", refused.Body.String())
	}
	if code := gatedRequest(t, gateAt(identitydomain.TierSowing), "seeds.sow", "seed").Code; code != http.StatusOK {
		t.Fatalf("a Tier-2 member could not sow: %d", code)
	}
}

func TestAnUnreadableTierIsAFaultNotAVerdict(t *testing.T) {
	// Answering 403 here would be indistinguishable from "you are unverified"
	// and would lock out members who have already earned the surface, which is
	// the worse of the two ways this can go wrong.
	gate := NewMemberGate(tierStub{tier: identitydomain.TierSowing, err: errors.New("mongo is down")})
	if code := gatedRequest(t, gate, "introductions.view", "introduction").Code; code != http.StatusInternalServerError {
		t.Fatalf("an unreadable tier answered %d, want 500", code)
	}
	// A gate wired with no tier read must not admit everybody.
	if code := gatedRequest(t, MemberGate{}, "introductions.view", "introduction").Code; code != http.StatusInternalServerError {
		t.Fatalf("a gate with no tier reader answered %d, want 500", code)
	}
}

func TestAnUngatedActionIsStillDeniedByDefault(t *testing.T) {
	// The kernel is deny-by-default: an action nobody wrote a grant for must
	// be refused, not waved through because it has no tier.
	recorder := gatedRequest(t, gateAt(identitydomain.TierSowing), "surfaces.invented", "thing")
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("an action with no grant answered %d, want 403", recorder.Code)
	}
}

func resourceFor(action string) string {
	if action == "introductions.view" {
		return "introduction"
	}
	return "room"
}

// exitRoutes are the routes a member must be able to reach standing on any
// rung, including Tier 0 after a safety demotion. Gating any of them would
// make the ladder a trap: you could not leave a room, take down your own
// recording, decline, or report the person you need to get away from.
var exitRoutes = []string{
	`DELETE /v1/introductions/{id}`,
	`DELETE /v1/seed/sources/{id}`,
	`POST /v1/seed/declines`,
	`POST /v1/courtship/rooms/{id}/pause`,
	`POST /v1/courtship/rooms/{id}/closure`,
	`POST /v1/courtship/rooms/{id}/safety/block`,
	`POST /v1/courtship/rooms/{id}/safety/report`,
	`POST /v1/courtship/proposals/{id}/reject`,
	`POST /v1/courtship/proposals/{id}/withdraw`,
	// Leaving is never participation. A member who cannot join a circle must
	// still be able to walk out of one they are already in.
	`POST /v1/circles/{id}/leave`,
	`POST /v1/game-cohorts/{cohortId}/leave`,
	`DELETE /v1/fires/{id}/rsvps/{memberId}`,
	// Taking your own entry back out of a circle room.
	`POST /v1/circle-room-entries/{id}/delete`,
	// Removing somebody from a circle protects the people in it. A host must
	// be able to do it at any rung, for the same reason blocking is open.
	`POST /v1/circles/{id}/members/{memberId}/expel`,
}

func TestTheGateNeverBlocksTheRoadOutOfTheGate(t *testing.T) {
	// Read from source rather than from a live router: the invariant is that
	// nobody ever wraps these lines, and that is a property of the
	// registration tables themselves.
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	registrations := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		source, err := os.ReadFile(filepath.Join(".", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(source), "\n") {
			for _, route := range exitRoutes {
				if strings.Contains(line, `mux.Handle("`+route+`"`) {
					registrations[route] = strings.TrimSpace(line)
				}
			}
		}
	}
	for _, route := range exitRoutes {
		line, found := registrations[route]
		if !found {
			// A renamed or deleted exit route must fail here rather than
			// quietly stop being checked.
			t.Fatalf("exit route %q is no longer registered; if it moved, move it in this test too", route)
		}
		if strings.Contains(line, "gate.guard") {
			t.Fatalf("exit route %q is behind the tier gate: %s", route, line)
		}
	}
}

func TestAnUnverifiedMemberMayLookButNotTakePart(t *testing.T) {
	// M1-AC-01, as the owner ruled it: browse yes, participate no. Before
	// this an unverified account could start a circle, ask to join one,
	// schedule a fire and enter a game cohort.
	for _, surface := range []struct{ action, resource string }{
		{"circles.participate", "circle"},
		{"games.play", "game"},
		{"fires.attend", "fire"},
	} {
		refused := gatedRequest(t, gateAt(identitydomain.TierUnverified), surface.action, surface.resource)
		if refused.Code != http.StatusForbidden {
			t.Fatalf("%s admitted an unverified member: %d", surface.action, refused.Code)
		}
		if !strings.Contains(refused.Body.String(), "tier_1_required") {
			t.Fatalf("%s refused without naming the rung: %s", surface.action, refused.Body.String())
		}
		if allowed := gatedRequest(t, gateAt(identitydomain.TierVerified), surface.action, surface.resource); allowed.Code != http.StatusOK {
			t.Fatalf("%s refused a verified member: %d", surface.action, allowed.Code)
		}
	}
}

// gatedRoutes are routes that must stay behind the tier gate. The exit test
// above is only half the invariant: it proves the gate never blocks the way
// out, and nothing proved the gate is still on the way in. Removing a
// gate.guard from a participation route was, until this existed, a silent
// change that every test still passed.
var gatedRoutes = []string{
	`POST /v1/introductions`,
	`GET /v1/introductions/{id}/audio`,
	`POST /v1/listening/heartbeats`,
	`POST /v1/seed/sources`,
	`POST /v1/seed/sprouts`,
	`POST /v1/courtship/rooms`,
	`POST /v1/courtship/rooms/{id}/turns`,
	`POST /v1/courtship/proposals`,
	`POST /v1/circles`,
	`POST /v1/circles/{id}/requests`,
	`POST /v1/fires`,
	`POST /v1/game-cohorts/{cohortId}/join`,
	// Taking a turn inside a circle. Membership already requires a gated
	// join, so these are gated twice over for a new account — but an account
	// that was already a member before the join was gated reaches them with
	// nothing else in the way.
	`POST /v1/circles/{circleId}/room`,
	`POST /v1/circles/{circleId}/stories`,
	`POST /v1/circles/{circleId}/stories/{storyId}/passages`,
	`POST /v1/circles/{circleId}/ampe`,
	`POST /v1/circles/{circleId}/ampe/{roundId}/commands`,
	`POST /v1/circles/{circleId}/ebe`,
	`POST /v1/circles/{circleId}/ebe/{duelId}/answers`,
	`POST /v1/circles/{circleId}/oware`,
	`POST /v1/circles/{circleId}/oware/{gameId}/moves`,
}

func TestEveryParticipationRouteIsStillBehindTheGate(t *testing.T) {
	registrations := map[string]string{}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		source, err := os.ReadFile(filepath.Join(".", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(source), "\n") {
			for _, route := range gatedRoutes {
				if strings.Contains(line, `mux.Handle("`+route+`"`) {
					registrations[route] = strings.TrimSpace(line)
				}
			}
		}
	}
	for _, route := range gatedRoutes {
		line, found := registrations[route]
		if !found {
			t.Fatalf("gated route %q is no longer registered; if it moved, move it in this test too", route)
		}
		if !strings.Contains(line, "gate.guard") {
			t.Fatalf("participation route %q is no longer behind the tier gate: %s", route, line)
		}
	}
}
