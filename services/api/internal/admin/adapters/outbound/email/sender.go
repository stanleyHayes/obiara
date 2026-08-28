// Package email adapts the transactional email channel to the admin
// operator-mailer port.
package email

import (
	"context"
	"strings"

	"github.com/stanleyHayes/obiara/internal/notifications/email/application"
	"github.com/stanleyHayes/obiara/internal/notifications/email/domain"
	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

// defaultConsoleURL is where an invited operator signs in. Overridden by
// ADMIN_CONSOLE_URL so a preview or staging invite does not point people at
// production.
const defaultConsoleURL = "https://admin.obiara.app"

// roleLabels renders a role the way the console names it. An invitation that
// said "ts_agent" would be telling someone their access in a vocabulary only
// this codebase uses.
var roleLabels = map[admindomain.Role]string{
	admindomain.RoleVerifier: "Verification",
	admindomain.RoleTSAgent:  "Trust & safety",
	admindomain.RoleHost:     "Community host",
	admindomain.RoleFinance:  "Finance",
	admindomain.RoleAdmin:    "Administrator",
}

// Sender delivers operator email through the email channel's templates.
type Sender struct {
	email      application.EmailService
	consoleURL string
}

func NewSender(email application.EmailService, consoleURL string) *Sender {
	url := strings.TrimSpace(consoleURL)
	if url == "" {
		url = defaultConsoleURL
	}
	return &Sender{email: email, consoleURL: url}
}

// SendInvite tells a new operator they have access and where to sign in.
func (sender *Sender) SendInvite(ctx context.Context, emailAddress string, roles []admindomain.Role) error {
	labels := make([]string, 0, len(roles))
	for _, role := range roles {
		if label, ok := roleLabels[role]; ok {
			labels = append(labels, label)
			continue
		}
		labels = append(labels, string(role))
	}
	_, err := sender.email.Send(ctx, emailAddress, domain.TemplateOperatorInvite, map[string]string{
		"roles":   strings.Join(labels, ", "),
		"console": sender.consoleURL,
	})
	return err
}

func (sender *Sender) SendMfaCode(ctx context.Context, emailAddress, code string) error {
	_, err := sender.email.Send(ctx, emailAddress, domain.TemplateAdminNotice, map[string]string{"code": code})
	return err
}
