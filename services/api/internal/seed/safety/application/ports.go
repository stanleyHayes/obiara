package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/safety/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Find(context.Context, string) (domain.Bucket, error)
	Create(context.Context, domain.Bucket) error
	Save(context.Context, domain.Bucket, uint64) error
	AppendCareSignal(context.Context, CareSignal) error
}

type Keyer interface {
	Key(string, string) (string, error)
}

type CareSignal struct {
	ActorKey, Code string
	WindowRevision uint64
}
