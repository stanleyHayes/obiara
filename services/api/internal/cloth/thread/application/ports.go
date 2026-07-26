package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/thread/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Find(context.Context, string) (domain.Thread, error)
	Save(context.Context, domain.Thread, uint64, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type RevealEvidence interface {
	ThemeOneRevealed(context.Context, string, string) (bool, error)
}
