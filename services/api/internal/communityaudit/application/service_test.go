package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/domain"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
	"time"
)

func TestEvidenceRequiresRecentMFAAndRecordsKeyedAccess(t *testing.T) {
	c := gomock.NewController(t)
	a := NewMockAuthorizer(c)
	m := NewMockMFAGate(c)
	k := NewMockKeyer(c)
	r := NewMockRepository(c)
	n := time.Now()
	a.EXPECT().Authorize(gomock.Any(), "admin:1", CapEvidence).Return(nil)
	m.EXPECT().Recent(gomock.Any(), "admin:1", gomock.Any()).Return(true)
	v, _ := domain.New("case:1", domain.KindCircle, "circle:key", "evidence:key", "legitimacy.review", n)
	r.EXPECT().Find(gomock.Any(), "case:1").Return(v, nil)
	k.EXPECT().Key("community_auditor", "admin:1").Return(strings.Repeat("a", 64), nil)
	k.EXPECT().Key("community_correlation", "corr:1").Return(strings.Repeat("b", 64), nil)
	k.EXPECT().Key("community_purpose", "legitimacy_review").Return(strings.Repeat("c", 64), nil)
	r.EXPECT().RecordEvidenceAccess(gomock.Any(), "case:1", strings.Repeat("a", 64), gomock.Any(), gomock.Any()).Return(nil)
	e, err := New(a, m, k, r, func() time.Time { return n }).Evidence(context.Background(), "admin:1", "case:1", "legitimacy_review", "corr:1")
	if err != nil || e.Reference != "evidence:key" {
		t.Fatal(e, err)
	}
}
