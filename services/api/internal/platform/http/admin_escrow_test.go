package apihttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	escrowdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	matchmakerdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type adminEscrowStub struct {
	settleCalls int
	escrow      escrowdomain.Escrow
	statement   escrowdomain.Statement
}

func (stub *adminEscrowStub) FundAudited(context.Context, string, string, string, uint64, escrowdomain.Terms, string, string) (escrowdomain.Escrow, error) {
	return escrowdomain.Escrow{}, nil
}
func (stub *adminEscrowStub) AddEvidenceAudited(context.Context, string, string, escrowdomain.EvidenceRole, string, string) (escrowdomain.Escrow, error) {
	return escrowdomain.Escrow{}, nil
}
func (stub *adminEscrowStub) SettleAudited(context.Context, string, string, string, string) (escrowdomain.Escrow, escrowdomain.Statement, error) {
	stub.settleCalls++
	return stub.escrow, stub.statement, nil
}

type operationsEngagementStub struct{}

func (operationsEngagementStub) FindForOperations(context.Context, string) (matchmakerdomain.Engagement, error) {
	return matchmakerdomain.Engagement{}, nil
}

func escrowTestKey(n int) string { return fmt.Sprintf("%064x", n) }

func TestEscrowSettlementRequiresFinanceScopeAndFreshMFA(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	escrow, err := escrowdomain.Fund(
		escrowTestKey(1), escrowTestKey(2), escrowTestKey(3), escrowTestKey(4), 1000,
		escrowdomain.Terms{ID: "terms.1", Version: 1, Milestones: []escrowdomain.MilestoneTerm{{ID: "consultation", GrossPesewas: 1000, FeePesewas: 100}}},
		"fund-1", now,
	)
	if err != nil {
		t.Fatal(err)
	}
	statement := escrowdomain.Statement{Ref: escrowTestKey(5), EscrowID: escrow.ID(), MilestoneID: "consultation", TermsID: "terms.1", TermsVersion: 1, GrossPesewas: 1000, FeePesewas: 100, NetPesewas: 900, SettledAt: now}
	tests := []struct {
		name      string
		principal admin.Principal
		want      int
		calls     int
	}{
		{"operations cannot settle", admin.Principal{ActorID: "ops", Scopes: []admin.Scope{admin.ScopeOperations}, MFAVerified: true}, http.StatusForbidden, 0},
		{"finance needs step up", admin.Principal{ActorID: "finance", Scopes: []admin.Scope{admin.ScopeFinance}}, http.StatusForbidden, 0},
		{"stepped up finance settles", admin.Principal{ActorID: "finance", Scopes: []admin.Scope{admin.ScopeFinance}, MFAVerified: true}, http.StatusOK, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &adminEscrowStub{escrow: escrow, statement: statement}
			mux := http.NewServeMux()
			RegisterAdminEscrowRoutes(mux, stub, operationsEngagementStub{}, func(*http.Request) (admin.Principal, error) {
				return test.principal, nil
			})
			request := httptest.NewRequest(http.MethodPost, "/v1/admin/escrows/"+escrow.ID()+"/milestones/consultation/settlement", http.NoBody)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "settle-1")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != test.want || stub.settleCalls != test.calls {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, stub.settleCalls, response.Body.String())
			}
		})
	}
}
