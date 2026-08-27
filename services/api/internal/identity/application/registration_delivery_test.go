package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// TestRequestOtpReportsDeliveryFailure covers what the production smoke run
// found: an SMS provider rejecting the send surfaced as a bare 500 with
// nothing in the logs, so a member saw "something went wrong" and an operator
// had no way to learn why.
func TestRequestOtpReportsDeliveryFailure(t *testing.T) {
	service, challenges, _, sender, _ := newRegistration(t)
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	providerErr := errors.New(`arkesel: delivery failed: provider status "error", provider code 104`)

	challenges.EXPECT().LatestByContact(gomock.Any(), contact).
		Return(domain.OtpChallenge{}, ErrChallengeNotFound)
	sender.EXPECT().Send(gomock.Any(), contact, gomock.Any()).Return(providerErr)
	// No Create expectation: gomock fails on any unexpected call, so this
	// asserts that an undelivered code is never persisted. Storing it would
	// ask the member for a code that does not exist and burn their attempts.

	_, err := service.RequestOtp(context.Background(), contact)
	if err == nil {
		t.Fatal("RequestOtp reported success while the code was never sent")
	}
	if !errors.Is(err, ErrCodeDeliveryFailed) {
		t.Errorf("error %v should wrap ErrCodeDeliveryFailed so transport maps it to a 503", err)
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("error %v lost the provider cause", err)
	}
	if !strings.Contains(err.Error(), "104") {
		t.Errorf("error %v should carry the provider code for triage", err)
	}
}
