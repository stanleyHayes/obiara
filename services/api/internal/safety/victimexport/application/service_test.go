package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type keyer struct{}

func (keyer) Key(ns, v string) (string, error) {
	switch ns {
	case "victim-export-member":
		return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
	case "victim-export-redaction":
		return "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", nil
	default:
		return "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
}

type ids struct{}

func (ids) NewID() string { return "export:1" }
func TestRequestRequiresAllowlistAndRedactionForExactPurpose(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	authority := NewMockAuthority(ctrl)
	allow := NewMockAllowlist(ctrl)
	redactor := NewMockRedactor(ctrl)
	now := time.Date(2026, 7, 26, 20, 0, 0, 0, time.UTC)
	authority.EXPECT().RequireMember(gomock.Any(), "actor", "member").Return(nil)
	allow.EXPECT().Require(gomock.Any(), domain.KindIncidentSummary, "case:1", domain.PurposeVictimSupport).Return(nil)
	redactor.EXPECT().RedactReporterAndThirdParties(gomock.Any(), domain.KindIncidentSummary, "case:1", domain.PurposeVictimSupport).Return("redaction:attested", nil)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, e domain.Export) error {
		if len(e.References()) != 1 || e.References()[0].RedactionKey == "" {
			t.Fatal("redaction missing")
		}
		return nil
	})
	s := NewService(repo, authority, allow, redactor, keyer{}, ids{}, func() time.Time { return now })
	_, e := s.Request(context.Background(), RequestCommand{Actor: "actor", Member: "member", CommandID: "request", Purpose: domain.PurposeVictimSupport, References: []ReferenceRequest{{Kind: domain.KindIncidentSummary, ReferenceID: "case:1"}}})
	if e != nil {
		t.Fatal(e)
	}
}
