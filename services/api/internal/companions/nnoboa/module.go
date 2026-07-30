package nnoboa

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	whatsappdomain "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/application"
)

// Module wires the Nnoboa nomination slice (FR-1302).
type Module struct {
	Nominations *application.NominationService
}

// SenderFunc adapts a plain function to application.NotificationSender.
type SenderFunc func(ctx context.Context, msg whatsappdomain.Message) error

// Send implements application.NotificationSender.
func (f SenderFunc) Send(ctx context.Context, msg whatsappdomain.Message) error {
	return f(ctx, msg)
}

// NewModule constructs the slice.
func NewModule(ctx context.Context, db *mongo.Database, sender application.NotificationSender, inviteSecret string) (*Module, error) {
	repo, err := mongodb.NewNominationRepository(ctx, db)
	if err != nil {
		return nil, err
	}
	return &Module{
		Nominations: application.NewNominationService(repo, sender, func() time.Time { return time.Now().UTC() }, inviteSecret),
	}, nil
}
