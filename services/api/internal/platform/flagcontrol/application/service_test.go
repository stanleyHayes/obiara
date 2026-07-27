package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

const (
	a = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	b = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

type ids struct{ n int }

func (i *ids) NewID() string {
	i.n++
	if i.n == 1 {
		return "proposal:1"
	}
	return "audit:1"
}

type clock struct{ at time.Time }

func (c clock) Now() time.Time { return c.at }
func TestProposeRequiresSteppedAuthorityAndPersistsAudit(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, auth := NewMockRepository(ctrl), NewMockAuthority(ctrl)
	command := ProposeCommand{CommandID: "command:1", SessionID: "session:1", Capability: domain.CapabilityPayments, Environment: domain.EnvironmentProduction, Market: domain.MarketGH, Action: domain.ActionKill, Reason: domain.ReasonIncident}
	auth.EXPECT().RequireSteppedController(gomock.Any(), "session:1", domain.CapabilityPayments).Return(a, nil)
	repo.EXPECT().CreateWithAudit(gomock.Any(), gomock.Any(), gomock.Any())
	service := New(repo, auth, NewMockRuntime(ctrl), &ids{}, clock{time.Unix(1, 0)})
	p, err := service.Propose(context.Background(), command)
	if err != nil || p.Status() != domain.StatusProposed {
		t.Fatalf("status=%s err=%v", p.Status(), err)
	}
}
func TestApplyRequiresTheDistinctApproverAndUsesRuntimePort(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, auth, runtime := NewMockRepository(ctrl), NewMockAuthority(ctrl), NewMockRuntime(ctrl)
	p, _ := domain.NewProposal("proposal:1", "command:1", a, domain.CapabilitySow, domain.EnvironmentProduction, domain.MarketGH, domain.ActionEnable, domain.ReasonStagedRollout, time.Unix(1, 0), time.Unix(1, 0).Add(domain.MaxLifetime))
	approved, _ := p.Approve(b, time.Unix(2, 0))
	repo.EXPECT().Find(gomock.Any(), "proposal:1").Return(approved, nil)
	auth.EXPECT().RequireSteppedController(gomock.Any(), "session:2", domain.CapabilitySow).Return(b, nil)
	runtime.EXPECT().Apply(gomock.Any(), domain.EnvironmentProduction, domain.MarketGH, domain.RuntimeChange{Capability: domain.CapabilitySow, Enabled: true})
	repo.EXPECT().SaveWithAudit(gomock.Any(), gomock.Any(), approved.Version(), gomock.Any())
	service := New(repo, auth, runtime, &ids{}, clock{time.Unix(3, 0)})
	got, err := service.Apply(context.Background(), "proposal:1", "session:2")
	if err != nil || got.Status() != domain.StatusApplied {
		t.Fatalf("status=%s err=%v", got.Status(), err)
	}
}
