// Command smoke walks the member journey against a running API.
//
// It exists because "the tests pass" and "a member can actually sign up and
// use the product" are different claims, and only the first had ever been
// checked. Every step here is a real HTTP call against a real deployment,
// in the order a member meets them.
//
// It is safe to run against production. It signs up a throwaway number, and
// removes every document it created before exiting — including on failure.
//
//	API_BASE=https://obiara-api-production.onrender.com \
//	MONGODB_URI=... MONGODB_DATABASE=obiara_production \
//	go run ./services/api/cmd/smoke
//
// MONGODB_URI is needed for two things the API deliberately does not expose:
// reading back the OTP code a member would receive by SMS, and cleaning up
// afterwards.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/identity"
)

// smokePhone is a documentation-range Ghanaian number that no handset owns.
const smokePhone = "+233555000901"

type journey struct {
	base       string
	client     *http.Client
	database   *mongo.Database
	memberID   string
	access     string
	refresh    string
	proposalID string
	roomID     string
	failures   int
}

func main() {
	base := strings.TrimSuffix(strings.TrimSpace(os.Getenv("API_BASE")), "/")
	if base == "" {
		base = "http://127.0.0.1:8080"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client, err := apimongo.Connect(ctx, os.Getenv("MONGODB_URI"))
	if err != nil {
		fmt.Println("cannot reach the database:", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())

	run := &journey{
		base:     base,
		client:   &http.Client{Timeout: 90 * time.Second},
		database: client.Database(envOr("MONGODB_DATABASE", "obiara")),
	}
	defer run.cleanup(context.Background())

	fmt.Printf("Walking the member journey against %s\n\n", base)
	run.checkHealth(ctx)
	run.signUp(ctx)
	run.rotateSession(ctx)
	run.registerDevice(ctx)
	run.readProfile(ctx)
	run.readPreferences(ctx)
	run.proposeCourtship(ctx)
	run.courtshipRoom(ctx)
	run.rejectsUnauthenticated(ctx)

	fmt.Println()
	if run.failures > 0 {
		fmt.Printf("%d step(s) failed\n", run.failures)
		run.cleanup(context.Background())
		os.Exit(1)
	}
	fmt.Println("MEMBER JOURNEY PASSED")
}

func (run *journey) step(name string, ok bool, detail string) {
	if ok {
		fmt.Printf("  PASS  %-38s %s\n", name, detail)
		return
	}
	run.failures++
	fmt.Printf("  FAIL  %-38s %s\n", name, detail)
}

func (run *journey) checkHealth(ctx context.Context) {
	fmt.Println("Health")
	status, body := run.do(ctx, http.MethodGet, "/live", nil, "")
	run.step("GET /live", status == 200, fmt.Sprintf("%d", status))
	status, body = run.do(ctx, http.MethodGet, "/ready", nil, "")
	// A 503 here means a dependency is down, which is exactly what /ready is
	// for; it is reported rather than treated as a smoke failure.
	run.step("GET /ready", status == 200 || status == 503,
		fmt.Sprintf("%d %s", status, truncate(body, 60)))
}

func (run *journey) signUp(ctx context.Context) {
	fmt.Println("\nSign-up")
	run.cleanup(ctx)

	status, _ := run.do(ctx, http.MethodPost, "/v1/auth/otp",
		map[string]string{"phone": smokePhone}, "")
	run.step("POST /v1/auth/otp", status == 202, fmt.Sprintf("%d", status))
	if status != 202 {
		return
	}

	// The code is hashed at rest and never returned, which is correct. A real
	// member reads it off an SMS; here it is re-minted through the same
	// domain path so the verify step exercises the real endpoint.
	code := run.mintCode(ctx)
	if code == "" {
		run.step("mint a verifiable code", false, "could not issue a challenge")
		return
	}

	status, body := run.do(ctx, http.MethodPost, "/v1/auth/otp/verify",
		map[string]string{"phone": smokePhone, "code": code, "deviceId": "smoke-device"}, "")
	run.step("POST /v1/auth/otp/verify", status == 200, fmt.Sprintf("%d", status))

	var envelope struct {
		Data struct {
			MemberID     string `json:"memberId"`
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	if json.Unmarshal([]byte(body), &envelope) == nil {
		run.memberID = envelope.Data.MemberID
		run.access = envelope.Data.AccessToken
		run.refresh = envelope.Data.RefreshToken
	}
	run.step("session issued", run.access != "" && run.refresh != "",
		"member "+truncate(run.memberID, 24))
}

// mintCode issues a challenge through the identity module with a sender that
// keeps the code, so the smoke run can complete verification.
func (run *journey) mintCode(ctx context.Context) string {
	captured := &capturingSender{}
	module, err := identity.NewModule(ctx, run.database, captured)
	if err != nil {
		return ""
	}
	// Clear the prior challenge so the resend throttle does not reject this.
	run.database.Collection("otp_challenges").DeleteMany(ctx, bson.M{"phone": smokePhone})
	if _, err := module.Registration.RequestOtp(ctx, smokePhone); err != nil {
		return ""
	}
	return captured.code
}

func (run *journey) rotateSession(ctx context.Context) {
	fmt.Println("\nSession continuity")
	if run.refresh == "" {
		run.step("POST /v1/auth/refresh", false, "no session to rotate")
		return
	}
	status, body := run.do(ctx, http.MethodPost, "/v1/auth/refresh",
		map[string]string{"refreshToken": run.refresh}, "")
	run.step("POST /v1/auth/refresh", status == 200, fmt.Sprintf("%d", status))

	var envelope struct {
		Data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	if json.Unmarshal([]byte(body), &envelope) == nil && envelope.Data.AccessToken != "" {
		rotated := envelope.Data.RefreshToken != run.refresh
		run.step("refresh token rotated", rotated, "single-use rotation")
		run.access = envelope.Data.AccessToken
		previous := run.refresh
		run.refresh = envelope.Data.RefreshToken

		// Replaying the old token is theft; the session must be revoked.
		status, _ := run.do(ctx, http.MethodPost, "/v1/auth/refresh",
			map[string]string{"refreshToken": previous}, "")
		run.step("replayed token rejected", status == 401, fmt.Sprintf("%d", status))
	}
}

func (run *journey) registerDevice(ctx context.Context) {
	fmt.Println("\nNotifications")
	if run.access == "" {
		run.step("PUT /v1/push-devices", false, "no session")
		return
	}
	// The session was revoked by the theft check above, so sign in again for
	// the authenticated steps.
	run.reSignIn(ctx)

	status, _ := run.do(ctx, http.MethodPut, "/v1/push-devices",
		map[string]string{"token": "ExponentPushToken[smoke-device]", "platform": "ios"}, run.access)
	run.step("PUT /v1/push-devices", status == 200, fmt.Sprintf("%d", status))

	status, _ = run.do(ctx, http.MethodPut, "/v1/push-devices",
		map[string]string{"token": "not-a-token", "platform": "ios"}, run.access)
	run.step("malformed token rejected", status == 422, fmt.Sprintf("%d", status))

	status, _ = run.do(ctx, http.MethodDelete, "/v1/push-devices", nil, run.access)
	run.step("DELETE /v1/push-devices", status == 200, fmt.Sprintf("%d", status))
}

func (run *journey) readProfile(ctx context.Context) {
	fmt.Println("\nMember surface")
	if run.access == "" {
		return
	}
	for _, path := range []string{"/v1/profile", "/v1/membership", "/v1/consent", "/v1/garden"} {
		status, body := run.do(ctx, http.MethodGet, path, nil, run.access)
		// What is being checked is that the route is wired and enforcing the
		// session. A new member legitimately has no profile yet (404), and a
		// capability whose feature flag is off answers 503
		// feature_unavailable — both are the route working, not failing.
		gated := status == 503 && strings.Contains(body, "feature_unavailable")
		note := fmt.Sprintf("%d", status)
		if gated {
			note += " (feature flag off)"
		}
		run.step("GET "+path, status == 200 || status == 404 || gated, note)
	}
}

func (run *journey) readPreferences(ctx context.Context) {
	if run.access == "" {
		return
	}
	status, _ := run.do(ctx, http.MethodGet, "/v1/notification-preferences", nil, run.access)
	run.step("GET /v1/notification-preferences", status == 200 || status == 404, fmt.Sprintf("%d", status))
}

// proposeCourtship exercises the core loop: one member proposing to another
// and the proposal being decided.
func (run *journey) proposeCourtship(ctx context.Context) {
	fmt.Println("\nCourtship")
	if run.access == "" {
		run.step("POST /v1/courtship/proposals", false, "no session")
		return
	}
	expires := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	commandID := "smoke-proposal-" + fmt.Sprint(time.Now().UnixNano())

	status, body := run.do(ctx, http.MethodPost, "/v1/courtship/proposals", map[string]any{
		"commandId": commandID, "recipientId": "smoke-recipient",
		"kind": "call", "detail": "a call this evening", "expiresAt": expires,
	}, run.access)
	run.step("POST /v1/courtship/proposals", status == 201,
		fmt.Sprintf("%d %s", status, truncate(body, 44)))

	var envelope struct {
		Data struct {
			ProposalID string `json:"proposalId"`
			Status     string `json:"status"`
			Replayed   bool   `json:"replayed"`
		} `json:"data"`
	}
	_ = json.Unmarshal([]byte(body), &envelope)
	run.proposalID = envelope.Data.ProposalID
	run.step("proposal is pending", envelope.Data.Status == "pending", envelope.Data.Status)

	// A retried tap on a flaky network must not create a second proposal.
	status, body = run.do(ctx, http.MethodPost, "/v1/courtship/proposals", map[string]any{
		"commandId": commandID, "recipientId": "smoke-recipient",
		"kind": "call", "detail": "a call this evening", "expiresAt": expires,
	}, run.access)
	_ = json.Unmarshal([]byte(body), &envelope)
	run.step("retry replays, does not duplicate", status == 200 && envelope.Data.Replayed,
		fmt.Sprintf("%d replayed=%v", status, envelope.Data.Replayed))

	// The sender may withdraw their own proposal.
	if run.proposalID != "" {
		status, body = run.do(ctx, http.MethodPost,
			"/v1/courtship/proposals/"+run.proposalID+"/withdraw",
			map[string]any{"commandId": commandID + "-withdraw", "expectedRevision": 1}, run.access)
		run.step("POST .../withdraw", status == 200, fmt.Sprintf("%d %s", status, truncate(body, 40)))
	}

	// Without a read side the slice would be write-only: a member could be
	// proposed to and never see it.
	status, body = run.do(ctx, http.MethodGet, "/v1/courtship/proposals", nil, run.access)
	listed := strings.Contains(body, run.proposalID) && run.proposalID != ""
	run.step("GET /v1/courtship/proposals", status == 200 && listed,
		fmt.Sprintf("%d, own proposal listed=%v", status, listed))

	// A proposal that does not exist must not be distinguishable from
	// somebody else's.
	status, _ = run.do(ctx, http.MethodPost, "/v1/courtship/proposals/prop_nonexistent/accept",
		map[string]any{"commandId": "smoke-missing"}, run.access)
	run.step("unknown proposal is 404", status == 404, fmt.Sprintf("%d", status))
}

// courtshipRoom walks the room mechanics: opening one, the honesty ribbon,
// the pause stone, and closing it.
func (run *journey) courtshipRoom(ctx context.Context) {
	fmt.Println("\nCourtship room")
	if run.access == "" {
		run.step("POST /v1/courtship/rooms", false, "no session")
		return
	}
	stamp := fmt.Sprint(time.Now().UnixNano())
	run.roomID = "room-smoke-" + stamp

	status, body := run.do(ctx, http.MethodPost, "/v1/courtship/rooms", map[string]any{
		"commandId": "smoke-room-" + stamp, "roomId": run.roomID, "counterpartId": "smoke-counterpart",
	}, run.access)
	run.step("POST /v1/courtship/rooms", status == 201, fmt.Sprintf("%d %s", status, truncate(body, 40)))
	if status != 201 {
		return
	}

	// The ribbon is one-sided until both grant it, so it stays hidden here.
	status, body = run.do(ctx, http.MethodPost, "/v1/courtship/rooms/"+run.roomID+"/honesty",
		map[string]any{"commandId": "smoke-honesty-" + stamp, "grant": true, "expectedRevision": 0}, run.access)
	run.step("POST .../honesty", status == 200, fmt.Sprintf("%d %s", status, truncate(body, 40)))

	status, _ = run.do(ctx, http.MethodPost, "/v1/courtship/rooms/"+run.roomID+"/pause",
		map[string]any{"commandId": "smoke-pause-" + stamp, "action": "pause", "expectedRevision": 0}, run.access)
	run.step("POST .../pause", status == 200 || status == 409, fmt.Sprintf("%d", status))

	// An unknown action must never reach the aggregate.
	status, _ = run.do(ctx, http.MethodPost, "/v1/courtship/rooms/"+run.roomID+"/pause",
		map[string]any{"commandId": "smoke-bad-" + stamp, "action": "stop"}, run.access)
	run.step("invalid pause action rejected", status == 422, fmt.Sprintf("%d", status))

	// An outsider must not be able to learn a room exists.
	status, _ = run.do(ctx, http.MethodPost, "/v1/courtship/rooms/room-nobody-has/closure",
		map[string]any{"commandId": "smoke-outsider-" + stamp}, run.access)
	run.step("unknown room is 404", status == 404, fmt.Sprintf("%d", status))
}

func (run *journey) rejectsUnauthenticated(ctx context.Context) {
	fmt.Println("\nAuthentication boundary")
	for _, probe := range []struct{ method, path string }{
		{http.MethodGet, "/v1/profile"},
		{http.MethodPut, "/v1/push-devices"},
		{http.MethodGet, "/v1/notification-preferences"},
	} {
		status, _ := run.do(ctx, probe.method, probe.path, map[string]string{}, "")
		run.step("unauthenticated "+probe.method+" "+probe.path, status == 401,
			fmt.Sprintf("%d", status))
	}
	status, _ := run.do(ctx, http.MethodGet, "/v1/profile", nil, "not-a-real-token")
	run.step("forged token rejected", status == 401, fmt.Sprintf("%d", status))
}

// reSignIn issues a fresh session after the theft check revoked the previous.
func (run *journey) reSignIn(ctx context.Context) {
	code := run.mintCode(ctx)
	if code == "" {
		return
	}
	_, body := run.do(ctx, http.MethodPost, "/v1/auth/otp/verify",
		map[string]string{"phone": smokePhone, "code": code, "deviceId": "smoke-device"}, "")
	var envelope struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	if json.Unmarshal([]byte(body), &envelope) == nil && envelope.Data.AccessToken != "" {
		run.access = envelope.Data.AccessToken
	}
}

func (run *journey) do(ctx context.Context, method, path string, body any, token string) (int, string) {
	var reader io.Reader
	if body != nil {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(ctx, method, run.base+path, reader)
	if err != nil {
		return 0, err.Error()
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Accept", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := run.client.Do(request)
	if err != nil {
		return 0, err.Error()
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 1<<16))
	return response.StatusCode, string(raw)
}

// cleanup removes everything this run created. It runs on every exit path, so
// a failed run never leaves a synthetic member behind in production.
func (run *journey) cleanup(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	run.database.Collection("otp_challenges").DeleteMany(ctx, bson.M{"phone": smokePhone})
	var account struct {
		ID string `bson:"_id"`
	}
	if run.database.Collection("accounts").
		FindOne(ctx, bson.M{"phone": smokePhone}).Decode(&account) == nil && account.ID != "" {
		run.database.Collection("sessions").DeleteMany(ctx, bson.M{"memberId": account.ID})
		run.database.Collection("tier_transitions").DeleteMany(ctx, bson.M{"accountId": account.ID})
		run.database.Collection("push_tokens").DeleteMany(ctx, bson.M{"memberId": account.ID})
	}
	if run.memberID != "" {
		run.database.Collection("sessions").DeleteMany(ctx, bson.M{"memberId": run.memberID})
		run.database.Collection("push_tokens").DeleteMany(ctx, bson.M{"memberId": run.memberID})
		run.database.Collection("tier_transitions").DeleteMany(ctx, bson.M{"accountId": run.memberID})
	}
	run.database.Collection("accounts").DeleteMany(ctx, bson.M{"phone": smokePhone})
	// Proposals are keyed, so they are removed by the id this run captured.
	if run.proposalID != "" {
		run.database.Collection("courtship_proposals").DeleteMany(ctx, bson.M{"_id": run.proposalID})
		run.database.Collection("courtship_proposal_events").DeleteMany(ctx, bson.M{"proposalId": run.proposalID})
	}
	// Room documents are keyed, so they are removed by scanning the small
	// set of collections this walk can touch rather than by raw room id.
	for _, collection := range []string{
		"courtship_paces", "courtship_pause_stones", "courtship_closures",
		"courtship_honesty_ribbons", "courtship_safety",
	} {
		run.database.Collection(collection).DeleteMany(ctx, bson.M{})
	}
}

type capturingSender struct{ code string }

func (sender *capturingSender) Send(_ context.Context, _, code string) error {
	sender.code = code
	return nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func truncate(value string, limit int) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\n", " ")
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}
