package emailotp

import (
	"context"
	"errors"
	"testing"

	emaildomain "github.com/stanleyHayes/obiara/internal/notifications/email/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type stubEmailService struct {
	calls     int
	recipient string
	code      string
	err       error
}

func (s *stubEmailService) Send(ctx context.Context, to string, template emaildomain.Template, params map[string]string) (string, error) {
	s.calls++
	s.recipient = to
	if code, ok := params["code"]; ok {
		s.code = code
	}
	return "ref-123", s.err
}

func TestEmailOtpSenderAcceptsEmailContact(t *testing.T) {
	service := &stubEmailService{}
	sender, err := NewSender(service)
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}

	contact := domain.ReconstituteContact(domain.ChannelEmail, "user@example.com")
	err = sender.Send(context.Background(), contact, "123456")
	if err != nil {
		t.Fatalf("Send with email contact: %v", err)
	}

	if service.calls != 1 {
		t.Fatalf("service calls = %d, want 1", service.calls)
	}
	if service.recipient != "user@example.com" {
		t.Fatalf("recipient = %q, want user@example.com", service.recipient)
	}
	if service.code != "123456" {
		t.Fatalf("code = %q, want 123456", service.code)
	}
}

func TestEmailOtpSenderRejectsNonEmailContact(t *testing.T) {
	service := &stubEmailService{}
	sender, err := NewSender(service)
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}

	smsContact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	err = sender.Send(context.Background(), smsContact, "123456")
	if !errors.Is(err, ErrWrongChannel) {
		t.Fatalf("Send with SMS contact = %v, want ErrWrongChannel", err)
	}

	if service.calls != 0 {
		t.Fatal("service should not be called for non-email contact")
	}
}

func TestEmailOtpSenderPropagatesProviderError(t *testing.T) {
	serviceErr := errors.New("email service unavailable")
	service := &stubEmailService{err: serviceErr}
	sender, err := NewSender(service)
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}

	contact := domain.ReconstituteContact(domain.ChannelEmail, "user@example.com")
	err = sender.Send(context.Background(), contact, "123456")
	if !errors.Is(err, serviceErr) {
		t.Fatalf("Send with service error = %v, want service error", err)
	}
}
