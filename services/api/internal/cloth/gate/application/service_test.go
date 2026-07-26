package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/gate/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestChangePersistsOnlyOpaqueCapability(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Now()
	current, _ := domain.Open("policy", domain.VersionV1, [2]string{key(1), key(2)}, domain.Command{ID: "open", ActorKey: key(1), Fingerprint: key(9), At: now})
	repo.EXPECT().Find(gomock.Any(), "policy").Return(current, nil)
	for _, item := range []struct {
		namespace, value string
		n                int
	}{{"member", "alice", 1}, {"reviewer", "reviewer", 3}, {"question", "question", 4}, {"material", "material", 5}} {
		keyer.EXPECT().Key(item.namespace, item.value).Return(key(item.n), nil)
	}
	repo.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "grant").DoAndReturn(func(_ context.Context, p domain.Policy, _ uint64, _ string) error {
		if p.State().Grants[0].Capability.ReviewerKey == "reviewer" {
			t.Fatal("raw ref persisted")
		}
		return nil
	})
	service := New(repo, keyer, ids, func() time.Time { return now })
	_, err := service.Grant(context.Background(), ChangeCommand{"grant", "policy", "alice", "reviewer", "question", "material", 1})
	if err != nil {
		t.Fatal(err)
	}
}
func key(n int) string {
	b := make([]byte, 64)
	for i := range b {
		b[i] = '0'
	}
	b[63] = byte('0' + n)
	return string(b)
}
