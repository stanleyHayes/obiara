package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/relay/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestSubmitAuthorizationFailClosed(t *testing.T) {
	c := gomock.NewController(t)
	r := NewMockRepository(c)
	k := NewMockKeyer(c)
	a := NewMockReviewerAuthorization(c)
	v, _ := domain.New(key(9), []string{key(1), key(2)}, key(3))
	k.EXPECT().Key("relay_pair", "pair").Return(key(9), nil)
	r.EXPECT().Find(gomock.Any(), key(9)).Return(v, nil)
	k.EXPECT().Key("relay_reviewer", "reviewer").Return(key(3), nil)
	a.EXPECT().Allowed(gomock.Any(), key(9), key(3)).Return(false, nil)
	_, e := NewService(r, k, a, nil, time.Now).Submit(context.Background(), "pair", "reviewer", "submit-command-01", "ref_questionabcdefghijklmnop", "ref_promptabcdefghijklmnop", 0)
	if e != domain.ErrDenied {
		t.Fatalf("%v", e)
	}
}
func TestProjectionRevalidatesAuthorizationAndConsent(t *testing.T) {
	c := gomock.NewController(t)
	r := NewMockRepository(c)
	k := NewMockKeyer(c)
	a := NewMockReviewerAuthorization(c)
	cons := NewMockConsentRevalidator(c)
	v, _ := domain.New(key(9), []string{key(1), key(2)}, key(3))
	v, _ = v.Submit(domain.Command{ID: "submit-command-01", ActorKey: key(3), QuestionRef: "ref_questionabcdefghijklmnop", PromptRef: "ref_promptabcdefghijklmnop", At: time.Now()})
	v, _ = v.Grant(domain.Command{ID: "grant-command-01", ActorKey: key(1), QuestionRef: "ref_questionabcdefghijklmnop", ResponseRef: "ref_responseabcdefghijklmnop", ExpectedRevision: 1, At: time.Now()})
	v, _ = v.Grant(domain.Command{ID: "grant-command-02", ActorKey: key(2), QuestionRef: "ref_questionabcdefghijklmnop", ResponseRef: "ref_responseabcdefghijklmnop", ExpectedRevision: 2, At: time.Now()})
	k.EXPECT().Key("relay_pair", "pair").Return(key(9), nil)
	r.EXPECT().Find(gomock.Any(), key(9)).Return(v, nil)
	k.EXPECT().Key("relay_reviewer", "reviewer").Return(key(3), nil)
	a.EXPECT().Allowed(gomock.Any(), key(9), key(3)).Return(true, nil)
	cons.EXPECT().Current(gomock.Any(), key(9), "ref_questionabcdefghijklmnop", "ref_responseabcdefghijklmnop", v.Members()).Return(true, nil)
	p, e := NewService(r, k, a, cons, time.Now).Project(context.Background(), "pair", "reviewer", "ref_questionabcdefghijklmnop")
	if e != nil || p.ResponseRef == "" {
		t.Fatalf("%+v %v", p, e)
	}
}
