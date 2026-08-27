package failover

import (
	"context"
	"errors"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type stubSender struct {
	err   error
	calls int
	code  string
}

func (sender *stubSender) Send(_ context.Context, _ domain.Contact, code string) error {
	sender.calls++
	sender.code = code
	return sender.err
}

func TestFirstWorkingRungWins(t *testing.T) {
	primary := &stubSender{}
	fallback := &stubSender{}
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	sender, err := NewSender(nil, primary, fallback)
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	if err := sender.Send(context.Background(), contact, "123456"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if primary.calls != 1 {
		t.Errorf("primary calls = %d, want 1", primary.calls)
	}
	// Falling through after a success would double-send and double-bill.
	if fallback.calls != 0 {
		t.Errorf("fallback calls = %d, want 0", fallback.calls)
	}
	if primary.code != "123456" {
		t.Errorf("primary received code %q", primary.code)
	}
}

func TestLadderFallsThroughToTheNextRung(t *testing.T) {
	smsDown := errors.New("arkesel unavailable")
	primary := &stubSender{err: smsDown}
	fallback := &stubSender{}
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	sender, err := NewSender(nil, primary, fallback)
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	if err := sender.Send(context.Background(), contact, "123456"); err != nil {
		t.Fatalf("Send returned %v; the fallback rung accepted the message", err)
	}
	if fallback.calls != 1 {
		t.Errorf("fallback calls = %d, want 1", fallback.calls)
	}
}

func TestExhaustedLadderReportsEveryFailure(t *testing.T) {
	first := errors.New("arkesel unavailable")
	second := errors.New("whatsapp unavailable")
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	sender, err := NewSender(nil, &stubSender{err: first}, &stubSender{err: second})
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}

	sendErr := sender.Send(context.Background(), contact, "123456")
	if sendErr == nil {
		t.Fatal("Send succeeded with every rung failing")
	}
	// Operators triaging a dark channel need both causes, not just the last.
	if !errors.Is(sendErr, first) || !errors.Is(sendErr, second) {
		t.Errorf("error %v should join both rung failures", sendErr)
	}
}

func TestCancelledContextStopsTheLadder(t *testing.T) {
	primary := &stubSender{}
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")
	sender, err := NewSender(nil, primary)
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := sender.Send(ctx, contact, "123456"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Send error = %v, want context.Canceled", err)
	}
	if primary.calls != 0 {
		t.Errorf("primary was called %d times on an abandoned request", primary.calls)
	}
}

func TestObserverSeesEveryRungOutcome(t *testing.T) {
	type outcome struct {
		index  int
		failed bool
	}
	var seen []outcome
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233544919953")

	sender, err := NewSender(func(_ context.Context, index int, err error) {
		seen = append(seen, outcome{index: index, failed: err != nil})
	}, &stubSender{err: errors.New("down")}, &stubSender{})
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	if err := sender.Send(context.Background(), contact, "123456"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	want := []outcome{{index: 0, failed: true}, {index: 1, failed: false}}
	if len(seen) != len(want) {
		t.Fatalf("observer saw %v, want %v", seen, want)
	}
	for i, expected := range want {
		if seen[i] != expected {
			t.Fatalf("observer saw %v, want %v", seen, want)
		}
	}
}

func TestNewSenderRejectsAnEmptyLadder(t *testing.T) {
	if _, err := NewSender(nil); !errors.Is(err, ErrNoSenders) {
		t.Fatalf("NewSender() error = %v, want ErrNoSenders", err)
	}
	// A ladder of nothing but nil rungs is equally unusable.
	var absent application.OtpSender
	if _, err := NewSender(nil, absent); !errors.Is(err, ErrNoSenders) {
		t.Fatalf("NewSender(nil rung) error = %v, want ErrNoSenders", err)
	}
}
