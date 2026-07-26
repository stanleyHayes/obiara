package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/recording/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestStartRevalidatesAuthorityAndMembershipBeforeRecorder(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	authority := NewMockAuthority(ctrl)
	membership := NewMockMembership(ctrl)
	recorder := NewMockRecorder(ctrl)
	now := time.Now()
	p, _ := domain.Open("policy", key(9), key(1), []string{key(1)}, domain.Command{ID: "open", ActorKey: key(1), Fingerprint: key(8), At: now})
	p, _ = p.Propose(domain.PurposeReflection, time.Hour, command("p", key(1), domain.ActionProposed, domain.PurposeReflection, time.Hour, 1, now))
	p, _ = p.OptIn(command("o", key(1), domain.ActionOptedIn, "", 0, 2, now))
	repo.EXPECT().Find(gomock.Any(), "policy").Return(p, nil)
	keyer.EXPECT().Key("participant", "host").Return(key(1), nil)
	authority.EXPECT().Authorize(gomock.Any(), p.State(), domain.ActionStarted, key(1)).Return(nil)
	membership.EXPECT().Current(gomock.Any(), key(9)).Return([]string{key(1)}, nil)
	recorder.EXPECT().Start(gomock.Any(), key(9), domain.PurposeReflection, time.Hour, "start").Return(key(7), nil)
	repo.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(3), "start").Return(nil)
	service := New(repo, keyer, ids, authority, membership, recorder, func() time.Time { return now })
	if _, err := service.Start(context.Background(), Command{"start", "policy", "host", "", domain.PurposeReflection, time.Hour, 3}); err != nil {
		t.Fatal(err)
	}
}
func command(id, actor string, action domain.Action, p domain.Purpose, r time.Duration, rev uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, Fingerprint: domain.Fingerprint("policy", action, actor, actor, p, r, rev), ExpectedRevision: rev, At: at}
}
func key(n int) string {
	b := make([]byte, 64)
	for i := range b {
		b[i] = '0'
	}
	b[63] = byte('0' + n)
	return string(b)
}
