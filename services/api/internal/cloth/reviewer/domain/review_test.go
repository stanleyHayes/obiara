package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var testNow = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func key(number int) string { return fmt.Sprintf("%064x", number) }
func command(id string, revision uint64) Command {
	return Command{ID: id, ExpectedRevision: revision, At: testNow}
}
func pending(t *testing.T) Review {
	t.Helper()
	review, err := Create(State{
		ID: "review-1", Members: []string{key(1), key(2)}, ReviewerKey: key(3),
		InviteDigest: key(4), OTPDigest: key(5), WatermarkRef: key(6),
		QuestionRefs: []string{"question-1"}, MaterialRefs: []string{"material-1"},
		OTPExpiresAt: testNow.Add(10 * time.Minute), InviteExpiresAt: testNow.Add(24 * time.Hour),
	}, testNow, command("create", 0))
	if err != nil {
		t.Fatal(err)
	}
	return review
}

func TestOTPExpirySingleUseRevokeAndBoundedProjection(t *testing.T) {
	review := pending(t)
	if _, err := review.Project(key(7), testNow); !errors.Is(err, ErrCredential) {
		t.Fatalf("bearer access before redemption=%v", err)
	}
	if _, err := review.Redeem(key(4), key(5), key(7), testNow.Add(10*time.Minute), command("late", 1)); !errors.Is(err, ErrExpired) {
		t.Fatalf("OTP boundary=%v", err)
	}
	redeemed, err := review.Redeem(key(4), key(5), key(7), testNow.Add(time.Minute), command("redeem", 1))
	if err != nil {
		t.Fatal(err)
	}
	projection, err := redeemed.Project(key(7), testNow.Add(2*time.Minute))
	if err != nil || projection.WatermarkRef != key(6) || len(projection.QuestionRefs) != 1 {
		t.Fatalf("projection=%+v err=%v", projection, err)
	}
	if _, err = redeemed.Redeem(key(4), key(5), key(8), testNow.Add(2*time.Minute), command("again", 2)); !errors.Is(err, ErrRedeemed) {
		t.Fatalf("second redemption=%v", err)
	}
	revoked, err := redeemed.Revoke(testNow.Add(3*time.Minute), command("revoke", 2))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = revoked.Project(key(7), testNow.Add(3*time.Minute)); !errors.Is(err, ErrRevoked) {
		t.Fatalf("revoked projection=%v", err)
	}
}

func TestValidityBoundsProperty(t *testing.T) {
	property := func(otpSeconds uint16, inviteMinutes uint16) bool {
		otp := time.Duration(otpSeconds) * time.Second
		invite := time.Duration(inviteMinutes) * time.Minute
		_, err := Create(State{
			ID: "review-1", Members: []string{key(1), key(2)}, ReviewerKey: key(3),
			InviteDigest: key(4), OTPDigest: key(5), WatermarkRef: key(6),
			QuestionRefs: []string{"question-1"},
			OTPExpiresAt: testNow.Add(otp), InviteExpiresAt: testNow.Add(invite),
		}, testNow, command("create", 0))
		valid := otp > 0 && otp <= MaxOTPValidity && invite > 0 && invite <= 24*time.Hour && otp <= invite
		return (err == nil) == valid
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}
