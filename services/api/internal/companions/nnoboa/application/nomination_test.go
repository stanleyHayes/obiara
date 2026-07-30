package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	whatsappdomain "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/domain"
)

var testClock = func() time.Time { return time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC) }

const testInviteSecret = "synthetic-nnoboa-invite-secret-at-least-32-bytes"

func validInput() NominateInput {
	return NominateInput{
		MemberID:     "mem_12345678",
		KinName:      "Auntie Efua",
		KinPhone:     "+233550000101",
		Relationship: "aunt",
	}
}

func TestNominatePersistsAndInvites(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockNominationRepository(ctrl)
	sender := NewMockNotificationSender(ctrl)
	svc := NewNominationService(repo, sender, testClock, testInviteSecret)

	repo.EXPECT().HasPendingForKin(gomock.Any(), "mem_12345678", "+233550000101").Return(false, nil)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, n domain.Nomination) error {
			if n.Status != domain.StatusPending || n.Relationship != domain.Aunt {
				t.Errorf("unexpected nomination: %+v", n)
			}
			return nil
		})
	sender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msg whatsappdomain.Message) error {
			if msg.Template() != whatsappdomain.TemplateNnoboaConsent {
				t.Errorf("template = %s", msg.Template())
			}
			if msg.To() != "+233550000101" || msg.Params()["kin_name"] != "Auntie Efua" {
				t.Errorf("unexpected message: %s %+v", msg.To(), msg.Params())
			}
			if len(msg.Params()["consent_token"]) < 32 {
				t.Errorf("consent token missing: %+v", msg.Params())
			}
			return nil
		})

	n, err := svc.Nominate(context.Background(), validInput())
	if err != nil {
		t.Fatalf("Nominate: %v", err)
	}
	if n.Status != domain.StatusPending {
		t.Fatalf("status = %s", n.Status)
	}
}

func TestNominateRejectsDuplicatePending(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockNominationRepository(ctrl)
	sender := NewMockNotificationSender(ctrl)
	svc := NewNominationService(repo, sender, testClock, testInviteSecret)

	repo.EXPECT().HasPendingForKin(gomock.Any(), "mem_12345678", "+233550000101").Return(true, nil)

	if _, err := svc.Nominate(context.Background(), validInput()); !errors.Is(err, ErrDuplicateNomination) {
		t.Fatalf("want ErrDuplicateNomination, got %v", err)
	}
}

func TestNominateRejectsInvalidInput(t *testing.T) {
	svc := NewNominationService(nil, nil, testClock, testInviteSecret)
	in := validInput()
	in.KinPhone = "0550000101"
	if _, err := svc.Nominate(context.Background(), in); !errors.Is(err, domain.ErrInvalidNomination) {
		t.Fatalf("want ErrInvalidNomination, got %v", err)
	}
}

func TestConsentTransitions(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockNominationRepository(ctrl)
	svc := NewNominationService(repo, nil, testClock, testInviteSecret)

	pending, err := domain.NewNomination("mem_12345678", "Auntie Efua", "+233550000101", "aunt", testClock())
	if err != nil {
		t.Fatal(err)
	}
	repo.EXPECT().FindByID(gomock.Any(), pending.ID).Return(pending, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, n domain.Nomination) error {
			if n.Status != domain.StatusConsented || n.RespondedAt == nil || n.Version != 2 {
				t.Errorf("unexpected update: %+v", n)
			}
			return nil
		})

	n, err := svc.Consent(context.Background(), pending.ID, svc.inviteToken(pending.ID))
	if err != nil {
		t.Fatalf("Consent: %v", err)
	}
	if n.Status != domain.StatusConsented {
		t.Fatalf("status = %s", n.Status)
	}
}

func TestConsentRejectsInvalidInviteTokenBeforeLookup(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockNominationRepository(ctrl)
	svc := NewNominationService(repo, nil, testClock, testInviteSecret)

	if _, err := svc.Consent(context.Background(), "nom_private", "wrong"); !errors.Is(err, ErrInvalidInviteToken) {
		t.Fatalf("want ErrInvalidInviteToken, got %v", err)
	}
}

func TestLazyExpiryOnGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockNominationRepository(ctrl)
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	svc := NewNominationService(repo, nil, func() time.Time { return now }, testInviteSecret)

	stale, err := domain.NewNomination("mem_12345678", "Auntie Efua", "+233550000101", "aunt", now.Add(-domain.NominationExpiry))
	if err != nil {
		t.Fatal(err)
	}
	repo.EXPECT().FindByID(gomock.Any(), stale.ID).Return(stale, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, n domain.Nomination) error {
			if n.Status != domain.StatusExpired {
				t.Errorf("status = %s", n.Status)
			}
			return nil
		})

	n, err := svc.Get(context.Background(), stale.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if n.Status != domain.StatusExpired {
		t.Fatalf("status = %s", n.Status)
	}
}

func TestDeclineAfterExpiryRejected(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockNominationRepository(ctrl)
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	svc := NewNominationService(repo, nil, func() time.Time { return now }, testInviteSecret)

	stale, err := domain.NewNomination("mem_12345678", "Auntie Efua", "+233550000101", "aunt", now.Add(-domain.NominationExpiry))
	if err != nil {
		t.Fatal(err)
	}
	repo.EXPECT().FindByID(gomock.Any(), stale.ID).Return(stale, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	if _, err := svc.Decline(context.Background(), stale.ID, svc.inviteToken(stale.ID)); !errors.Is(err, domain.ErrNotPending) {
		t.Fatalf("want ErrNotPending, got %v", err)
	}
}
