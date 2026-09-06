package mtn

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
)

var minted = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

// collector stands in for MTN. It records what it was asked and answers the
// way the real API does.
type collector struct {
	tokenCalls   int
	tokenExpires int64
	collectCode  int
	seen         *http.Request
	seenBody     map[string]any
}

func (c *collector) handler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/collection/token/") {
			c.tokenCalls++
			expires := c.tokenExpires
			if expires == 0 {
				expires = 3600
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "token-1", "expires_in": expires,
			})
			return
		}
		c.seen = r
		_ = json.NewDecoder(r.Body).Decode(&c.seenBody)
		code := c.collectCode
		if code == 0 {
			code = http.StatusAccepted
		}
		w.WriteHeader(code)
	})
}

func provider(t *testing.T, c *collector) (*Provider, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(c.handler(t))
	t.Cleanup(server.Close)
	built, err := New(Config{
		BaseURL: server.URL, SubscriptionKey: "sub", APIUser: "user", APIKey: "key",
		TargetEnvironment: "sandbox", CallbackURL: "https://obiara.example/callback",
	}, func() time.Time { return minted })
	if err != nil {
		t.Fatal(err)
	}
	return built, server
}

func request() application.ProviderRequest {
	return application.ProviderRequest{
		RequestRef: "ref-1", PhoneRef: "233200000000",
		AmountPesewas: 12_050, Currency: "GHS",
	}
}

func TestAskingForACollectionSendsWhatTheProviderNeeds(t *testing.T) {
	c := &collector{}
	built, _ := provider(t, c)
	reference, err := built.RequestCollection(context.Background(), request())
	if err != nil {
		t.Fatal(err)
	}
	// The reference it was given, back. The payment context checks this: a
	// provider answering with a different one is not talking about this
	// collection, and treating that as success would attach somebody else's
	// payment to this intent.
	if reference != "ref-1" {
		t.Fatalf("reference = %q", reference)
	}
	if c.seen.Header.Get("X-Reference-Id") != "ref-1" {
		t.Fatalf("reference header = %q", c.seen.Header.Get("X-Reference-Id"))
	}
	if c.seen.Header.Get("X-Target-Environment") != "sandbox" {
		t.Fatal("the wrong environment silently charges nobody")
	}
	if c.seen.Header.Get("Ocp-Apim-Subscription-Key") != "sub" {
		t.Fatal("the subscription key did not travel")
	}
	if c.seen.Header.Get("X-Callback-Url") != "https://obiara.example/callback" {
		t.Fatal("without a callback url nothing ever settles the intent")
	}
}

func TestPesewasBecomeMajorUnitsWithoutFloatingPoint(t *testing.T) {
	// The rest of this codebase counts pesewas; MTN takes a decimal string.
	// Money and floating point do not mix, so the conversion is integer
	// arithmetic and it is checked.
	for pesewas, want := range map[uint64]string{
		12_050: "120.50", 1: "0.01", 100: "1.00", 99: "0.99", 1_000_000: "10000.00",
	} {
		if got := majorUnits(pesewas); got != want {
			t.Fatalf("%d pesewas = %q, want %q", pesewas, got, want)
		}
	}
}

func TestOnlyAcceptedIsASentPrompt(t *testing.T) {
	// 202 means the prompt is on its way. Anything else, a 200 included, is
	// not a prompt this adapter can vouch for — and a member has not paid
	// until they approve one.
	for _, code := range []int{http.StatusOK, http.StatusBadRequest, http.StatusConflict,
		http.StatusInternalServerError} {
		c := &collector{collectCode: code}
		built, _ := provider(t, c)
		if _, err := built.RequestCollection(context.Background(), request()); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("status %d was treated as a sent prompt", code)
		}
	}
}

func TestTheTokenIsReusedUntilItIsNearlySpent(t *testing.T) {
	// Minting one per collection would double every payment's latency and its
	// failure surface.
	c := &collector{}
	built, _ := provider(t, c)
	for range 3 {
		if _, err := built.RequestCollection(context.Background(), request()); err != nil {
			t.Fatal(err)
		}
	}
	if c.tokenCalls != 1 {
		t.Fatalf("minted %d tokens for three collections", c.tokenCalls)
	}
}

func TestATokenThatExpiresTooSoonIsNotUsed(t *testing.T) {
	// Refreshed a minute early, so a token cannot expire between the check
	// and the request it authorizes. One that lives less than that minute is
	// no use at all, and pretending otherwise fails a member's payment for a
	// reason they could never understand.
	c := &collector{tokenExpires: 30}
	built, _ := provider(t, c)
	if _, err := built.RequestCollection(context.Background(), request()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestTheCredentialIsSentAsTheProviderExpects(t *testing.T) {
	var seen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "t", "expires_in": 3600})
	}))
	defer server.Close()

	built, err := New(Config{
		BaseURL: server.URL, SubscriptionKey: "sub", APIUser: "user", APIKey: "key",
		TargetEnvironment: "sandbox",
	}, func() time.Time { return minted })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := built.accessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:key"))
	if seen != want {
		t.Fatalf("authorization = %q", seen)
	}
}

func TestAHalfConfiguredProviderIsNotBuilt(t *testing.T) {
	// Half a credential is a misconfiguration that should read as "off", not
	// as "on and broken" — the same rule the object store follows.
	for name, config := range map[string]Config{
		"no base url":    {SubscriptionKey: "s", APIUser: "u", APIKey: "k", TargetEnvironment: "sandbox"},
		"no key":         {BaseURL: "https://x", SubscriptionKey: "s", APIUser: "u", TargetEnvironment: "sandbox"},
		"no environment": {BaseURL: "https://x", SubscriptionKey: "s", APIUser: "u", APIKey: "k"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(config, nil); !errors.Is(err, ErrConfigured) {
				t.Fatalf("err = %v, want ErrConfigured", err)
			}
		})
	}
}

func TestARequestThatCouldNotBeACollectionIsRefusedBeforeTheNetwork(t *testing.T) {
	c := &collector{}
	built, _ := provider(t, c)
	for name, broken := range map[string]application.ProviderRequest{
		"no reference":   {PhoneRef: "233200000000", AmountPesewas: 100, Currency: "GHS"},
		"no phone":       {RequestRef: "ref-1", AmountPesewas: 100, Currency: "GHS"},
		"nothing to pay": {RequestRef: "ref-1", PhoneRef: "233200000000", Currency: "GHS"},
		// GHS only, deliberately. The intent aggregate is GHS-only too, and a
		// currency mismatch here would charge the right number of the wrong
		// unit.
		"another currency": {RequestRef: "r", PhoneRef: "2332", AmountPesewas: 100, Currency: "USD"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := built.RequestCollection(context.Background(), broken); !errors.Is(err, ErrUnavailable) {
				t.Fatalf("err = %v, want ErrUnavailable", err)
			}
			if c.seen != nil {
				t.Fatal("an impossible collection reached the network")
			}
		})
	}
}
