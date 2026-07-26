package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/grammar/domain"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestServiceHMACBoundaryPrecedesRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	values := []struct{ namespace, raw, key string }{{"pair", "alice", key(1)}, {"pair", "bob", key(2)}, {"theme", "warm", key(3)}, {"provenance", "mutual", key(4)}}
	for _, value := range values {
		keyer.EXPECT().Key(value.namespace, value.raw).Return(value.key, nil)
	}
	repository.EXPECT().Store(gomock.Any(), gomock.Any(), uint64(0)).DoAndReturn(func(_ context.Context, recipe domain.Recipe, _ uint64) (domain.Recipe, bool, error) {
		if recipe.Pair()[0] == "alice" {
			t.Fatal("raw member reached repository")
		}
		return recipe, false, nil
	})
	service := New(repository, keyer)
	result, err := service.Compile(context.Background(), Command{"command", domain.VersionV1, [2]string{"alice", "bob"}, []string{"warm"}, []string{"mutual"}})
	if err != nil || result.Recipe.RenderSeed() == "" {
		t.Fatal(result, err)
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
