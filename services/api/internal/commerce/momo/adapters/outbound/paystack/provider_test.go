package paystack

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
)

const secret = "sk_test_0123456789abcdef0123456789abcdef"

// processor stands in for Paystack.
type processor struct {
	status   int
	body     string
	seen     *http.Request
	seenBody map[string]any
}

func (p *processor) serve(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.seen = r
		_ = json.NewDecoder(r.Body).Decode(&p.seenBody)
		code := p.status
		if code == 0 {
			code = http.StatusOK
		}
		body := p.body
		if body == "" {
			body = `{"status":true,"message":"Charge attempted","data":` +
				`{"reference":"` + reference() + `","status":"pay_offline",` +
				`"display_text":"Please complete authorization on your phone"}}`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func reference() string { return strings.Repeat("1", 64) }

func charge() application.ProviderRequest {
	return application.ProviderRequest{
		RequestRef: reference(), Phone: "0551234987", Email: "member@example.test",
		Network: "mtn", AmountPesewas: 12_050, Currency: "GHS",
	}
}

func provider(t *testing.T, p *processor) *Provider {
	t.Helper()
	built, err := New(Config{BaseURL: p.serve(t).URL, SecretKey: secret})
	if err != nil {
		t.Fatal(err)
	}
	return built
}

func TestAChargeSendsWhatPaystackNeeds(t *testing.T) {
	p := &processor{}
	got, err := provider(t, p).RequestCollection(context.Background(), charge())
	if err != nil {
		t.Fatal(err)
	}
	if got != reference() {
		t.Fatalf("reference = %q", got)
	}
	if p.seen.Header.Get("Authorization") != "Bearer "+secret {
		t.Fatalf("authorization = %q", p.seen.Header.Get("Authorization"))
	}
	// Amounts go in the subunit — pesewas — which is what this codebase
	// already counts in, so there is nothing to convert and nothing to round.
	if p.seenBody["amount"] != float64(12_050) {
		t.Fatalf("amount = %v, want pesewas unconverted", p.seenBody["amount"])
	}
	if p.seenBody["currency"] != "GHS" || p.seenBody["reference"] != reference() {
		t.Fatalf("body = %#v", p.seenBody)
	}
	money, _ := p.seenBody["mobile_money"].(map[string]any)
	// The real number, not a digest. A processor cannot dial an HMAC.
	if money["phone"] != "0551234987" {
		t.Fatalf("phone = %v", money["phone"])
	}
	if money["provider"] != "mtn" {
		t.Fatalf("provider = %v", money["provider"])
	}
}

func TestANetworkPaystackDoesNotKnowIsRefusedBeforeTheNetwork(t *testing.T) {
	// Refused here rather than becoming a charge Paystack rejects after the
	// member has already been told something is happening.
	p := &processor{}
	request := charge()
	request.Network = "some-network"
	if _, err := provider(t, p).RequestCollection(context.Background(), request); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if p.seen != nil {
		t.Fatal("an unknown network reached Paystack")
	}
}

func TestGhanasNetworksMapToPaystacksCodes(t *testing.T) {
	for name, want := range map[string]string{
		"mtn": "mtn", "MTN": "mtn",
		"vodafone": "vod", "Telecel": "vod",
		"airteltigo": "atl", "AT": "atl",
	} {
		code, ok := Network(name)
		if !ok || code != want {
			t.Fatalf("%q mapped to %q (%v), want %q", name, code, ok, want)
		}
	}
	if _, ok := Network("glo"); ok {
		t.Fatal("a network Paystack does not serve in Ghana was accepted")
	}
}

func TestOnlyAPromptOrAnInstantChargeCounts(t *testing.T) {
	// send_otp and send_pin are card flows this product does not offer.
	// Treating one as a sent prompt leaves a member waiting for something
	// that will never arrive.
	for _, status := range []string{"send_otp", "send_pin", "send_phone", "failed", ""} {
		p := &processor{body: `{"status":true,"message":"x","data":{"reference":"` +
			reference() + `","status":"` + status + `"}}`}
		if _, err := provider(t, p).RequestCollection(context.Background(), charge()); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("status %q was treated as a sent prompt", status)
		}
	}
	for _, status := range []string{"pay_offline", "success"} {
		p := &processor{body: `{"status":true,"message":"x","data":{"reference":"` +
			reference() + `","status":"` + status + `"}}`}
		if _, err := provider(t, p).RequestCollection(context.Background(), charge()); err != nil {
			t.Fatalf("status %q was refused: %v", status, err)
		}
	}
}

func TestAnAnswerAboutADifferentCollectionIsNotAnAnswer(t *testing.T) {
	p := &processor{body: `{"status":true,"message":"x","data":{"reference":"somebody-else",` +
		`"status":"pay_offline"}}`}
	if _, err := provider(t, p).RequestCollection(context.Background(), charge()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestPaystackSayingNoIsNotAPrompt(t *testing.T) {
	// status:false is Paystack's own refusal, even under a 200.
	p := &processor{body: `{"status":false,"message":"Invalid key","data":{}}`}
	if _, err := provider(t, p).RequestCollection(context.Background(), charge()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	for _, code := range []int{http.StatusUnauthorized, http.StatusBadRequest, http.StatusInternalServerError} {
		p := &processor{status: code}
		if _, err := provider(t, p).RequestCollection(context.Background(), charge()); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("status %d was treated as a sent prompt", code)
		}
	}
}

// signed is what Paystack actually sends: an HMAC-SHA512 of the exact bytes,
// hex, keyed with the secret key.
func signed(body string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

const successBody = `{"event":"charge.success","data":{"id":59214,"domain":"live",` +
	`"status":"success","reference":"ref-1","amount":12050,"currency":"GHS",` +
	`"channel":"mobile_money","gateway_response":"Approved"}}`

func TestAWebhookIsVerifiedOverExactlyTheBytesReceived(t *testing.T) {
	event, err := VerifyWebhook(secret, []byte(successBody), signed(successBody))
	if err != nil {
		t.Fatal(err)
	}
	if !event.Succeeded() {
		t.Fatalf("event = %#v", event)
	}
	if event.Reference != "ref-1" || event.AmountPesewas != 12_050 || event.Currency != "GHS" {
		t.Fatalf("event = %#v", event)
	}
}

func TestAWebhookWhoseBytesChangedIsRefused(t *testing.T) {
	// The signature covers the body, so a single altered byte — an amount, a
	// reference — must not verify. This is the only thing between a stranger
	// and a free membership.
	tampered := strings.Replace(successBody, `"amount":12050`, `"amount":1`, 1)
	if _, err := VerifyWebhook(secret, []byte(tampered), signed(successBody)); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("err = %v, want ErrBadSignature", err)
	}
	if _, err := VerifyWebhook(secret, []byte(successBody), signed(tampered)); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("err = %v, want ErrBadSignature", err)
	}
}

func TestAWebhookSignedWithTheWrongKeyIsRefused(t *testing.T) {
	other := hmac.New(sha512.New, []byte("sk_test_someone_elses_key"))
	other.Write([]byte(successBody))
	if _, err := VerifyWebhook(
		secret, []byte(successBody), hex.EncodeToString(other.Sum(nil)),
	); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("err = %v, want ErrBadSignature", err)
	}
	if _, err := VerifyWebhook("", []byte(successBody), signed(successBody)); !errors.Is(err, ErrConfigured) {
		t.Fatal("an unconfigured secret verified something")
	}
	for _, signature := range []string{"", "not hex", strings.Repeat("0", 128)} {
		if _, err := VerifyWebhook(secret, []byte(successBody), signature); !errors.Is(err, ErrBadSignature) {
			t.Fatalf("signature %q was accepted", signature)
		}
	}
}

func TestAWebhookCarryingAFieldThisBuildHasNotHeardOfStillVerifies(t *testing.T) {
	// A webhook that refused an unknown field would stop settling payments
	// the first time Paystack extended its event.
	extended := `{"event":"charge.success","data":{"reference":"ref-1","status":"success",` +
		`"amount":12050,"currency":"GHS","something_new":{"nested":true}},"invented":42}`
	event, err := VerifyWebhook(secret, []byte(extended), signed(extended))
	if err != nil {
		t.Fatalf("an extended event was refused: %v", err)
	}
	if !event.Succeeded() || event.Reference != "ref-1" {
		t.Fatalf("event = %#v", event)
	}
}

func TestAnEventThatIsNotASuccessDoesNotClaimToBeOne(t *testing.T) {
	// Both the event name and the transaction status. Taking either alone
	// accepts an event that is not the one it looks like.
	failed := `{"event":"charge.success","data":{"reference":"ref-1","status":"failed",` +
		`"amount":12050,"currency":"GHS"}}`
	event, err := VerifyWebhook(secret, []byte(failed), signed(failed))
	if err != nil {
		t.Fatal(err)
	}
	if event.Succeeded() {
		t.Fatal("a failed charge reported success")
	}

	other := `{"event":"transfer.success","data":{"reference":"ref-1","status":"success",` +
		`"amount":12050,"currency":"GHS"}}`
	event, err = VerifyWebhook(secret, []byte(other), signed(other))
	if err != nil {
		t.Fatal(err)
	}
	if event.Succeeded() {
		t.Fatal("a transfer was read as a member's payment")
	}
}

func TestAnUnreadableBodyIsRefusedAfterItIsVerified(t *testing.T) {
	// Verified and still unusable. The order matters: it is authenticated
	// first, so this is a real Paystack request this build cannot parse
	// rather than a stranger's.
	for _, body := range []string{`not json`, `{}`, `{"event":"charge.success","data":{}}`} {
		if _, err := VerifyWebhook(secret, []byte(body), signed(body)); !errors.Is(err, ErrBadEvent) {
			t.Fatalf("body %q gave %v, want ErrBadEvent", body, err)
		}
	}
}

func TestAHalfConfiguredProviderIsNotBuilt(t *testing.T) {
	for name, config := range map[string]Config{
		"no base url": {SecretKey: secret},
		"no key":      {BaseURL: "https://api.paystack.co"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(config); !errors.Is(err, ErrConfigured) {
				t.Fatalf("err = %v, want ErrConfigured", err)
			}
		})
	}
}
