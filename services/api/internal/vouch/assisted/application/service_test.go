package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/vouch/assisted/domain"
)

func TestOnlyNamedVoucherCanConsent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	request, err := domain.NewRequest(
		"vouch-1", strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64),
		now.Add(time.Hour), domain.Command{
			ID: "request-1", ActorKey: strings.Repeat("b", 64),
			ReasonCode: "assisted_request", At: now,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(repository, authorizer, keyer, NewMockIDSource(ctrl), func() time.Time { return now })
	repository.EXPECT().Find(gomock.Any(), "vouch-1").Return(request, nil)
	authorizer.EXPECT().Require(gomock.Any(), "wrong-voucher", "vouch.consent", "vouch-1").Return(nil)
	keyer.EXPECT().Key("vouch:actor", "wrong-voucher").Return(strings.Repeat("d", 64), nil)
	_, err = service.Consent(context.Background(), Command{
		ID: "consent-1", RequestID: "vouch-1", ActorID: "wrong-voucher",
		ExpectedRevision: 1, ReasonCode: "voucher_consented",
	})
	if !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("wrong voucher consent = %v", err)
	}
}

func TestOperatorDecisionRequiresAuthorizationBeforeSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	service := NewService(repository, authorizer, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now)
	repository.EXPECT().Find(gomock.Any(), "vouch-1").Return(domain.Request{}, nil)
	authorizer.EXPECT().Require(gomock.Any(), "operator-1", "vouch.decide", "").Return(errors.New("denied"))
	_, err := service.Decide(context.Background(), Command{
		ID: "decision-1", RequestID: "vouch-1", ActorID: "operator-1",
		ExpectedRevision: 2, ReasonCode: "identity_confirmed",
	}, domain.DecisionApprove)
	if !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("unauthorized decision = %v", err)
	}
}
