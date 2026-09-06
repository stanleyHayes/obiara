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
	promotionapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/application"
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
	payer       momoapplication.Payer
	findErr     error
}

func (s *paymentsStub) Create(
	_ context.Context, memberKey, phoneRef string, amount uint64, _ string,
) (momodomain.Intent, error) {
	s.memberKey, s.phoneRef, s.amount = memberKey, phoneRef, amount
	return s.intent, s.createErr
}

func (s *paymentsStub) Confirm(
	_ context.Context, _, _ string, payer momoapplication.Payer,
) (momodomain.Intent, error) {
	s.confirmed = true
	s.payer = payer
	return s.intent, s.confirmErr
}

func (s *paymentsStub) Settle(
	context.Context, string, string, string, bool,
) (momodomain.Intent, error) {
	return s.intent, s.callbackErr
}

func (s *paymentsStub) Find(context.Context, string) (momodomain.Intent, error) {
	return s.intent, s.findErr
}

// orderBook remembers what a collection was opened to buy.
type orderBook struct {
	recorded Order
	err      error
	findErr  error
	order    Order
}

func (o *orderBook) Record(_ context.Context, order Order) error {
	o.recorded = order
	if o.order.IntentID == "" {
		o.order = order
	}
	return o.err
}

func (o *orderBook) Find(context.Context, string) (Order, error) {
	if o.findErr != nil {
		return Order{}, o.findErr
	}
	return o.order, nil
}

func membershipOrder() Order {
	return Order{
		IntentID: strings.Repeat("1", 64), SKUKey: "membership.monthly", SKUVersion: 1,
		MemberID: "member-1", AmountPesewas: 5000,
	}
}

// receipts is where a payment receipt goes.
type receipts struct {
	email string
	err   error
}

func (r receipts) Email(context.Context, string) (string, error) { return r.email, r.err }

type passesStub struct {
	memberKey   string
	passID      string
	receiptRef  string
	passVersion uint64
	granted     bool
	cancelled   bool
	paidThrough time.Time
	grace       time.Duration
	grantErr    error
	cancelErr   error
}

func (s *passesStub) Grant(
	_ context.Context, memberKey, passID, receiptRef string, passVersion uint64,
	paidThrough time.Time, grace time.Duration, _ string,
) (membershipdomain.Pass, error) {
	s.granted = true
	s.memberKey, s.passID, s.receiptRef, s.passVersion = memberKey, passID, receiptRef, passVersion
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
	return New(catalog, payments, passes, digestKeyer{}, ledger, func() time.Time { return now }).
		WithOrders(&orderBook{order: membershipOrder()}).
		WithMembers(receipts{email: "member@example.test"})
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
		SKUVersion: 1, Phone: "0200000000", Network: "mtn",
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
		SKUVersion: 1, Phone: "0200000000", Network: "mtn",
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
					SKUVersion: 1, Phone: "0200000000", Network: "mtn",
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
		SKUVersion: 1, Phone: "0200000000", Network: "mtn",
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
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: true,
			AmountPesewas: 5000, Currency: "GHS",
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
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: false,
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
		Settle(context.Background(), Outcome{
			CallbackID: "cb_2", Reference: strings.Repeat("1", 64), Success: false,
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
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: true,
			AmountPesewas: 5000, Currency: "GHS",
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
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: true,
			AmountPesewas: 5000, Currency: "GHS",
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
	if err := (Service{}).Settle(context.Background(), Outcome{}); !errors.Is(err, ErrUnavailable) {
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

// codeStub applies a fixed discount.
type codeStub struct {
	applied promotionapplication.Applied
	err     error
	price   int64
	sku     string
	called  bool
}

func (s *codeStub) Apply(
	_ context.Context, _, _, skuID string, priceMinor int64, _ string,
) (promotionapplication.Applied, error) {
	s.called, s.price, s.sku = true, priceMinor, skuID
	return s.applied, s.err
}

func TestACodeComesOffWhatTheMemberIsCharged(t *testing.T) {
	// The discount is applied to the amount collected, not reported and then
	// ignored: a member shown 25 cedis and charged 50 would have been lied to.
	codes := &codeStub{applied: promotionapplication.Applied{Code: "ASHESI26", DiscountMinor: 2500}}
	payments := &paymentsStub{intent: intentFor(t)}
	started, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).WithDiscounts(codes).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ashesi26",
	})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 2500 {
		t.Fatalf("collected %d, want the discounted price", payments.amount)
	}
	if started.AmountPesewas != 2500 || started.DiscountPesewas != 2500 {
		t.Fatalf("started = %#v", started)
	}
	// The code is priced against the SKU it names, not against whatever the
	// client sent.
	if codes.sku != "sku_membership" || codes.price != 5000 {
		t.Fatalf("priced %d against %q", codes.price, codes.sku)
	}
}

func TestAPurchaseWithoutDiscountsComposedStillWorks(t *testing.T) {
	// A deployment with no promotion context charges everybody full price.
	// That is a working product, not a broken one.
	payments := &paymentsStub{intent: intentFor(t)}
	started, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ashesi26",
	})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 5000 || started.DiscountPesewas != 0 {
		t.Fatalf("started = %#v, collected %d", started, payments.amount)
	}
}

