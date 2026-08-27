package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewChallengeValidation(t *testing.T) {
	validContact := ReconstituteContact(ChannelSMS, "+233550000101")
	if _, err := NewChallenge("", validContact, "123456", 1, testNow); err != ErrChallengeIDNeeded {
		t.Fatalf("empty id = %v", err)
	}
	for _, phone := range []string{"", "0550000101", "+233 55 000 0101", "+0123", "abc"} {
		_, err := NewContact(ChannelSMS, phone)
		if err == nil {
			t.Fatalf("phone %q should fail validation", phone)
		}
	}
	if _, err := NewChallenge("ch-1", validContact, "12345", 1, testNow); err == nil {
		t.Fatal("short code must fail")
	}
	challenge, err := NewChallenge("ch-1", validContact, "123456", 1, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if challenge.CodeHash() == "123456" || strings.Contains(challenge.CodeHash(), "123456") {
		t.Fatal("code must be stored hashed")
	}
	if challenge.ExpiresAt() != testNow.Add(OtpLifetime) {
		t.Fatalf("expiry = %v", challenge.ExpiresAt())
	}
}

func TestVerifyFlow(t *testing.T) {
	contact := ReconstituteContact(ChannelSMS, "+233550000101")
	challenge, err := NewChallenge("ch-1", contact, "123456", 1, testNow)
	if err != nil {
		t.Fatal(err)
	}

	// Wrong code: attempts increment, challenge stays unconsumed.
	for i := 1; i <= OtpMaxAttempts; i++ {
		err := challenge.Verify("000000", testNow)
		if i < OtpMaxAttempts && err != ErrOtpMismatch {
			t.Fatalf("attempt %d = %v, want ErrOtpMismatch", i, err)
		}
	}
	if challenge.Attempts() != OtpMaxAttempts {
		t.Fatalf("attempts = %d", challenge.Attempts())
	}
	if err := challenge.Verify("123456", testNow); err != ErrOtpAttemptsExceeded {
		t.Fatalf("post-limit verify = %v, want ErrOtpAttemptsExceeded", err)
	}

	// Fresh challenge: expiry then success.
	fresh, _ := NewChallenge("ch-2", contact, "123456", 1, testNow)
	if err := fresh.Verify("123456", testNow.Add(OtpLifetime+time.Minute)); err != ErrOtpExpired {
		t.Fatalf("expired verify = %v, want ErrOtpExpired", err)
	}
	if err := fresh.Verify("123456", testNow.Add(time.Minute)); err != nil {
		t.Fatalf("valid verify = %v", err)
	}
	if fresh.ConsumedAt() == nil {
		t.Fatal("challenge must be consumed on success")
	}
	if err := fresh.Verify("123456", testNow.Add(2*time.Minute)); err != ErrOtpConsumed {
		t.Fatalf("re-verify = %v, want ErrOtpConsumed", err)
	}
}

func TestCheckResendPolicy(t *testing.T) {
	contact := ReconstituteContact(ChannelSMS, "+233550000101")
	recent, _ := NewChallenge("ch-1", contact, "123456", 1, testNow)

	if err := CheckResend(nil, testNow); err != nil {
		t.Fatalf("no prior challenge = %v", err)
	}
	if err := CheckResend(&recent, testNow.Add(30*time.Second)); err != ErrOtpRateLimited {
		t.Fatalf("immediate resend = %v, want rate limited", err)
	}
	if err := CheckResend(&recent, testNow.Add(61*time.Second)); err != nil {
		t.Fatalf("resend after interval = %v", err)
	}

	hourly, _ := NewChallenge("ch-2", contact, "123456", OtpHourlyLimit, testNow)
	if err := CheckResend(&hourly, testNow.Add(30*time.Minute)); err != ErrOtpRateLimited {
		t.Fatalf("hourly cap = %v, want rate limited", err)
	}
	if err := CheckResend(&hourly, testNow.Add(61*time.Minute)); err != nil {
		t.Fatalf("after window = %v", err)
	}
}

func TestGenerateCode(t *testing.T) {
	code, err := GenerateCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != otpCodeDigits {
		t.Fatalf("code length = %d", len(code))
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			t.Fatalf("non-digit in code %q", code)
		}
	}
}

func TestAccountLifecycle(t *testing.T) {
	contact, err := NewContact(ChannelSMS, "+233550000101")
	if err != nil {
		t.Fatal(err)
	}
	account, err := NewAccount("id_1", contact, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := account.Usable(); err != nil {
		t.Fatalf("active account = %v", err)
	}
	blocked := ReconstituteAccount("id_1", contact, AccountBlocked, TierUnverified, 1, nil, testNow)
	if err := blocked.Usable(); err != ErrAccountNotUsable {
		t.Fatalf("blocked account = %v, want ErrAccountNotUsable", err)
	}
	if _, err := NewContact(ChannelSMS, "bad-phone"); err == nil {
		t.Fatalf("bad phone should fail validation")
	}
}
