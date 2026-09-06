package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sowmedia "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/adapters/outbound/media"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

type recorderStub struct {
	recording sowmedia.Recording
	err       error
	ownerID   string
	size      int64
	checksum  string
	duration  time.Duration
	calls     int
}

func (stub *recorderStub) Open(
	_ context.Context, ownerID, _ string, sizeBytes int64, checksum string, duration time.Duration,
) (sowmedia.Recording, error) {
	stub.calls++
	stub.ownerID, stub.size, stub.checksum, stub.duration = ownerID, sizeBytes, checksum, duration
	return stub.recording, stub.err
}

func recordingRequest(t *testing.T, stub *recorderStub, body string, gate MemberGate) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterSowRecordingRoutes(mux, stub, sessionStub{memberID: "member_1"}, gate)
	request := httptest.NewRequest(http.MethodPost, "/v1/seed/sows/recordings", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	return response
}

func TestOpeningASowRecordingHandsBackAGrantAndTheRefTheSowCarries(t *testing.T) {
	stub := &recorderStub{recording: sowmedia.Recording{
		AssetID:   "sow_asset_1",
		URL:       "https://bucket.example/sows/1?sig=put",
		ExpiresAt: time.Date(2026, time.September, 6, 12, 15, 0, 0, time.UTC),
	}}
	response := recordingRequest(t, stub,
		`{"contentType":"audio/ogg","sizeBytes":240000,"checksum":"`+
			strings.Repeat("A", 64)+`","durationMs":60000}`, sowingGate())

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	// The recording is the session's own. A member cannot open one as
	// somebody else and then sow it as theirs.
	if stub.ownerID != "member_1" {
		t.Fatalf("owner = %q", stub.ownerID)
	}
	if stub.size != 240_000 || stub.duration != time.Minute {
		t.Fatalf("described as %d bytes, %v", stub.size, stub.duration)
	}
	// Lower-cased on the way through: the digest is hex, and one differing
	// only in case would be refused as malformed.
	if stub.checksum != strings.Repeat("a", 64) {
		t.Fatalf("checksum = %q", stub.checksum)
	}
	body := response.Body.String()
	if !strings.Contains(body, "sow_asset_1") || !strings.Contains(body, "sig=put") {
		t.Fatalf("body = %s", body)
	}
}

func TestAnImplausibleSowRecordingIsRefusedAsInput(t *testing.T) {
	stub := &recorderStub{err: sowmedia.ErrNotRecordable}
	response := recordingRequest(t, stub,
		`{"contentType":"audio/ogg","sizeBytes":40000,"checksum":"`+
			strings.Repeat("a", 64)+`","durationMs":90000}`, sowingGate())

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "durationMs") {
		t.Fatalf("the member was not told which field: %s", response.Body.String())
	}
}

func TestAMemberWhoCannotSowCannotRecordForOne(t *testing.T) {
	// Behind the sowing rung, because somebody who cannot sow has nothing to
	// record for — and an upload grant is a write on the bucket.
	stub := &recorderStub{}
	response := recordingRequest(t, stub,
		`{"contentType":"audio/ogg","sizeBytes":240000,"checksum":"`+
			strings.Repeat("a", 64)+`","durationMs":60000}`, gateAt(identityTierUnverified()))

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
	if stub.calls != 0 {
		t.Fatal("an unverified member's request reached the recorder")
	}
}

func TestAHalfSentRecordingIsNotToldItIsSomebodyElses(t *testing.T) {
	// Their own recording, still uploading. A member told "that is not yours"
	// would go looking for a problem that is not there.
	sow := &sowStub{err: sowapplication.ErrMediaNotArrived}
	response := sowRequest(t, sow,
		`{"targetId":"member_2","body":"hi","mediaRefs":["mine"],"confirmed":true}`, "cmd-1")
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	body := strings.ToLower(response.Body.String())
	if !strings.Contains(body, "recording_not_arrived") {
		t.Fatalf("body = %s", response.Body.String())
	}
	if strings.Contains(body, "yourself") || strings.Contains(body, "made yourself") {
		t.Fatalf("a half-sent recording was called somebody else's: %s", response.Body.String())
	}
}
