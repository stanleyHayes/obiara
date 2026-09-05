package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	introapplication "github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	introdomain "github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

type introServiceStub struct {
	begin   func(context.Context, introapplication.BeginUploadRequest) (introapplication.BeginUploadResult, error)
	confirm func(context.Context, string, string) (introdomain.Introduction, error)
	speak   func(context.Context, string, string) (introapplication.TranscribeResult, error)
	revoke  func(context.Context, string, string) (introdomain.Introduction, error)
}

func (s introServiceStub) BeginUpload(ctx context.Context, r introapplication.BeginUploadRequest) (introapplication.BeginUploadResult, error) {
	return s.begin(ctx, r)
}
func (s introServiceStub) ConfirmUpload(ctx context.Context, id, cmd string) (introdomain.Introduction, error) {
	return s.confirm(ctx, id, cmd)
}
func (s introServiceStub) Transcribe(ctx context.Context, id, cmd string) (introapplication.TranscribeResult, error) {
	return s.speak(ctx, id, cmd)
}
func (s introServiceStub) Revoke(ctx context.Context, id, cmd string) (introdomain.Introduction, error) {
	return s.revoke(ctx, id, cmd)
}

type introReaderStub struct {
	introduction introdomain.Introduction
	err          error
}

type introPlaybackStub struct {
	access introapplication.UploadAccess
	err    error
}

func (s introPlaybackStub) AuthorizePlayback(context.Context, string, string) (introapplication.UploadAccess, error) {
	return s.access, s.err
}

func (s introReaderStub) FindByID(context.Context, string) (introdomain.Introduction, error) {
	return s.introduction, s.err
}

func introFixture(t *testing.T, ownerID string) introdomain.Introduction {
	t.Helper()
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	digest := strings.Repeat("a", 64)
	consent, err := introdomain.NewConsentSnapshot("voice.introduction", 1, now)
	if err != nil {
		t.Fatal(err)
	}
	media, err := introdomain.NewMediaRef("intro_asset_1", "audio/ogg", 184_320, 42*time.Second, digest)
	if err != nil {
		t.Fatal(err)
	}
	introduction, err := introdomain.New(
		"introduction_1", ownerID, introdomain.PromptArrival, consent, media,
		introdomain.NewRetention(now.Add(180*24*time.Hour), false),
		introdomain.Command{ID: "cmd_1", Fingerprint: digest, At: now},
	)
	if err != nil {
		t.Fatal(err)
	}
	return introduction
}

func introMux(t *testing.T, service VoiceIntroduction, reader IntroductionReader, memberID string) http.Handler {
	t.Helper()
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	session := identitydomain.Reconstitute(
		"session-1", memberID, "device-1", identitydomain.StatusActive,
		"access", now.Add(time.Hour), "refresh", "", now.Add(24*time.Hour), 1, now, now,
	)
	mux := http.NewServeMux()
	RegisterIntroductionRoutes(mux, service, reader, introPlaybackStub{
		access: introapplication.UploadAccess{
			URL:       "https://bucket.example/obj?sig=read",
			ExpiresAt: time.Date(2026, time.September, 5, 12, 10, 0, 0, time.UTC),
		},
	},
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return session, nil
		}}, verifiedGate())
	return Correlation(mux)
}

