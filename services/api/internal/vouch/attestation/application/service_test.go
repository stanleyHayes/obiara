package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestProposalRequiresPolicyBound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	keyer := NewMockKeyer(ctrl)
	policy := NewMockStakePolicy(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, authorizer, keyer, policy, ids, func() time.Time { return now })
	authorizer.EXPECT().Require(gomock.Any(), "voucher-1", "vouch.attestation.propose", "").Return(nil)
	policy.EXPECT().Validate(gomock.Any(), "circle", uint8(25)).Return("policy-v1", nil)
	keyer.EXPECT().Key("vouch-attestation:subject", "subject-1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("vouch-attestation:actor", "voucher-1").Return(strings.Repeat("b", 64), nil)
	keyer.EXPECT().Key("vouch-attestation:scope", "circle-1").Return(strings.Repeat("c", 64), nil)
	ids.EXPECT().NewID().Return("attestation-1")
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	result, err := service.Propose(context.Background(), Command{
		ID: "propose-1", ActorID: "voucher-1", ReasonCode: "voucher_proposed",
	}, Proposal{
		SubjectID: "subject-1", VoucherID: "voucher-1", ScopeKind: "circle",
		ScopeID: "circle-1", StakeUnits: 25, TTL: time.Hour,
	})
	if err != nil || result.Attestation.StakeUnits() != 25 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