func TestACodeThatTakesTheWholePriceIsNotAPayment(t *testing.T) {
	// Nothing here can collect zero, and a free membership is a decision
	// somebody made rather than a purchase to push through a payment rail.
	codes := &codeStub{applied: promotionapplication.Applied{Code: "FREE", DiscountMinor: 5000}}
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).WithDiscounts(codes).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "FREE",
	}); !errors.Is(err, ErrNotPurchasable) {
		t.Fatalf("err = %v, want ErrNotPurchasable", err)
	}
	if payments.confirmed {
		t.Fatal("a member was prompted to pay nothing")
	}
}

func TestADiscountThatCannotBeEstablishedStopsThePurchase(t *testing.T) {
	// Not the same as a code that does not apply. If the promotion context
	// cannot answer, charging full price would charge a member who believes
	// they have a discount, and charging the discount would give one nothing
	// recorded.
	codes := &codeStub{err: errors.New("promotions unavailable")}
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).WithDiscounts(codes).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ASHESI26",
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if payments.confirmed {
		t.Fatal("a member was prompted at a price nobody could establish")
	}
}

func TestASettlementForTheWrongAmountGrantsNothing(t *testing.T) {
	// A signed webhook is authentic, and authentic is not the same as
	// correct. An outcome reporting less than the collection was opened for
	// would otherwise buy a whole month for whatever the payer felt like
	// sending.
	passes, ledger := &passesStub{}, &ledgerStub{}
	err := service(t, catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, ledger).
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: true,
			AmountPesewas: 1, Currency: "GHS",
		})
	if !errors.Is(err, ErrWrongAmount) {
		t.Fatalf("err = %v, want ErrWrongAmount", err)
	}
	if passes.granted {
		t.Fatal("a short payment bought a membership")
	}
	if ledger.recorded {
		t.Fatal("a short payment was booked at full value")
	}
}

func TestASettlementInTheWrongCurrencyGrantsNothing(t *testing.T) {
	// The right number of the wrong unit. 5000 naira is not 5000 pesewas.
	passes := &passesStub{}
	err := service(t, catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, &ledgerStub{}).
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: true,
			AmountPesewas: 5000, Currency: "NGN",
		})
	if !errors.Is(err, ErrWrongAmount) {
		t.Fatalf("err = %v, want ErrWrongAmount", err)
	}
	if passes.granted {
		t.Fatal("a foreign-currency payment bought a membership")
	}
}

