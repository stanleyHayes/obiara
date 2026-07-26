// Package domain defines the WhatsApp channel for P0 (E13-S05): OTP codes
// and pod alerts only, over provider-approved templates. FR-701: no room
// content ever crosses this channel — templates take bounded parameters,
// never free text.
package domain

import (
	"errors"
	"regexp"
	"strings"
)

// Template is a provider-approved message template.
type Template string

const (
	TemplateOtpCode  Template = "otp_code"
	TemplatePodAlert Template = "pod_alert"
)

var (
	ErrInvalidPhone       = errors.New("recipient must be an E.164 phone number")
	ErrInvalidTemplate    = errors.New("unknown whatsapp template")
	ErrInvalidOtpCode     = errors.New("otp code must be 6 digits")
	ErrPodRefRequired     = errors.New("pod reference is required")
	ErrFreeTextNotAllowed = errors.New("whatsapp templates take bounded parameters, never free text")
)

var (
	e164Pattern = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	otpPattern  = regexp.MustCompile(`^\d{6}$`)
)

// Message is one outbound WhatsApp message.
type Message struct {
	to       string
	template Template
	params   map[string]string
}

// NewOtpMessage builds the OTP message (identity-safety class).
func NewOtpMessage(to, code string) (Message, error) {
	if !e164Pattern.MatchString(to) {
		return Message{}, ErrInvalidPhone
	}
	if !otpPattern.MatchString(code) {
		return Message{}, ErrInvalidOtpCode
	}
	return Message{to: to, template: TemplateOtpCode, params: map[string]string{"code": code}}, nil
}

// NewPodAlertMessage builds the pod-arrival alert (pods notification class).
// The reference is an opaque pod identifier — never pod content.
func NewPodAlertMessage(to, podRef string) (Message, error) {
	if !e164Pattern.MatchString(to) {
		return Message{}, ErrInvalidPhone
	}
	if strings.TrimSpace(podRef) == "" {
		return Message{}, ErrPodRefRequired
	}
	return Message{to: to, template: TemplatePodAlert, params: map[string]string{"ref": podRef}}, nil
}

func (message Message) To() string                { return message.to }
func (message Message) Template() Template        { return message.template }
func (message Message) Params() map[string]string { return message.params }
