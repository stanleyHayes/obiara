// Package push is the composition root of the device push channel.
package push

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/notifications/push/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/push/application"
)

// ErrSenderRequired reports a module built with no push transport.
var ErrSenderRequired = errors.New("push module requires a sender")

type Module struct {
	Push application.Service
}

// NewModule builds the channel. sender is the provider adapter chosen by the
// composition root from PUSH_PROVIDER and must not be nil.
func NewModule(ctx context.Context, database *mongo.Database, sender application.Sender) (Module, error) {
	if sender == nil {
		return Module{}, ErrSenderRequired
	}
	registry := mongodb.NewRegistry(database, time.Now)
	if err := registry.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{Push: application.NewService(registry, sender, time.Now)}, nil
}
