package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedIDs struct{ n int }

func (f *fixedIDs) NewID() string { f.n++; return "id:" + string(rune('a'+f.n)) }

type fixedClock struct{ at time.Time }

func (f fixedClock) Now() time.Time { return f.at }

func TestApplyRequiresSignatureAndExactLedgerProof(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, signatures, ledger, keys := NewMockRepository(ctrl), NewMockSignatureVerifier(ctrl), NewMockLedger(ctrl), NewMockKeyer(ctrl)
	env := SignedEnvelope{Provider: "provider-value", EventID: "event-value", Reference: "reference-value", LedgerCommand: "ledger:1", Currency: "GHS", Status: "settled", Minor: 500, OccurredUnix: 10}
	signatures.EXPECT().Verify(gomock.Any(), env)
	keys.EXPECT().Key("reconciliation-provider", env.Provider).Return(testKey('a'), nil)
	keys.EXPECT().Key("reconciliation-event:"+testKey('a'), env.EventID).Return(testKey('b'), nil)
	keys.EXPECT().Key("reconciliation-reference:"+testKey('a'), env.Reference).Return(testKey('c'), nil)
	repo.EXPECT().AppendFact(gomock.Any(), gomock.Any())
	ledger.EXPECT().Proof(gomock.Any(), "ledger:1").Return(domain.LedgerProof{CommandID: "ledger:1", ReferenceKey: testKey('c'), Currency: domain.CurrencyGHS, Minor: 501, Balanced: true}, true, nil)
	repo.EXPECT().AppendAudit(gomock.Any(), gomock.Any())
	s := New(repo, signatures, ledger, keys, &fixedIDs{}, fixedClock{time.Unix(11, 0)})
	_, decision, err := s.Apply(context.Background(), env)
	if err != nil || decision.Outcome() != domain.OutcomeException || decision.Exception() != domain.ExceptionAmount {
		t.Fatalf("decision=%s/%s err=%v", decision.Outcome(), decision.Exception(), err)
	}
}

func TestUnsignedEnvelopeCannotMutate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, signatures := NewMockRepository(ctrl), NewMockSignatureVerifier(ctrl)
	env := SignedEnvelope{}
	signatures.EXPECT().Verify(gomock.Any(), env).Return(errors.New("bad signature"))
	s := New(repo, signatures, NewMockLedger(ctrl), NewMockKeyer(ctrl), &fixedIDs{}, fixedClock{})
	if _, _, err := s.Apply(context.Background(), env); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestIdempotentReplayRequiresSameSemanticFact(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, signatures, ledger, keys := NewMockRepository(ctrl), NewMockSignatureVerifier(ctrl), NewMockLedger(ctrl), NewMockKeyer(ctrl)
	env := SignedEnvelope{Provider: "p", EventID: "e", Reference: "r", LedgerCommand: "ledger:1", Currency: "GHS", Status: "settled", Minor: 500, OccurredUnix: 10}
	signatures.EXPECT().Verify(gomock.Any(), env)
	keys.EXPECT().Key(gomock.Any(), gomock.Any()).DoAndReturn(func(ns, v string) (string, error) {
		switch {
		case ns == "reconciliation-provider":
			return testKey('a'), nil
		case v == "e":
			return testKey('b'), nil
		default:
			return testKey('c'), nil
		}
	}).Times(3)
	now := time.Unix(11, 0)
	existing, _ := domain.NewFact("existing", testKey('a'), testKey('b'), testKey('c'), "ledger:1", domain.CurrencyGHS, domain.StatusSettled, 500, time.Unix(10, 0), now)
	repo.EXPECT().AppendFact(gomock.Any(), gomock.Any()).Return(ErrApplied)
	repo.EXPECT().FindFactByEvent(gomock.Any(), testKey('b')).Return(existing, nil)
	ledger.EXPECT().Proof(gomock.Any(), "ledger:1").Return(domain.LedgerProof{CommandID: "ledger:1", ReferenceKey: testKey('c'), Currency: domain.CurrencyGHS, Minor: 500, Balanced: true}, true, nil)
	repo.EXPECT().AppendAudit(gomock.Any(), gomock.Any()).Return(ErrApplied)
	s := New(repo, signatures, ledger, keys, &fixedIDs{}, fixedClock{now})
	if _, d, err := s.Apply(context.Background(), env); err != nil || d.Outcome() != domain.OutcomeReconciled {
		t.Fatalf("d=%s err=%v", d.Outcome(), err)
	}
}
func testKey(ch byte) string {
	b := make([]byte, 64)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}
