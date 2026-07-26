// Package notifications is the composition root of the Notifications and
// Localization bounded context slice for preferences and caps (E13-S01).
// Channel dispatchers (push, SMS, WhatsApp, Resend) land with E13-S03+.
package notifications

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/notifications/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/notifications/application"
)

type Module struct {
	Notifications application.NotificationService
}

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Notifications: application.NewNotificationService(repository, repository, time.Now),
	}, nil
}
