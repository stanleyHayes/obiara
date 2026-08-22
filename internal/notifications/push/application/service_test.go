package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/internal/notifications/push/domain"
)

type stubRegistry struct {
	tokens    []string
	forgotten []string
	err       error
}

func (r *stubRegistry) Register(context.Context, domain.Registration) error { return r.err }
func (r *stubRegistry) TokensFor(context.Context, string) ([]string, error) {
	return r.tokens, r.err
}
func (r *stubRegistry) Forget(_ context.Context, tokens []string) error {
	r.forgotten = append(r.forgotten, tokens...)
	return nil
}
func (r *stubRegistry) ForgetMember(context.Context, string) error { return nil }

type stubSender struct {
	sentTo []string
	copy   domain.Copy
	dead   []string
	err    error
}

func (s *stubSender) Send(_ context.Context, tokens []string, copy domain.Copy, _ string) ([]string, error) {
	s.sentTo = append(s.sentTo, tokens...)
	s.copy = copy
	return s.dead, s.err
}

func newService(registry *stubRegistry, sender *stubSender) Service {
	return NewService(registry, sender, func() time.Time {
		return time.Date(2026, 8, 22, 6, 0, 0, 0, time.UTC)
	})
}

func TestDeliverPushesToEveryDevice(t *testing.T) {
	registry := &stubRegistry{tokens: []string{"ExponentPushToken[a]", "ExponentPushToken[b]"}}
	sender := &stubSender{}

	ref, err := newService(registry, sender).Deliver(context.Background(), "mem_1", "morning_greeting", "2026-08-22")
	if err != nil {
		t.Fatalf("Deliver: %v", err)
	}
	if ref == "" {
		t.Error("Deliver returned no provider reference")
	}
	if len(sender.sentTo) != 2 {
		t.Errorf("sent to %d devices, want 2", len(sender.sentTo))
	}
	if sender.copy.Body == "" {
		t.Error("no copy was sent")
	}
}

// TestNoDevicesFallsThrough keeps a member without the app installed from
// losing the notification: the rung reports it cannot serve, and the ladder
// drops to the in-app inbox.
func TestNoDevicesFallsThrough(t *testing.T) {
	_, err := newService(&stubRegistry{}, &stubSender{}).
		Deliver(context.Background(), "mem_1", "morning_greeting", "ref")
	if !errors.Is(err, ErrNoDevices) {
		t.Fatalf("Deliver = %v, want ErrNoDevices", err)
	}
}

// TestDeadTokensArePruned matters because Expo rejects whole batches that
// carry too many stale tokens: a registry that never forgets decays into a
// channel that cannot deliver at all.
func TestDeadTokensArePruned(t *testing.T) {
	registry := &stubRegistry{tokens: []string{"ExponentPushToken[a]", "ExponentPushToken[b]"}}
	sender := &stubSender{dead: []string{"ExponentPushToken[b]"}}

	if _, err := newService(registry, sender).
		Deliver(context.Background(), "mem_1", "morning_greeting", "ref"); err != nil {
		t.Fatalf("Deliver: %v", err)
	}
	if len(registry.forgotten) != 1 || registry.forgotten[0] != "ExponentPushToken[b]" {
		t.Errorf("forgotten = %v, want the dead token pruned", registry.forgotten)
	}
}

func TestUnknownTemplateStillDelivers(t *testing.T) {
	registry := &stubRegistry{tokens: []string{"ExponentPushToken[a]"}}
	sender := &stubSender{}

	if _, err := newService(registry, sender).
		Deliver(context.Background(), "mem_1", "a_template_nobody_mapped", "ref"); err != nil {
		t.Fatalf("Deliver: %v", err)
	}
	if sender.copy.Body != domain.FallbackCopy().Body {
		t.Errorf("copy = %q, want the safe fallback", sender.copy.Body)
	}
}

func TestRegisterValidates(t *testing.T) {
	service := newService(&stubRegistry{}, &stubSender{})
	if err := service.Register(context.Background(), "mem_1", "nonsense", domain.PlatformIOS); !errors.Is(err, domain.ErrTokenInvalid) {
		t.Errorf("Register = %v, want ErrTokenInvalid", err)
	}
	if err := service.Register(context.Background(), "mem_1", "ExponentPushToken[ok]", domain.PlatformAndroid); err != nil {
		t.Errorf("Register = %v, want nil", err)
	}
}
