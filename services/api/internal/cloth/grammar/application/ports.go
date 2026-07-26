package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/grammar/domain"
)

var (
	ErrConcurrentChange = errors.New("cloth recipe changed concurrently")
	ErrUnavailable      = errors.New("cloth grammar unavailable")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Store(context.Context, domain.Recipe, uint64) (domain.Recipe, bool, error)
	Find(context.Context, string) (domain.Recipe, error)
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
