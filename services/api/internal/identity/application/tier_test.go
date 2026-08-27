package application

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

func TestTierTransitionPersistsAudit(t *testing.T) {
	ctrl := gomock.NewController(t)
	accounts := NewMockAccountRepository(ctrl)
	service := NewTierService(accounts, fixedNow)

	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	account := domain.ReconstituteAccount("id_1", contact, domain.AccountActive, domain.TierUnverified, 1, nil, testNow)
	accounts.EXPECT().FindByID(gomock.Any(), "id_1").Return(account, nil)
	accounts.EXPECT().UpdateWithAudit(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated domain.Account, transition domain.TierTransition) error {
			if updated.Tier() != domain.TierVerified || updated.Version() != 2 {
				t.Fatalf("updated account tier=%d version=%d", updated.Tier(), updated.Version())
			}
			if transition.From != domain.TierUnverified || transition.To != domain.TierVerified || transition.ActorID != "verifier-1" {
				t.Fatalf("transition = %#v", transition)
			}
			return nil
		})

	transition, err := service.Transition(context.Background(), "id_1", domain.TierVerified, "ghana card verified", "verifier-1")
	if err != nil {
		t.Fatal(err)
	}
	if transition.To != domain.TierVerified {
		t.Fatalf("transition.To = %d", transition.To)
	}
}

func TestTierTransitionRejectsInvalidWithoutPersistence(t *testing.T) {
	ctrl := gomock.NewController(t)
	accounts := NewMockAccountRepository(ctrl)
	service := NewTierService(accounts, fixedNow)

	contact := domain.ReconstituteContact(domain.ChannelSMS, "+233550000101")
	account := domain.ReconstituteAccount("id_1", contact, domain.AccountActive, domain.TierUnverified, 1, nil, testNow)
	accounts.EXPECT().FindByID(gomock.Any(), "id_1").Return(account, nil)
	// No UpdateWithAudit expectation: any persistence call fails the test.

	if _, err := service.Transition(context.Background(), "id_1", domain.TierSowing, "skip", "actor-1"); err != domain.ErrInvalidTierTransition {
		t.Fatalf("Transition = %v, want ErrInvalidTierTransition", err)
	}
}
