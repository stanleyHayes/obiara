// Package domain defines the transactional email channel (E13-S04) backed
// by Resend: transactional and admin/ops notifications over bounded
// templates. Member marketing email does not exist at Obiara.
package domain

import (
	"errors"
	"net/mail"
	"strings"
)

// Template is a known email template. Content renders from the template
// registry; params are bounded placeholders only.
type Template string

const (
	TemplateOpsAlert         Template = "ops_alert"
	TemplateAdminNotice      Template = "admin_notice"
	TemplateVerificationHelp Template = "verification_help"
	// TemplateMemberSignIn carries a member's sign-in code when their
	// verified contact is an email address rather than a phone number.
	TemplateMemberSignIn Template = "member_sign_in"
	// TemplateOperatorInvite tells someone they have been given operator
	// access. It carries no credential: an enrolled operator signs in with a
	// code sent to this same address, so a secret in the invitation would be
	// a bearer token sitting in an inbox for nothing.
	TemplateOperatorInvite Template = "operator_invite"
)

var (
	ErrInvalidRecipient = errors.New("recipient must be a valid email address")
	ErrInvalidTemplate  = errors.New("unknown email template")
	ErrParamTooLong     = errors.New("email template parameter too long")
)

// Message is one outbound email.
type Message struct {
	to       string
	template Template
	params   map[string]string
}

func NewMessage(to string, template Template, params map[string]string) (Message, error) {
	address, err := mail.ParseAddress(strings.TrimSpace(to))
	if err != nil || address.Address != strings.TrimSpace(to) {
		return Message{}, ErrInvalidRecipient
	}
	switch template {
	case TemplateOpsAlert, TemplateAdminNotice, TemplateVerificationHelp, TemplateMemberSignIn,
		TemplateOperatorInvite:
	default:
		return Message{}, ErrInvalidTemplate
	}
	for key, value := range params {
		if len(value) > 500 {
			return Message{}, ErrParamTooLong
		}
		_ = key
	}
	return Message{to: address.Address, template: template, params: params}, nil
}

func (message Message) To() string                { return message.to }
func (message Message) Template() Template        { return message.template }
func (message Message) Params() map[string]string { return message.params }
