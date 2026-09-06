package purchase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	catalogdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
	ledgerapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/application"
	ledgerdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
	membershipdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
	momoapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
	momodomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
)

var now = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

type catalogStub struct {
	sku catalogdomain.SKU
	err error
}

func (s catalogStub) ReadPublished(context.Context, string, uint64) (catalogdomain.SKU, error) {
	return s.sku, s.err
}

type paymentsStub struct {
	intent      momodomain.Intent
	createErr   error
	confirmErr  error
	callbackErr error
	memberKey   string
	phoneRef    string
	amount      uint64
	confirmed   bool
}

func (s *paymentsStub) Create(
	_ context.Context, memberKey, phoneRef string, amount uint64, _ string,
) (momodomain.Intent, error) {
	s.memberKey, s.phoneRef, s.amount = memberKey, phoneRef, amount
	return s.intent, s.createErr
}

func (s *paymentsStub) Confirm(context.Context, string, string) (momodomain.Intent, error) {
	s.confirmed = true
	return s.intent, s.confirmErr
}

func (s *paymentsStub) Callback(
	context.Context, momoapplication.Callback,
) (momodomain.Intent, error) {
	return s.intent, s.callbackErr
}

type passesStub struct {
	granted     bool
	cancelled   bool
	paidThrough time.Time
	grace       time.Duration
	grantErr    error
	cancelErr   error
}

func (s *passesStub) Grant(
	_ context.Context, _, _, _ string, _ uint64,
	paidThrough time.Time, grace time.Duration, _ string,
) (membershipdomain.Pass, error) {
	s.granted = true
	s.paidThrough, s.grace = paidThrough, grace
	return membershipdomain.Pass{}, s.grantErr
}

func (s *passesStub) Cancel(context.Context, string, string) (membershipdomain.Pass, error) {
	s.cancelled = true
	return membershipdomain.Pass{}, s.cancelErr
}

// digestKeyer behaves like the real keyers: a one-way digest, distinct per
// namespace. It actually hashes rather than decorating the input, because a
// stub that carried the value through would let the test that checks nothing
// raw escapes pass or fail for reasons that are about the stub.
type digestKeyer struct{ err error }

func (k digestKeyer) MemberKey(memberID string) (string, error) {
	return k.digest("membership_member", memberID)
}

func (k digestKeyer) PhoneRef(phone string) (string, error) {
	return k.digest("momo_phone", phone)
}

func (k digestKeyer) digest(namespace, value string) (string, error) {
	if k.err != nil {
		return "", k.err
	}
	sum := sha256.Sum256([]byte(namespace + "\x00" + value))
	return hex.EncodeToString(sum[:]), nil
}

type ledgerStub struct {
	recorded bool
	minor    int64
	err      error
}

func (l *ledgerStub) RecordSale(_ context.Context, _ string, minor int64, _ string, _ time.Time) error {
	l.recorded, l.minor = true, minor
	return l.err
}

