package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/decline/domain"
)

var (
	ErrUnavailable = errors.New("seed decline service unavailable")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Store interface {
	Record(context.Context, domain.Decline, domain.Notification) (stored domain.Decline, replayed bool, err error)
	IsExcluded(context.Context, string, string, time.Time) (bool, error)
	// IsPairExcluded reports whether this decliner is still shielded from
	// this sower. IsExcluded above locks a seed; this locks the pair, which
	// is what "you may not reach for them again for 90 days" means.
	IsPairExcluded(ctx context.Context, declinerKey, sowerKey string, at time.Time) (bool, error)
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface {
	NewID() string
}
