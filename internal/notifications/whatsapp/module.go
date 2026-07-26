// Package whatsapp is the composition root of the WhatsApp channel slice
// (E13-S05): OTP and pod alerts over strict templates. The production
// WhatsApp Business adapter replaces the simulator after vendor scoring.
package whatsapp

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/internal/notifications/whatsapp/application"
)

type Module struct {
	Channel application.ChannelService
	// Sender is the active provider adapter (simulator until a scored
	// WhatsApp Business vendor is selected).
	Sender application.Sender
}

// NewModule builds the channel. decider may be nil to bypass the E13-S01
// preference gate (composition wires the notifications service).
func NewModule(ctx context.Context, database *mongo.Database, decider application.PreferenceDecider) (Module, error) {
	deliveryLog := mongodb.NewDeliveryLog(database)
	if err := deliveryLog.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	sender := simulator.NewSender()
	return Module{
		Channel: application.NewChannelService(sender, deliveryLog, decider, time.Now),
		Sender:  sender,
	}, nil
}

// OtpSenderAdapter adapts the channel to the identity context's OtpSender
// port, so OTP delivery can switch from the identity simulator to WhatsApp
// at composition without touching the identity module.
type OtpSenderAdapter struct {
	Channel application.ChannelService
}

func (adapter OtpSenderAdapter) Send(ctx context.Context, phone, code string) error {
	_, err := adapter.Channel.SendOtp(ctx, phone, code)
	return err
}
