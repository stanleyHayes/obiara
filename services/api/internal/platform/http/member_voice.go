package apihttp

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

// MemberVoices reads another member's usable Voice of Introduction recordings.
type MemberVoices interface {
	RecordedByOwner(ctx context.Context, ownerID string) ([]domain.Introduction, error)
}

// RegisterMemberVoiceRoutes exposes hearing somebody else.
//
// This is what was missing: the media context would authorize a non-owner to
// hear a Voice of Introduction (agent_plan.md §63), and no route ever asked it
// to. Every playback path was scoped to the caller's own recording, so the
// twenty seconds of verified listening that arm a sow (FR-202) could never be
// accumulated against anybody, and no sow was ever possible.
//
// Gated at `introductions.view`, because hearing somebody is the romantic
// surface FR-101 puts behind Tier 1.
func RegisterMemberVoiceRoutes(
	mux *http.ServeMux,
	voices MemberVoices,
	playback IntroductionPlayback,
	sessions SessionAuthenticator,
	gate MemberGate,
) {
	mux.Handle("GET /v1/members/{memberId}/voice", gate.guard(
		sessions, "introductions.view", "introduction",
		memberVoiceHandler(voices, playback, sessions),
	))
}

type memberVoiceTake struct {
	// Prompt names which of the three questions this answers, so a listener
	// hears them in the order they were asked rather than in storage order.
	Prompt     string `json:"prompt"`
	AssetID    string `json:"assetId"`
	URL        string `json:"url"`
	ExpiresAt  string `json:"expiresAt"`
	DurationMs int64  `json:"durationMs"`
}

type memberVoiceResponse struct {
	MemberID string            `json:"memberId"`
	Takes    []memberVoiceTake `json:"takes"`
}

func memberVoiceHandler(
	voices MemberVoices,
	playback IntroductionPlayback,
	sessions SessionAuthenticator,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listenerID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if voices == nil || playback == nil {
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "introduction_unavailable", Message: "Voice introductions are unavailable.",
			})
			return
		}
		memberID := strings.TrimSpace(r.PathValue("memberId"))
		if memberID == "" || memberID == listenerID {
			// Your own voice is read through /v1/introductions, which is
			// where withdrawal and re-recording live. Answering here too
			// would be a second path to the same thing with different rules.
			writeError(w, r, http.StatusNotFound, APIError{
				Code: "voice_not_found", Message: "There is no voice to hear here.",
			})
			return
		}
		recorded, err := voices.RecordedByOwner(r.Context(), memberID)
		if err != nil {
			logServerError(r.Context(), r, http.StatusServiceUnavailable, "introduction_unavailable", err)
			writeError(w, r, http.StatusServiceUnavailable, APIError{
				Code: "introduction_unavailable", Message: "Voice introductions are unavailable.",
			})
			return
		}

		takes := make([]memberVoiceTake, 0, len(recorded))
		for _, introduction := range recorded {
			asset := introduction.Media().AssetID()
			if asset == "" {
				continue
			}
			// One grant per recording, each authorized on its own. Whether
			// this listener may hear it is the media context's decision — a
			// block, a withdrawal — and it is asked per asset rather than
			// assumed from the first answer.
			access, grantErr := playback.AuthorizePlayback(r.Context(), listenerID, asset)
			if grantErr != nil {
				continue
			}
			takes = append(takes, memberVoiceTake{
				Prompt:     string(introduction.Prompt()),
				AssetID:    asset,
				URL:        access.URL,
				ExpiresAt:  access.ExpiresAt.Format(time.RFC3339),
				DurationMs: introduction.Media().Duration().Milliseconds(),
			})
		}
		if len(takes) == 0 {
			// Nothing to hear and refused are answered the same way. A
			// distinct refusal would tell a member they had been blocked,
			// which is the signal a block exists to withhold — and a distinct
			// "they exist but have recorded nothing" would make this an
			// endpoint for finding out who exists.
			writeError(w, r, http.StatusNotFound, APIError{
				Code: "voice_not_found", Message: "There is no voice to hear here.",
			})
			return
		}
		sort.Slice(takes, func(i, j int) bool {
			return promptOrder(takes[i].Prompt) < promptOrder(takes[j].Prompt)
		})
		writeSuccess(w, r, http.StatusOK, memberVoiceResponse{MemberID: memberID, Takes: takes})
	})
}

// promptOrder is the order the three questions are asked in. Storage order is
// whatever the member happened to record first, which is not the order they
// should be heard in.
func promptOrder(prompt string) int {
	for index, known := range domain.Prompts {
		if string(known) == prompt {
			return index
		}
	}
	return len(domain.Prompts)
}
