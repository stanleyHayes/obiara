package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

type memory struct{ i domain.Intent }

func (m *memory) Create(_ context.Context, i domain.Intent) error         { m.i = i; return nil }
func (m *memory) Find(_ context.Context, _ string) (domain.Intent, error) { return m.i, nil }
func (m *memory) Save(_ context.Context, i domain.Intent, _ uint64, _ string) error {
	m.i = i
	return nil
}
func TestConfirmProviderAndSignedCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	p := NewMockProvider(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	repo := &memory{}
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	// The intent is opened against the digest of the number it will prompt,
	// which is what Confirm checks the raw number against.
	phoneRef, e := PhoneRef(secret, "+233201234567")
	if e != nil {
		t.Fatal(e)
	}
	ids.EXPECT().NewID().Return(key(1))
	clock.EXPECT().Now().Return(now)
	s := New(repo, p, ids, clock, secret)
	i, e := s.Create(context.Background(), key(2), phoneRef, 500, "create")
	if e != nil {
		t.Fatal(e)
	}
	clock.EXPECT().Now().Return(now)
	// The processor is given the intent's own id as its reference, and the
	// RAW phone rather than the stored digest — a digest cannot be dialled.
	phone := "+233201234567"
	p.EXPECT().RequestCollection(gomock.Any(), ProviderRequest{
		RequestRef: i.ID(), Phone: phone, Email: "member@example.test", Network: "mtn",
		AmountPesewas: 500, Currency: "GHS",
	}).Return(i.ID(), nil)
	clock.EXPECT().Now().Return(now)
	i, e = s.Confirm(context.Background(), i.ID(), "confirm", Payer{
		Phone: phone, Email: "member@example.test", Network: "mtn",
	})
	if e != nil {
		t.Fatal(e)
	}
	clock.EXPECT().Now().Return(now)
	i, e = s.Settle(context.Background(), "callback-1", i.ID(), i.ID(), true)
	if e != nil || i.State().Status != domain.Succeeded {
		t.Fatal(e)
	}
}

func TestConfirmRefusesANumberTheIntentWasNotOpenedFor(t *testing.T) {
	// The intent stores only a digest of the phone, so the raw number is
	// supplied at collection time and checked against it. Without that check
	// a caller could open an intent for one member and have the prompt sent
	// to a different phone entirely.
	ctrl := gomock.NewController(t)
	p := NewMockProvider(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	repo := &memory{}
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

	mine, e := PhoneRef(secret, "+233201234567")
	if e != nil {
		t.Fatal(e)
	}
	ids.EXPECT().NewID().Return(key(1))
	clock.EXPECT().Now().Return(now)
	s := New(repo, p, ids, clock, secret)
	i, e := s.Create(context.Background(), key(2), mine, 500, "create")
	if e != nil {
		t.Fatal(e)
	}
	// No RequestCollection expectation: somebody else's number must never
	// reach the processor.
	if _, e = s.Confirm(context.Background(), i.ID(), "confirm", Payer{
		Phone: "+233209999999", Email: "member@example.test", Network: "mtn",
	}); !errors.Is(e, ErrInvalid) {
		t.Fatalf("e = %v, want ErrInvalid", e)
	}
}
func TestPhoneRefHMACRedacts(t *testing.T) {
	r, e := PhoneRef([]byte("01234567890123456789012345678901"), "+233201234567")
	if e != nil || r == "+233201234567" || len(r) != 64 {
		t.Fatal(r, e)
	}
}
