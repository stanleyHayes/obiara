package application

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

func newRegistration(t *testing.T) (
	RegistrationService,
	*MockOtpChallengeRepository,
	*MockAccountRepository,
	*MockOtpSender,
	*MockSessionRepository,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	challenges := NewMockOtpChallengeRepository(ctrl)
	accounts := NewMockAccountRepository(ctrl)
	sender := NewMockOtpSender(ctrl)
	sessions := NewMockSessionRepository(ctrl)
	service := NewRegistrationService(
		challenges,
		accounts,
		sender,
		NewSessionService(sessions, fixedNow, func() string { return "sess_reg" }),
		fixedNow,
		func() string { return "ch_test" },
	)
	return service, challenges, accounts, sender, sessions
}

func TestRequestOtpSendsAndPersists(t *testing.T) {
	service, challenges, _, sender, _ := newRegistration(t)
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	challenges.EXPECT().LatestByContact(gomock.Any(), contact).Return(domain.OtpChallenge{}, ErrChallengeNotFound)
	sender.EXPECT().Send(gomock.Any(), contact, gomock.Any()).Return(nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	request, err := service.RequestOtp(context.Background(), contact)
	if err != nil {
		t.Fatal(err)
	}
	if request.ChallengeID != "ch_test" {
		t.Fatalf("challenge id = %q", request.ChallengeID)
	}
}

func TestRequestOtpRateLimited(t *testing.T) {
	service, challenges, _, _, _ := newRegistration(t)
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	recent, _ := domain.NewChallenge("ch_old", contact, "123456", 1, testNow)
	challenges.EXPECT().LatestByContact(gomock.Any(), contact).Return(recent, nil)

	if _, err := service.RequestOtp(context.Background(), contact); err != domain.ErrOtpRateLimited {
		t.Fatalf("RequestOtp = %v, want ErrOtpRateLimited", err)
	}
}

func TestRequestOtpRejectsZeroContact(t *testing.T) {
	service, challenges, _, _, _ := newRegistration(t)
	// A zero contact should be rejected at the domain level
	zeroContact := domain.Contact{}
	// The service will call LatestByContact, so we need to set up an expectation
	// to avoid mock complaints, but the error should come from NewChallenge which
	// rejects zero contacts.
	challenges.EXPECT().LatestByContact(gomock.Any(), zeroContact).Return(domain.OtpChallenge{}, ErrChallengeNotFound)
	if _, err := service.RequestOtp(context.Background(), zeroContact); err != domain.ErrContactRequired {
		t.Fatalf("RequestOtp with zero contact = %v, want ErrContactRequired", err)
	}
}

func TestVerifyOtpCreatesAccountAndSession(t *testing.T) {
	service, challenges, accounts, _, sessions := newRegistration(t)
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")

	// Stored challenge whose hash matches code 654321.
	stored, err := domain.NewChallenge("ch_test", contact, "654321", 1, testNow)
	if err != nil {
		t.Fatal(err)
	}
	challenges.EXPECT().LatestByContact(gomock.Any(), contact).Return(stored, nil)
	challenges.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, challenge domain.OtpChallenge) error {
			if challenge.ConsumedAt() == nil {
				t.Fatal("challenge must be consumed after successful verify")
			}
			return nil
		})
	accounts.EXPECT().FindByContact(gomock.Any(), contact).Return(domain.Account{}, ErrAccountNotFound)
	accounts.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sessions.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	issued, err := service.VerifyOtp(context.Background(), contact, "654321", "device-1")
	if err != nil {
		t.Fatal(err)
	}
	if issued.AccessToken == "" || issued.RefreshToken == "" {
		t.Fatal("session tokens must be issued")
	}
}

func TestVerifyOtpWrongCodePersistsAttempt(t *testing.T) {
	service, challenges, _, _, _ := newRegistration(t)
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	stored, _ := domain.NewChallenge("ch_test", contact, "654321", 1, testNow)
	challenges.EXPECT().LatestByContact(gomock.Any(), contact).Return(stored, nil)
	challenges.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, challenge domain.OtpChallenge) error {
			if challenge.Attempts() != 1 {
				t.Fatalf("attempts = %d, want 1", challenge.Attempts())
			}
			return nil
		})

	if _, err := service.VerifyOtp(context.Background(), contact, "000000", "device-1"); err != domain.ErrOtpMismatch {
		t.Fatalf("VerifyOtp = %v, want ErrOtpMismatch", err)
	}
}

func TestVerifyOtpBlockedAccountGetsNoSession(t *testing.T) {
	service, challenges, accounts, _, _ := newRegistration(t)
	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	stored, _ := domain.NewChallenge("ch_test", contact, "654321", 1, testNow)
	challenges.EXPECT().LatestByContact(gomock.Any(), contact).Return(stored, nil)
	challenges.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	accounts.EXPECT().FindByContact(gomock.Any(), contact).Return(
		domain.ReconstituteAccount("id_1", contact, domain.AccountBlocked, domain.TierUnverified, 1, nil, testNow), nil)

	if _, err := service.VerifyOtp(context.Background(), contact, "654321", "device-1"); err != domain.ErrAccountNotUsable {
		t.Fatalf("VerifyOtp = %v, want ErrAccountNotUsable", err)
	}
}
