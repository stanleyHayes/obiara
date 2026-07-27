// Package email adapts the transactional email channel to the admin MFA
// code sender port.
package email

import (
	"context"

	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
)

// Sender delivers MFA codes through the email channel's admin_notice
// template.
type Sender struct {
	email application.EmailService
}

func NewSender(email application.EmailService) *Sender {
	return &Sender{email: email}
}

func (sender *Sender) SendMfaCode(ctx context.Context, emailAddress, code string) error {
	_, err := sender.email.Send(ctx, emailAddress, domain.TemplateAdminNotice, map[string]string{"code": code})
	return err
}
