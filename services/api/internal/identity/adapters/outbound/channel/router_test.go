package channel

import (
	"context"
	"errors"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type stubSender struct {
	channel domain.Channel
	called  bool
	err     error
}

func (s *stubSender) Send(ctx context.Context, contact domain.Contact, code string) error {
	s.called = true
	if contact.Channel() != s.channel {
		return errors.New("wrong channel for this sender")
	}
	return s.err
}

func TestRouterSendsSMSToSMSSender(t *testing.T) {
	smsSender := &stubSender{channel: domain.ChannelSMS}
	emailSender := &stubSender{channel: domain.ChannelEmail}

	router, err := NewRouter(map[domain.Channel]application.OtpSender{
		domain.ChannelSMS:   smsSender,
		domain.ChannelEmail: emailSender,
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	smsContact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	if err := router.Send(context.Background(), smsContact, "123456"); err != nil {
		t.Fatalf("Send SMS: %v", err)
	}

	if !smsSender.called {
		t.Fatal("SMS sender was not called")
	}
	if emailSender.called {
		t.Fatal("Email sender should not be called for SMS contact")
	}
}

func TestRouterSendsEmailToEmailSender(t *testing.T) {
	smsSender := &stubSender{channel: domain.ChannelSMS}
	emailSender := &stubSender{channel: domain.ChannelEmail}

	router, err := NewRouter(map[domain.Channel]application.OtpSender{
		domain.ChannelSMS:   smsSender,
		domain.ChannelEmail: emailSender,
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	emailContact := domain.ReconstituteContact(domain.ChannelEmail, "user@example.com")
	if err := router.Send(context.Background(), emailContact, "123456"); err != nil {
		t.Fatalf("Send email: %v", err)
	}

	if !emailSender.called {
		t.Fatal("Email sender was not called")
	}
	if smsSender.called {
		t.Fatal("SMS sender should not be called for email contact")
	}
}

func TestRouterRejectsUnavailableChannel(t *testing.T) {
	smsSender := &stubSender{channel: domain.ChannelSMS}

	router, err := NewRouter(map[domain.Channel]application.OtpSender{
		domain.ChannelSMS: smsSender,
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	emailContact := domain.ReconstituteContact(domain.ChannelEmail, "user@example.com")
	err = router.Send(context.Background(), emailContact, "123456")
	if !errors.Is(err, ErrChannelUnavailable) {
		t.Fatalf("Send email with unavailable channel = %v, want ErrChannelUnavailable", err)
	}

	if smsSender.called {
		t.Fatal("No sender should be called for unavailable channel")
	}
}

func TestRouterSupports(t *testing.T) {
	router, err := NewRouter(map[domain.Channel]application.OtpSender{
		domain.ChannelSMS:   &stubSender{channel: domain.ChannelSMS},
		domain.ChannelEmail: &stubSender{channel: domain.ChannelEmail},
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	if !router.Supports(domain.ChannelSMS) {
		t.Fatal("Router should support SMS channel")
	}
	if !router.Supports(domain.ChannelEmail) {
		t.Fatal("Router should support email channel")
	}
	if router.Supports("unknown") {
		t.Fatal("Router should not support unknown channel")
	}
}

func TestRouterChannels(t *testing.T) {
	router, err := NewRouter(map[domain.Channel]application.OtpSender{
		domain.ChannelSMS:   &stubSender{channel: domain.ChannelSMS},
		domain.ChannelEmail: &stubSender{channel: domain.ChannelEmail},
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	channels := router.Channels()
	if len(channels) != 2 {
		t.Fatalf("Channels() = %v, want 2 channels", len(channels))
	}

	// Verify stable order
	channels1 := router.Channels()
	channels2 := router.Channels()
	for i := range channels1 {
		if channels1[i] != channels2[i] {
			t.Fatalf("Channels() order not stable")
		}
	}
}

func TestRouterRejectsEmptyMap(t *testing.T) {
	_, err := NewRouter(map[domain.Channel]application.OtpSender{})
	if !errors.Is(err, ErrNoRoutes) {
		t.Fatalf("NewRouter(empty) = %v, want ErrNoRoutes", err)
	}
}

func TestRouterRejectsNilSender(t *testing.T) {
	_, err := NewRouter(map[domain.Channel]application.OtpSender{
		domain.ChannelSMS: nil,
	})
	if !errors.Is(err, ErrNoRoutes) {
		t.Fatalf("NewRouter(nil sender) = %v, want ErrNoRoutes", err)
	}
}

func TestRouterPropagatesProviderError(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	failingSender := &stubSender{channel: domain.ChannelSMS, err: providerErr}

	router, err := NewRouter(map[domain.Channel]application.OtpSender{
		domain.ChannelSMS: failingSender,
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	smsContact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	err = router.Send(context.Background(), smsContact, "123456")
	if !errors.Is(err, providerErr) {
		t.Fatalf("Send with provider error = %v, want provider error", err)
	}
}
