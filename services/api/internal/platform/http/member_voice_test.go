package apihttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	introapplication "github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	introdomain "github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

// voicesStub is one member's recordings.
type voicesStub struct {
	recorded []introdomain.Introduction
	err      error
	askedFor string
}

func (s *voicesStub) RecordedByOwner(_ context.Context, ownerID string) ([]introdomain.Introduction, error) {
	s.askedFor = ownerID
	return s.recorded, s.err
}

// grantsFor mints a grant for the assets it knows and refuses the rest, which
// is how the media context answers a block or a withdrawal.
type grantsFor struct {
	allowed map[string]bool
	asked   []string
}

func (g *grantsFor) AuthorizePlayback(_ context.Context, _, assetID string) (introapplication.UploadAccess, error) {
	g.asked = append(g.asked, assetID)
	if !g.allowed[assetID] {
		return introapplication.UploadAccess{}, introapplication.ErrNotFound
	}
	return introapplication.UploadAccess{
		URL:       "https://bucket.example/" + assetID + "?sig=read",
		ExpiresAt: time.Date(2026, time.September, 5, 12, 10, 0, 0, time.UTC),
	}, nil
}

func takeFixture(t *testing.T, ownerID, id, assetID string, prompt introdomain.Prompt) introdomain.Introduction {
	t.Helper()
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	digest := strings.Repeat("a", 64)
	consent, err := introdomain.NewConsentSnapshot("voice.introduction", 1, now)
	if err != nil {
		t.Fatal(err)
	}
	media, err := introdomain.NewMediaRef(assetID, "audio/ogg", 184_320, 42*time.Second, digest)
	if err != nil {
		t.Fatal(err)
	}
	introduction, err := introdomain.New(
		id, ownerID, prompt, consent, media,
		introdomain.NewRetention(now.Add(180*24*time.Hour), false),
		introdomain.Command{ID: "cmd_" + id, Fingerprint: digest, At: now},
	)
	if err != nil {
		t.Fatal(err)
	}
	return introduction
}

func voiceRequest(t *testing.T, voices MemberVoices, playback IntroductionPlayback, memberID string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterMemberVoiceRoutes(mux, voices, playback, sessionStub{memberID: "listener-1"}, verifiedGate())
	request := httptest.NewRequest(http.MethodGet, "/v1/members/"+memberID+"/voice", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	return response
}

func TestAMemberCanHearAnotherMembersVoice(t *testing.T) {
	// The whole point. Until this route existed every playback path was
	// scoped to the caller's own recording, so the twenty seconds that arm a
	// sow could never be accumulated against anybody.
	voices := &voicesStub{recorded: []introdomain.Introduction{
		takeFixture(t, "member-2", "intro_3", "asset_welcome", introdomain.PromptWelcome),
		takeFixture(t, "member-2", "intro_1", "asset_arrival", introdomain.PromptArrival),
	}}
	grants := &grantsFor{allowed: map[string]bool{"asset_welcome": true, "asset_arrival": true}}

	response := voiceRequest(t, voices, grants, "member-2")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{"asset_arrival", "asset_welcome", "durationMs"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q: %s", want, body)
		}
	}
	// Heard in the order the three questions are asked, not in storage order.
	if strings.Index(body, "arrival") > strings.Index(body, "welcome") {
		t.Fatalf("takes came back out of order: %s", body)
	}
	if voices.askedFor != "member-2" {
		t.Fatalf("asked for %q", voices.askedFor)
	}
}

func TestEachRecordingIsAuthorizedOnItsOwn(t *testing.T) {
	// A grant is per asset. Minting them from one answer would hand out a
	// recording the media context had refused — a withdrawn take, say.
	voices := &voicesStub{recorded: []introdomain.Introduction{
		takeFixture(t, "member-2", "intro_1", "asset_arrival", introdomain.PromptArrival),
		takeFixture(t, "member-2", "intro_2", "asset_ordinary", introdomain.PromptOrdinary),
	}}
	grants := &grantsFor{allowed: map[string]bool{"asset_arrival": true}}

	response := voiceRequest(t, voices, grants, "member-2")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if len(grants.asked) != 2 {
		t.Fatalf("%d assets authorized, want both asked about", len(grants.asked))
	}
	if strings.Contains(response.Body.String(), "asset_ordinary") {
		t.Fatalf("a refused recording was handed out: %s", response.Body.String())
	}
}

func TestNothingToHearAndRefusedAnswerTheSameWay(t *testing.T) {
	// A distinct refusal would tell a member they had been blocked, which is
	// the signal a block exists to withhold. A distinct "they exist but have
	// recorded nothing" would make this an endpoint for finding out who
	// exists.
	for name, setup := range map[string]struct {
		voices   *voicesStub
		playback *grantsFor
	}{
		"nobody there": {
			voices: &voicesStub{}, playback: &grantsFor{},
		},
		"recorded but refused": {
			voices: &voicesStub{recorded: []introdomain.Introduction{
				takeFixture(t, "member-2", "intro_1", "asset_1", introdomain.PromptArrival),
			}},
			playback: &grantsFor{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			response := voiceRequest(t, setup.voices, setup.playback, "member-2")
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404: %s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), "voice_not_found") {
				t.Fatalf("body = %s", response.Body.String())
			}
		})
	}
}

func TestYourOwnVoiceIsNotReadThroughThisRoute(t *testing.T) {
	// /v1/introductions is where a member's own recording lives, along with
	// withdrawing and re-recording it. A second path to the same thing with
	// different rules is how the two drift apart.
	voices := &voicesStub{}
	response := voiceRequest(t, voices, &grantsFor{}, "listener-1")
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
	if voices.askedFor != "" {
		t.Fatal("the store was searched for the caller's own voice")
	}
}

func TestAnUnverifiedMemberHearsNobody(t *testing.T) {
	// FR-101: hearing somebody is a romantic surface, behind Tier 1.
	voices := &voicesStub{}
	mux := http.NewServeMux()
	RegisterMemberVoiceRoutes(mux, voices, &grantsFor{}, sessionStub{memberID: "listener-1"},
		gateAt(identityTierUnverified()))
	request := httptest.NewRequest(http.MethodGet, "/v1/members/member-2/voice", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
	if voices.askedFor != "" {
		t.Fatal("an unverified member's request reached the store")
	}
}

func TestAnUnreadableStoreIsNotAnEmptyOne(t *testing.T) {
	voices := &voicesStub{err: errors.New("mongo down")}
	response := voiceRequest(t, voices, &grantsFor{}, "member-2")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", response.Code, response.Body.String())
	}
}