func membershipSKU(t *testing.T, currency catalogdomain.Currency, minor int64) catalogdomain.SKU {
	t.Helper()
	price, err := catalogdomain.NewPrice(currency, minor)
	if err != nil {
		t.Fatal(err)
	}
	sku, err := catalogdomain.Rehydrate(catalogdomain.State{
		// The key is a slug and the title is a keyed reference, not the other
		// way round: a catalogue that carried readable titles would name what
		// members buy in every log line that mentions a SKU.
		ID: "sku_membership", SKUKey: "membership.monthly", TitleRef: strings.Repeat("a", 64),
		Version: 1, Kind: catalogdomain.KindDigitalService, Price: price,
		Status: catalogdomain.StatusPublished, PublishedAt: now, Revision: 1,
		Events:   []catalogdomain.Event{{Sequence: 1, CommandID: "cmd_1", Action: "created", At: now}},
		Commands: []catalogdomain.Applied{{ID: "cmd_1", Fingerprint: strings.Repeat("b", 64), Revision: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return sku
}

func intentFor(t *testing.T) momodomain.Intent {
	t.Helper()
	// The payment context's ids are all 64-hex digests: an intent id is not a
	// readable reference either.
	intent, err := momodomain.Create(
		strings.Repeat("1", 64), strings.Repeat("2", 64), strings.Repeat("3", 64),
		5000, "cmd_1", now)
	if err != nil {
		t.Fatal(err)
	}
	return intent
}

func service(t *testing.T, catalog Catalog, payments Payments, passes Passes, ledger Ledger) Service {
	t.Helper()
	return New(catalog, payments, passes, digestKeyer{}, ledger, func() time.Time { return now })
}

func TestStartingAPurchasePricesItFromTheCatalog(t *testing.T) {
	// The price is administered, not deployed. Reading it here is what makes
	// changing it an operator action and what a discount code will later
	// reduce.
	payments := &paymentsStub{intent: intentFor(t)}
	started, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if started.AmountPesewas != 5000 {
		t.Fatalf("amount = %d, want the SKU's price", started.AmountPesewas)
	}
	if payments.amount != 5000 {
		t.Fatalf("the intent was opened for %d", payments.amount)
	}
	// The prompt is sent in the same request: the member is standing there,
	// and a second round trip only adds a place to abandon it.
	if !payments.confirmed {
		t.Fatal("the member was never prompted")
	}
}

func TestNothingRawReachesThePaymentContext(t *testing.T) {
	// A raw phone number in a payment row is a contact directory; a raw
	// member id is a purchase history.
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000",
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(payments.phoneRef, "0200000000") {
		t.Fatalf("the phone number was stored raw: %q", payments.phoneRef)
	}
	if strings.Contains(payments.memberKey, "member-1") {
		t.Fatalf("the member id was stored raw: %q", payments.memberKey)
	}
	// And the two are keyed differently, so one cannot be matched against the
	// other across contexts.
	if payments.phoneRef == payments.memberKey {
		t.Fatal("the member and their phone share a digest")
	}
}

func TestOnlyAPublishedGhanaCediMembershipIsPurchasable(t *testing.T) {
	// A price in another currency would be charged as the right number of the
	// wrong unit, and an unpublished SKU is not something anybody offered.
	for name, catalog := range map[string]Catalog{
		"not published": catalogStub{err: errors.New("unavailable")},
		"another currency": catalogStub{
			sku: membershipSKU(t, catalogdomain.CurrencyUSD, 5000),
		},
	} {
		t.Run(name, func(t *testing.T) {
			payments := &paymentsStub{intent: intentFor(t)}
			if _, err := service(t, catalog, payments, &passesStub{}, &ledgerStub{}).
				Start(context.Background(), StartCommand{
					CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku",
					SKUVersion: 1, Phone: "0200000000",
				}); !errors.Is(err, ErrNotPurchasable) {
				t.Fatalf("err = %v, want ErrNotPurchasable", err)
			}
			if payments.confirmed {
				t.Fatal("a member was prompted for something they cannot buy")
			}
		})
	}
}

func TestNoPassIsGrantedUntilTheProviderSaysSo(t *testing.T) {
	// A member who never approves the prompt has not paid. Starting a
	// purchase grants nothing.
	passes := &passesStub{}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		&paymentsStub{intent: intentFor(t)}, passes, &ledgerStub{},
	).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000",
	}); err != nil {
		t.Fatal(err)
	}
	if passes.granted {
		t.Fatal("a pass was granted before anybody paid")
	}
}

func TestASuccessfulCallbackGrantsThePassAndBooksTheMoney(t *testing.T) {
	passes, ledger := &passesStub{}, &ledgerStub{}
	err := service(t, catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, ledger).
		Settle(context.Background(), momoapplication.Callback{
			CallbackID: "cb_1", IntentID: "intent_1", ProviderRef: "ref-1", Success: true,
		})
	if err != nil {
		t.Fatal(err)
	}
	if !passes.granted {
		t.Fatal("a paid member got no pass")
	}
	// Thirty days, seven days grace: what one payment buys.
	if !passes.paidThrough.Equal(now.Add(Period)) {
		t.Fatalf("paid through %v", passes.paidThrough)
	}
	if passes.grace != Grace {
		t.Fatalf("grace = %v", passes.grace)
	}
	// Revenue is visible in the book rather than implied by a pass appearing.
	if !ledger.recorded || ledger.minor != 5000 {
		t.Fatalf("ledger recorded = %v, minor = %d", ledger.recorded, ledger.minor)
	}
}

func TestAFailedPaymentGrantsNothing(t *testing.T) {
	passes, ledger := &passesStub{}, &ledgerStub{}
	if err := service(t, catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, ledger).
		Settle(context.Background(), momoapplication.Callback{
			CallbackID: "cb_1", IntentID: "intent_1", ProviderRef: "ref-1", Success: false,
		}); err != nil {
		t.Fatal(err)
	}
	if passes.granted {
		t.Fatal("a failed payment granted a pass")
	}
	if ledger.recorded {
		t.Fatal("a failed payment was booked as revenue")
	}
}

func TestAReversalCancelsThePassRatherThanRemovingIt(t *testing.T) {
	// The trail should show granted-then-cancelled rather than a pass that
	// quietly vanished.
	passes := &passesStub{}
	if err := service(t, catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, &ledgerStub{}).
		Settle(context.Background(), momoapplication.Callback{
			CallbackID: "cb_2", IntentID: "intent_1", ProviderRef: "ref-1", Success: false,
		}); err != nil {
		t.Fatal(err)
	}
	if !passes.cancelled {
		t.Fatal("a reversal left the pass standing")
	}
}

func TestAnUnverifiedCallbackSettlesNothing(t *testing.T) {
	// The payment context checks the signature. If it refuses, nothing here
	// grants anything — a forged callback must not be able to mint a pass.
	passes := &passesStub{}
	payments := &paymentsStub{intent: intentFor(t), callbackErr: errors.New("bad signature")}
	if err := service(t, catalogStub{}, payments, passes, &ledgerStub{}).
		Settle(context.Background(), momoapplication.Callback{
			CallbackID: "cb_1", IntentID: "intent_1", Success: true,
		}); err == nil {
		t.Fatal("an unverified callback was accepted")
	}
	if passes.granted {
		t.Fatal("a forged callback minted a pass")
	}
}

func TestABookkeepingFailureDoesNotUnGrantAPaidPass(t *testing.T) {
	// The member paid. A posting that did not land is a gap for
	// reconciliation to find, not a reason to take away what they bought.
	passes := &passesStub{}
	ledger := &ledgerStub{err: errors.New("ledger unavailable")}
	if err := service(t, catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, ledger).
		Settle(context.Background(), momoapplication.Callback{
			CallbackID: "cb_1", IntentID: "intent_1", ProviderRef: "ref-1", Success: true,
		}); err != nil {
		t.Fatalf("a bookkeeping failure refused a paid member: %v", err)
	}
	if !passes.granted {
		t.Fatal("the member paid and got nothing")
	}
}

func TestAnUncomposedPurchaseRefuses(t *testing.T) {
	if _, err := (Service{}).Start(context.Background(), StartCommand{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if err := (Service{}).Settle(context.Background(), momoapplication.Callback{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// postedLines records what reached the double-entry book.
type postedLines struct {
	command ledgerapplication.PostCommand
	err     error
}

func (p *postedLines) Post(
	_ context.Context, command ledgerapplication.PostCommand,
) (ledgerdomain.Posting, error) {
	p.command = command
	return ledgerdomain.Posting{}, p.err
}

func TestAMembershipSaleBalances(t *testing.T) {
	// Two lines, because that is what a sale is: cash held at the provider
	// goes up and membership revenue goes up with it. One side would balance
	// nowhere.
	poster := &postedLines{}
	if err := NewSaleBook(poster, "system").RecordSale(
		context.Background(), "intent-1", 5000, "GHS", now,
	); err != nil {
		t.Fatal(err)
	}
	lines := poster.command.Lines
	if len(lines) != 2 {
		t.Fatalf("%d lines, want a debit and a credit", len(lines))
	}
	var debits, credits int64
	for _, line := range lines {
		switch line.Side {
		case ledgerdomain.SideDebit:
			debits += line.Minor
		case ledgerdomain.SideCredit:
			credits += line.Minor
		}
	}
	if debits != credits || debits != 5000 {
		t.Fatalf("debits %d, credits %d", debits, credits)
	}
	if poster.command.Purpose != ledgerdomain.PurposeSaleSettlement {
		t.Fatalf("purpose = %q", poster.command.Purpose)
	}
	// Keyed by the intent, so a retried callback posts once.
	if poster.command.CommandID != "membership_sale:intent-1" {
		t.Fatalf("command id = %q", poster.command.CommandID)
	}
}

func TestNoAccountLineNamesAMember(t *testing.T) {
	// A ledger carrying a member key per line would be a purchase history in
	// the accounts. The reference is already enough to trace one sale.
	poster := &postedLines{}
	_ = NewSaleBook(poster, "system").RecordSale(
		context.Background(), "intent-1", 5000, "GHS", now)
	for _, line := range poster.command.Lines {
		if strings.Contains(line.AccountID, "member") && line.AccountID != "membership_revenue" {
			t.Fatalf("an account named a member: %q", line.AccountID)
		}
	}
}

func TestNothingIsPostedForNothing(t *testing.T) {
	poster := &postedLines{}
	if err := NewSaleBook(poster, "system").RecordSale(
		context.Background(), "intent-1", 0, "GHS", now,
	); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if poster.command.CommandID != "" {
		t.Fatal("a zero sale reached the book")
	}
}
