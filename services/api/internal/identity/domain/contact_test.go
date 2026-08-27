package domain

import (
	"testing"
)

func TestNewContactSMS(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"e164 valid", "+233550000101", nil},
		{"e164 with spaces", "  +233550000101  ", nil},
		{"not e164 no plus", "233550000101", ErrInvalidPhone},
		{"not e164 spaces", "+233 55 000 0101", ErrInvalidPhone},
		{"too short", "+0123", ErrInvalidPhone},
		{"empty", "", ErrInvalidPhone},
		{"non-digit", "abc", ErrInvalidPhone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			contact, err := NewContact(ChannelSMS, tc.value)
			if err != tc.wantErr {
				t.Fatalf("NewContact = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && contact.Value() != contact.Value() {
				t.Fatalf("value mismatch")
			}
		})
	}
}

func TestNewContactEmail(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr error
		wantVal string // after normalization
	}{
		{"valid lowercase", "user@example.com", nil, "user@example.com"},
		{"valid mixed case normalized", "User@Example.COM", nil, "user@example.com"},
		{"with spaces trimmed", "  user@example.com  ", nil, "user@example.com"},
		{"display name rejected", "User <user@example.com>", ErrInvalidEmail, ""},
		{"not an email", "not-an-email", ErrInvalidEmail, ""},
		{"empty", "", ErrInvalidEmail, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			contact, err := NewContact(ChannelEmail, tc.value)
			if err != tc.wantErr {
				t.Fatalf("NewContact = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil {
				if contact.Value() != tc.wantVal {
					t.Fatalf("value = %q, want %q", contact.Value(), tc.wantVal)
				}
				if contact.Channel() != ChannelEmail {
					t.Fatalf("channel = %v, want %v", contact.Channel(), ChannelEmail)
				}
			}
		})
	}
}

func TestParseChannel(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantChan Channel
		wantErr  error
	}{
		{"empty defaults to sms", "", ChannelSMS, nil},
		{"sms explicit", "sms", ChannelSMS, nil},
		{"SMS uppercase", "SMS", ChannelSMS, nil},
		{"email explicit", "email", ChannelEmail, nil},
		{"EMAIL uppercase", "EMAIL", ChannelEmail, nil},
		{"unknown channel", "carrier-pigeon", "", ErrUnknownChannel},
		{"whitespace trimmed", "  sms  ", ChannelSMS, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			channel, err := ParseChannel(tc.input)
			if err != tc.wantErr {
				t.Fatalf("ParseChannel(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}
			if tc.wantErr == nil && channel != tc.wantChan {
				t.Fatalf("ParseChannel(%q) = %v, want %v", tc.input, channel, tc.wantChan)
			}
		})
	}
}

func TestContactIsZero(t *testing.T) {
	cases := []struct {
		name     string
		contact  Contact
		wantZero bool
	}{
		{"zero value", Contact{}, true},
		{"empty channel", Contact{channel: "", value: "+233550000101"}, true},
		{"empty value", Contact{channel: ChannelSMS, value: ""}, true},
		{"both empty", Contact{channel: "", value: ""}, true},
		{"sms contact valid", ReconstituteContact(ChannelSMS, "+233550000101"), false},
		{"email contact valid", ReconstituteContact(ChannelEmail, "user@example.com"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.contact.IsZero()
			if got != tc.wantZero {
				t.Fatalf("IsZero() = %v, want %v", got, tc.wantZero)
			}
		})
	}
}
