package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/domain"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return key(i.n) }

type clock struct{ at time.Time }

func (c clock) Now() time.Time { return c.at }
func TestPrepareConfirmAndAccountUseExactIsolatedFacts(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	now := time.Now().UTC()
	a := NewMockAuthority(ctrl)
	catalog := NewMockCatalog(ctrl)
	v := NewMockConfirmationVerifier(ctrl)
	ledger := NewMockSettlementLedger(ctrl)
	repo := NewMockRepository(ctrl)
	source := &ids{}
	quote := domain.Quote{SKUKey: key(10), Version: 3, Currency: domain.CurrencyUSD, AmountMinor: 1499, ValidUntil: now.Add(time.Hour)}
	a.EXPECT().RequireDiasporaCheckout(ctx, key(9)).Return(nil)
	catalog.EXPECT().CurrentDiasporaQuote(ctx, "diaspora.intent", domain.CurrencyUSD).Return(quote, nil)
	repo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	s := New(a, catalog, v, ledger, repo, source, clock{now})
	prepared, e := s.Prepare(ctx, key(9), "diaspora.intent", domain.CurrencyUSD, "prepare-1")
	if e != nil {
		t.Fatal(e)
	}
	confirmation := ProviderConfirmation{prepared.Checkout.ID(), prepared.RequestRef, key(11), domain.CurrencyUSD, 1499, true, now, key(12)}
	repo.EXPECT().Find(ctx, prepared.Checkout.ID()).Return(prepared.Checkout, nil)
	v.EXPECT().Verify(ctx, confirmation).Return(nil)
	repo.EXPECT().Save(ctx, gomock.Any(), uint64(1), "confirm-1").Return(nil)
	confirmed, e := s.Confirm(ctx, confirmation, "confirm-1")
	if e != nil {
		t.Fatal(e)
	}
	repo.EXPECT().Find(ctx, confirmed.ID()).Return(confirmed, nil)
	catalog.EXPECT().CurrentDiasporaQuote(ctx, key(10), domain.CurrencyUSD).Return(quote, nil)
	ledger.EXPECT().RecordPlatformSale(ctx, PlatformSale{"ledger-1", confirmed.ID(), key(11), domain.CurrencyUSD, 1499}).Return(key(13), nil)
	repo.EXPECT().Save(ctx, gomock.Any(), uint64(2), "ledger-1").Return(nil)
	accounted, e := s.Account(ctx, confirmed.ID(), "ledger-1")
	if e != nil || accounted.State().Status != domain.Accounted {
		t.Fatal(e)
	}
}
func TestMismatchedConfirmationNeverReachesVerifierOrLedger(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	now := time.Now().UTC()
	repo := NewMockRepository(ctrl)
	checkout, _ := domain.Create(key(1), key(2), domain.Quote{SKUKey: key(3), Version: 1, Currency: domain.CurrencyGBP, AmountMinor: 1199, ValidUntil: now.Add(time.Hour)}, key(4), "prepare-1", now)
	repo.EXPECT().Find(ctx, key(1)).Return(checkout, nil)
	s := New(NewMockAuthority(ctrl), NewMockCatalog(ctrl), NewMockConfirmationVerifier(ctrl), NewMockSettlementLedger(ctrl), repo, &ids{}, clock{now})
	if _, e := s.Confirm(ctx, ProviderConfirmation{key(1), key(4), key(5), domain.CurrencyGBP, 1200, true, now, key(6)}, "confirm-1"); e != ErrInvalid {
		t.Fatalf("got %v", e)
	}
}

func TestLedgerPortCannotExpressMemberTransferOrArbitraryLines(t *testing.T) {
	typ := reflect.TypeOf(PlatformSale{})
	got := make([]string, typ.NumField())
	for index := range typ.NumField() {
		got[index] = typ.Field(index).Name
	}
	want := []string{"CommandID", "CheckoutRef", "ProviderRef", "Currency", "AmountMinor"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ledger contract expanded unsafely: %v", got)
	}
}
