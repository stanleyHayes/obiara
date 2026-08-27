// Package emailotp delivers a member's sign-in code over the email channel.
//
// It is a thin adapter rather than a second Resend client on purpose: the
// email context already owns rendering, delivery logging and the
// provider-reference correlation that the delivery-status webhook depends
// on. Going around it would leave member sign-in emails invisible to the
// same operational view that covers every other message the platform sends.
package emailotp

import (
	"context"
	"errors"
	"fmt"

	emaildomain "github.com/stanleyHayes/obiara/internal/notifications/email/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// ErrWrongChannel reports a contact that is not an email address. The
// channel router should never hand one over, so this is a composition bug
// rather than a member-facing condition.
var ErrWrongChannel = errors.New("email otp sender received a non-email contact")

// EmailService is the outbound port onto the email context.
type EmailService interface {
	Send(ctx context.Context, to string, template emaildomain.Template, params map[string]string) (string, error)
}

// Sender implements the identity OTP sender port over email.
type Sender struct {
	emails EmailService
}

// NewSender builds the adapter.
func NewSender(emails EmailService) (*Sender, error) {
	if emails == nil {
		return nil, errors.New("email otp sender requires an email service")
	}
	return &Sender{emails: emails}, nil
}

// Send renders and delivers the sign-in code.
//
// The provider reference is deliberately dropped: identity's sender port
// reports only success or failure, and correlating a delivery back to a
// member is exactly the linkage the email context's own log is scoped to
// hold rather than something identity should carry.
func (sender *Sender) Send(ctx context.Context, contact domain.Contact, code string) error {
	if contact.Channel() != domain.ChannelEmail {
		return fmt.Errorf("%w: %s", ErrWrongChannel, contact.Channel())
	}
	_, err := sender.emails.Send(ctx, contact.Value(), emaildomain.TemplateMemberSignIn,
		map[string]string{"code": code})
	return err
}
