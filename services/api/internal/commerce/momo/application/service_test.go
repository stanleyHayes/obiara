package application

import (
	"context"
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
	ids.EXPECT().NewID().Return(key(1))
	clock.EXPECT().Now().Return(now)
	s := New(repo, p, ids, clock, secret)
	i, e := s.Create(context.Background(), key(2), key(3), 500, "create")
	if e != nil {
		t.Fatal(e)
	}
	clock.EXPECT().Now().Return(now)
	ids.EXPECT().NewID().Return(key(4))
	p.EXPECT().RequestCollection(gomock.Any(), ProviderRequest{key(4), key(3), 500, "GHS"}).Return(key(4), nil)
	clock.EXPECT().Now().Return(now)
	i, e = s.Confirm(context.Background(), i.ID(), "confirm")
	if e != nil {
		t.Fatal(e)
	}
	cb := Callback{CallbackID: "callback-1", IntentID: i.ID(), ProviderRef: key(5), Success: true, OccurredUnix: now.Unix()}
	cb.Signature = SignCallback(secret, cb)
	clock.EXPECT().Now().Return(now)
	i, e = s.Callback(context.Background(), cb)
	if e != nil || i.State().Status != domain.Succeeded {
		t.Fatal(e)
	}
	cb.Signature = "00"
	if _, e = s.Callback(context.Background(), cb); e == nil {
		t.Fatal("unsigned callback accepted")
	}
}
func TestPhoneRefHMACRedacts(t *testing.T) {
	r, e := PhoneRef([]byte("01234567890123456789012345678901"), "+233201234567")
	if e != nil || r == "+233201234567" || len(r) != 64 {
		t.Fatal(r, e)
	}
}