func TestBeginUploadReturnsAGrantTheClientUploadsWith(t *testing.T) {
	introduction := introFixture(t, "member-1")
	service := introServiceStub{
		begin: func(_ context.Context, request introapplication.BeginUploadRequest) (introapplication.BeginUploadResult, error) {
			if request.OwnerID != "member-1" {
				t.Fatalf("owner = %q; the session must decide it, never the body", request.OwnerID)
			}
			if request.PurposeID != "voice.introduction" {
				t.Fatalf("purpose = %q", request.PurposeID)
			}
			return introapplication.BeginUploadResult{
				Introduction: introduction,
				Access: introapplication.UploadAccess{
					URL: "https://bucket.example/obj?sig=x", ExpiresAt: time.Now().Add(time.Minute),
				},
			}, nil
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/introductions", strings.NewReader(`{"contentType":"audio/ogg","prompt":"arrival"}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "cmd-begin-1")
	response := httptest.NewRecorder()
	introMux(t, service, introReaderStub{introduction: introduction}, "member-1").ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct{ Data introductionResponse }
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.UploadURL == "" {
		t.Fatal("no upload grant returned; the client has nowhere to send the audio")
	}
	if envelope.Data.DurationMs != 42_000 {
		t.Fatalf("durationMs = %d, want 42000", envelope.Data.DurationMs)
	}
}

func TestBeginUploadRequiresAnIdempotencyKey(t *testing.T) {
	// Without one a retried recording opens a second aggregate and the member
	// is billed a second transcription for the same take.
	service := introServiceStub{begin: func(context.Context, introapplication.BeginUploadRequest) (introapplication.BeginUploadResult, error) {
		t.Fatal("service must not be reached without a command id")
		return introapplication.BeginUploadResult{}, nil
	}}
	request := httptest.NewRequest(http.MethodPost, "/v1/introductions", strings.NewReader(`{}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	introMux(t, service, introReaderStub{}, "member-1").ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
}

func TestAnotherMembersRecordingReadsAsAbsentNotForbidden(t *testing.T) {
	// 403 would confirm the id exists. On a surface built to avoid disclosing
	// who is a member, that is the disclosure.
	owned := introFixture(t, "member-OTHER")
	service := introServiceStub{
		revoke: func(context.Context, string, string) (introdomain.Introduction, error) {
			t.Fatal("the service must not be reached for someone else's recording")
			return introdomain.Introduction{}, nil
		},
	}
	for _, probe := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v1/introductions/introduction_1"},
		{http.MethodDelete, "/v1/introductions/introduction_1"},
	} {
		request := httptest.NewRequest(probe.method, probe.path, nil)
		request.Header.Set("Authorization", "Bearer token")
		request.Header.Set("Idempotency-Key", "cmd-x")
		response := httptest.NewRecorder()
		introMux(t, service, introReaderStub{introduction: owned}, "member-1").ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s %s = %d, want 404", probe.method, probe.path, response.Code)
		}
		if strings.Contains(response.Body.String(), "member-OTHER") {
			t.Fatal("the response disclosed the real owner")
		}
	}
}

func TestWithdrawnConsentIsRefusedAsForbidden(t *testing.T) {
	service := introServiceStub{
		begin: func(context.Context, introapplication.BeginUploadRequest) (introapplication.BeginUploadResult, error) {
			return introapplication.BeginUploadResult{}, introapplication.ErrConsentRequired
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/introductions", strings.NewReader(`{"contentType":"audio/ogg","prompt":"arrival"}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "cmd-1")
	response := httptest.NewRecorder()
	introMux(t, service, introReaderStub{}, "member-1").ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
	if !strings.Contains(response.Body.String(), "consent_required") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestPlaybackGrantIsWhatMakesTheTwentySecondGateReachable(t *testing.T) {
	// Sowing needs twenty seconds of verified listening (FR-202). The gate is
	// keyed on this assetId and was already on the wire; without a way to hear
	// the audio it could never be satisfied, so Sow could never arm.
	introduction := introFixture(t, "member-1")
	request := httptest.NewRequest(http.MethodGet, "/v1/introductions/introduction_1/audio", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	introMux(t, introServiceStub{}, introReaderStub{introduction: introduction}, "member-1").
		ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct{ Data playbackResponse }
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.URL == "" {
		t.Fatal("no play URL; the recording still cannot be heard")
	}
	// The asset id ties the grant to the listening gate that counts against it.
	if envelope.Data.AssetID != introduction.Media().AssetID() {
		t.Fatalf("assetId = %q, want %q", envelope.Data.AssetID, introduction.Media().AssetID())
	}
	if envelope.Data.DurationMs != 42_000 {
		t.Fatalf("durationMs = %d, want 42000", envelope.Data.DurationMs)
	}
}

func TestAWithdrawnRecordingStopsPlayingBeforeErasureCatchesUp(t *testing.T) {
	// Handing out a signed URL for erased bytes would 404 at the bucket, which
	// reads to the member as a broken player rather than a recording they
	// chose to withdraw.
	introduction := introFixture(t, "member-1")
	digest := strings.Repeat("a", 64)
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	revoked, err := introduction.Revoke(
		introdomain.Command{ID: "cmd_revoke", Fingerprint: digest, At: now},
		introduction.Version(),
	)
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	// Revoking moves the data to purge_pending: the bytes are still on disk
	// until the sweep runs, and this is the window that matters. The member
	// has said they want it gone, so playback must stop now rather than when
	// erasure catches up. The domain will not let it be purged outright while
	// the 180-day retention is live — that is RET-01 doing its job.
	if revoked.DataStatus() != introdomain.DataPurgePending {
		t.Fatalf("expected purge_pending after revoke, got %v", revoked.DataStatus())
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/introductions/introduction_1/audio", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	introMux(t, introServiceStub{}, introReaderStub{introduction: revoked}, "member-1").
		ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestAnotherMembersAudioIsNotGranted(t *testing.T) {
	introduction := introFixture(t, "member-OTHER")
	request := httptest.NewRequest(http.MethodGet, "/v1/introductions/introduction_1/audio", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	introMux(t, introServiceStub{}, introReaderStub{introduction: introduction}, "member-1").
		ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestARecordingMustSayWhichQuestionItAnswers(t *testing.T) {
	// A finished Voice of Introduction is all three questions, so the server
	// has to know which one each recording answers. Without this the sowing
	// rung could be earned by three takes of the same answer.
	handler := introMux(t, introServiceStub{}, introReaderStub{}, "member_1")
	for name, body := range map[string]string{
		"absent":  `{"contentType":"audio/ogg"}`,
		"unknown": `{"contentType":"audio/ogg","prompt":"favourite-colour"}`,
		"blank":   `{"contentType":"audio/ogg","prompt":"  "}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/v1/introductions", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer token")
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "cmd_1")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%s prompt: status = %d, want 422", name, response.Code)
		}
		if !strings.Contains(response.Body.String(), `"field":"prompt"`) {
			t.Fatalf("%s prompt: refusal did not name the field: %s", name, response.Body.String())
		}
	}
}
