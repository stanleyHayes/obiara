package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

// Channel is the transport that carries a member's sign-in code.
//
// It is part of the identity, not a delivery preference: an account is
// anchored to the channel-and-value pair that was verified, so a member who
// verified a phone number and a member who verified an email address are
// two identities even when they are the same person. Linking them is a
// deliberate, audited act and does not happen implicitly here.
type Channel string

const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
)

var (
	// ErrUnknownChannel reports a channel outside the supported set.
	ErrUnknownChannel = errors.New("contact channel must be sms or email")
	// ErrInvalidEmail reports an address that is not a single valid mailbox.
	ErrInvalidEmail = errors.New("email address is not valid")
)

// e164Pattern bounds SMS contacts. Ghana numbers are the common case but
// the pattern is deliberately generic: the identity model must not encode a
// single country.
var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

// Contact is one verified way to reach a member, and the identity their
// account is keyed by (FR-102: exactly one active account per verified
// identity — enforced per channel-and-value pair).
type Contact struct {
	channel Channel
	value   string
}

// NewContact validates a contact for its channel and normalises it into the
// single form that will be stored and compared.
func NewContact(channel Channel, value string) (Contact, error) {
	trimmed := strings.TrimSpace(value)
	switch channel {
	case ChannelSMS:
		if !e164Pattern.MatchString(trimmed) {
			return Contact{}, ErrInvalidPhone
		}
		return Contact{channel: channel, value: trimmed}, nil
	case ChannelEmail:
		address, err := mail.ParseAddress(trimmed)
		// A display name ("Ama <ama@example.com>") parses but is not an
		// identity, so the address has to be the whole input.
		if err != nil || address.Address != trimmed {
			return Contact{}, ErrInvalidEmail
		}
		// Case is folded before storage because mailbox comparison is
		// case-insensitive in practice. Without this, Ama@example.com and
		// ama@example.com would key two accounts for one inbox and quietly
		// defeat the uniqueness FR-102 depends on.
		return Contact{channel: channel, value: strings.ToLower(trimmed)}, nil
	default:
		return Contact{}, ErrUnknownChannel
	}
}

// ParseChannel resolves a wire value, defaulting to SMS so that callers
// written before email existed keep working unchanged.
func ParseChannel(value string) (Channel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(ChannelSMS):
		return ChannelSMS, nil
	case string(ChannelEmail):
		return ChannelEmail, nil
	default:
		return "", ErrUnknownChannel
	}
}

// ReconstituteContact rebuilds a stored contact without policy checks.
func ReconstituteContact(channel Channel, value string) Contact {
	return Contact{channel: channel, value: value}
}

func (contact Contact) Channel() Channel { return contact.channel }
func (contact Contact) Value() string    { return contact.value }

// IsZero reports an unset contact, which no valid account or challenge ever
// carries.
func (contact Contact) IsZero() bool {
	return contact.channel == "" || contact.value == ""
}