func TestTheRawPhoneReachesTheProcessorAndTheDigestDoesNot(t *testing.T) {
	// The defect this closes: the collection stored an HMAC of the number and
	// then handed that digest to the processor as the number to dial. A
	// payment processor cannot charge a 64-character hex string.
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000", Network: "mtn",
	}); err != nil {
		t.Fatal(err)
	}
	if payments.payer.Phone != "0200000000" {
		t.Fatalf("the processor was given %q to dial", payments.payer.Phone)
	}
	if payments.payer.Network != "mtn" {
		t.Fatalf("network = %q", payments.payer.Network)
	}
	// And a receipt address, which a processor will not open a collection
	// without.
	if payments.payer.Email != "member@example.test" {
		t.Fatalf("email = %q", payments.payer.Email)
	}
	// The stored reference is still a digest: the raw number is used and not
	// written down.
	if payments.phoneRef == "0200000000" || len(payments.phoneRef) != 64 {
		t.Fatalf("the intent stored %q", payments.phoneRef)
	}
}

func TestNoReceiptAddressMeansNoCollection(t *testing.T) {
	// A processor will not open one without somewhere to send a receipt, and
	// inventing an address would send a member's receipt into a hole.
	payments := &paymentsStub{intent: intentFor(t)}
	bare := New(catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, digestKeyer{}, &ledgerStub{}, func() time.Time { return now }).
		WithMembers(receipts{email: ""})
	if _, err := bare.Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000", Network: "mtn",
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if payments.confirmed {
		t.Fatal("a collection was opened with nowhere to send a receipt")
	}
}

func TestAPurchaseWithoutANetworkIsRefused(t *testing.T) {
	// The processor needs to know which network the number is on and cannot
	// reliably infer it.
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, &ledgerStub{},
	).Start(context.Background(), StartCommand{
		CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
		SKUVersion: 1, Phone: "0200000000",
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if payments.confirmed {
		t.Fatal("a collection was opened without a network")
	}
}

func TestAPassIsGrantedForTheProductAndNotForThePayment(t *testing.T) {
	// The bug this closes would have taken money and granted nothing for most
	// payments. A pass is granted for a product: the membership context wants
	// a slug it can recognise, and a payment id is a hex string. Passing the
	// payment id meant roughly five purchases in eight failed validation
	// after the member had already paid.
	passes := &passesStub{}
	orders := &orderBook{order: membershipOrder()}
	err := New(catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, digestKeyer{},
		&ledgerStub{}, func() time.Time { return now }).
		WithOrders(orders).WithMembers(receipts{email: "member@example.test"}).
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: true,
			AmountPesewas: 5000, Currency: "GHS",
		})
	if err != nil {
		t.Fatal(err)
	}
	if passes.passID != "membership.monthly" {
		t.Fatalf("granted pass %q, want the product's key", passes.passID)
	}
	if passes.passVersion != 1 {
		t.Fatalf("pass version = %d, want the version that was sold", passes.passVersion)
	}
	// The payment is the receipt, which is the opaque reference it should be.
	if passes.receiptRef != strings.Repeat("1", 64) {
		t.Fatalf("receipt = %q", passes.receiptRef)
	}
}

func TestAPurchaseRecordsWhatItWasOpenedToBuy(t *testing.T) {
	// Settlement happens later and reads this. Without it, granting on a
	// webhook is guessing.
	orders := &orderBook{}
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := New(
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, digestKeyer{}, &ledgerStub{}, func() time.Time { return now },
	).WithOrders(orders).WithMembers(receipts{email: "member@example.test"}).
		Start(context.Background(), StartCommand{
			CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
			SKUVersion: 1, Phone: "0200000000", Network: "mtn",
		}); err != nil {
		t.Fatal(err)
	}
	if orders.recorded.SKUKey != "membership.monthly" || orders.recorded.SKUVersion != 1 {
		t.Fatalf("recorded %#v", orders.recorded)
	}
	if orders.recorded.AmountPesewas != 5000 {
		t.Fatalf("recorded amount %d", orders.recorded.AmountPesewas)
	}
}

func TestAnOrderThatCannotBeWrittenStopsThePurchase(t *testing.T) {
	// A payment nobody can attribute to a product takes a member's money and
	// leaves settlement guessing.
	orders := &orderBook{err: errors.New("mongo down")}
	if _, err := New(
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		&paymentsStub{intent: intentFor(t)}, &passesStub{}, digestKeyer{},
		&ledgerStub{}, func() time.Time { return now },
	).WithOrders(orders).WithMembers(receipts{email: "member@example.test"}).
		Start(context.Background(), StartCommand{
			CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
			SKUVersion: 1, Phone: "0200000000", Network: "mtn",
		}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestSettlementWithNoOrderGrantsNothing(t *testing.T) {
	passes := &passesStub{}
	orders := &orderBook{findErr: ErrOrderNotFound}
	if err := New(catalogStub{}, &paymentsStub{intent: intentFor(t)}, passes, digestKeyer{},
		&ledgerStub{}, func() time.Time { return now }).
		WithOrders(orders).WithMembers(receipts{email: "x@example.test"}).
		Settle(context.Background(), Outcome{
			CallbackID: "cb_1", Reference: strings.Repeat("1", 64), Success: true,
			AmountPesewas: 5000, Currency: "GHS",
		}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if passes.granted {
		t.Fatal("a pass was granted for a purchase nothing recorded")
	}
}

// fundStub is an organization's balance.
type fundStub struct {
	drawn    bool
	err      error
	seatRef  string
	amount   int64
	issuerID string
}

func (f *fundStub) Draw(
	_ context.Context, organizationID, seatRef string, amountPesewas int64,
) (bool, error) {
	f.issuerID, f.seatRef, f.amount = organizationID, seatRef, amountPesewas
	return f.drawn, f.err
}

func sponsoredCode() *codeStub {
	return &codeStub{applied: promotionapplication.Applied{
		Code: "ASHESISEATS", DiscountMinor: 5000, Sponsored: true, IssuerID: "org_1",
	}}
}

func TestASponsoredSeatIsPaidByTheOrganizationAndGrantedAtOnce(t *testing.T) {
	// The money arrived when the organization deposited it. There is no
	// prompt to send and nothing to wait for.
	passes, ledger := &passesStub{}, &ledgerStub{}
	fund := &fundStub{drawn: true}
	payments := &paymentsStub{intent: intentFor(t)}
	started, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, passes, ledger,
	).WithDiscounts(sponsoredCode()).WithSponsors(fund).
		Start(context.Background(), StartCommand{
			CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
			SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ASHESISEATS",
		})
	if err != nil {
		t.Fatal(err)
	}
	// The member is charged nothing and is never prompted.
	if started.AmountPesewas != 0 || started.Status != "sponsored" {
		t.Fatalf("started = %#v", started)
	}
	if payments.confirmed {
		t.Fatal("a member was prompted to pay for a seat somebody else bought")
	}
	// The organization is charged the full price.
	if fund.amount != 5000 || fund.issuerID != "org_1" {
		t.Fatalf("drew %d from %q", fund.amount, fund.issuerID)
	}
	// And the pass is granted here, for the product.
	if !passes.granted || passes.passID != "membership.monthly" {
		t.Fatalf("granted %q", passes.passID)
	}
	// Revenue in full: a sponsorship is not a discount and must not be booked
	// as one.
	if !ledger.recorded || ledger.minor != 5000 {
		t.Fatalf("booked %d", ledger.minor)
	}
}

func TestAnOrganizationThatCannotCoverTheSeatDoesNotBlockTheMember(t *testing.T) {
	// Refusing the sponsorship rather than the purchase: the member can still
	// buy their own membership.
	passes := &passesStub{}
	fund := &fundStub{drawn: false}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		&paymentsStub{intent: intentFor(t)}, passes, &ledgerStub{},
	).WithDiscounts(sponsoredCode()).WithSponsors(fund).
		Start(context.Background(), StartCommand{
			CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
			SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ASHESISEATS",
		}); !errors.Is(err, ErrSponsorshipUnavailable) {
		t.Fatalf("err = %v, want ErrSponsorshipUnavailable", err)
	}
	if passes.granted {
		t.Fatal("a seat nobody paid for was granted")
	}
}

func TestARetriedSponsoredPurchaseDrawsTheSameSeat(t *testing.T) {
	// The seat reference is the purchase command, so an organization is not
	// charged twice for one member.
	first, second := &fundStub{drawn: true}, &fundStub{drawn: true}
	for _, fund := range []*fundStub{first, second} {
		if _, err := service(t,
			catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
			&paymentsStub{intent: intentFor(t)}, &passesStub{}, &ledgerStub{},
		).WithDiscounts(sponsoredCode()).WithSponsors(fund).
			Start(context.Background(), StartCommand{
				CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
				SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ASHESISEATS",
			}); err != nil {
			t.Fatal(err)
		}
	}
	if first.seatRef == "" || first.seatRef != second.seatRef {
		t.Fatalf("%q then %q", first.seatRef, second.seatRef)
	}
}

func TestASponsoredCodeWithNoSponsorshipComposedIsRefused(t *testing.T) {
	// The code covers the whole price with nobody paying it. Refused rather
	// than given away.
	passes := &passesStub{}
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := service(t,
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, passes, &ledgerStub{},
	).WithDiscounts(sponsoredCode()).
		Start(context.Background(), StartCommand{
			CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
			SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ASHESISEATS",
		}); !errors.Is(err, ErrNotPurchasable) {
		t.Fatalf("err = %v, want ErrNotPurchasable", err)
	}
	if passes.granted || payments.confirmed {
		t.Fatal("a membership was given away")
	}
}

func TestNoMemberIsChargedBeforeTheOrderIsWrittenDown(t *testing.T) {
	// The window this closes: a member charged, and nothing knowing what for.
	// Settlement looks the order up, cannot find it, and the pass is never
	// granted — so the money is gone and nobody gets anything.
	//
	// An order for a collection that is never paid is harmless by comparison.
	orders := &orderBook{err: errors.New("mongo down")}
	payments := &paymentsStub{intent: intentFor(t)}
	if _, err := New(
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		payments, &passesStub{}, digestKeyer{}, &ledgerStub{}, func() time.Time { return now },
	).WithOrders(orders).WithMembers(receipts{email: "member@example.test"}).
		Start(context.Background(), StartCommand{
			CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
			SKUVersion: 1, Phone: "0200000000", Network: "mtn",
		}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if payments.confirmed {
		t.Fatal("a member was prompted to pay for something nothing recorded")
	}
}

func TestNoFundIsDrawnBeforeTheSeatIsWrittenDown(t *testing.T) {
	// An organization's balance debited with no order against it is money
	// gone with nothing saying what it bought.
	orders := &orderBook{err: errors.New("mongo down")}
	fund := &fundStub{drawn: true}
	if _, err := New(
		catalogStub{sku: membershipSKU(t, catalogdomain.CurrencyGHS, 5000)},
		&paymentsStub{intent: intentFor(t)}, &passesStub{}, digestKeyer{},
		&ledgerStub{}, func() time.Time { return now },
	).WithOrders(orders).WithSponsors(fund).WithDiscounts(sponsoredCode()).
		WithMembers(receipts{email: "member@example.test"}).
		Start(context.Background(), StartCommand{
			CommandID: "cmd_1", MemberID: "member-1", SKUID: "sku_membership",
			SKUVersion: 1, Phone: "0200000000", Network: "mtn", Code: "ASHESISEATS",
		}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if fund.seatRef != "" {
		t.Fatal("an organization was charged for a seat nothing recorded")
	}
}
