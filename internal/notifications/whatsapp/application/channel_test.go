package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/notifications/domain"
)

var channelNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestSendOtpBypassesPreferences(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := NewMockSender(ctrl)
	log := NewMockDeliveryLog(ctrl)
	// No decider at all: OTP must never consult preferences.

	sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return("ref-1", nil)
	log.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, entry DeliveryEntry) error {
			if entry.Template != "otp_code" || entry.Status != "sent" || entry.ProviderRef != "ref-1" {
				t.Fatalf("entry = %#v", entry)
			}
			return nil
		})

	service := NewChannelService(sender, log, nil, func() time.Time { return channelNow })
	ref, err := service.SendOtp(context.Background(), "+233550000101", "123456")
	if err != nil || ref != "ref-1" {
		t.Fatalf("SendOtp = %q, %v", ref, err)
	}
}

func TestPodAlertSuppressedByPreferences(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := NewMockSender(ctrl)
	log := NewMockDeliveryLog(ctrl)
	decider := NewMockPreferenceDecider(ctrl)
	// No Send or Record expectation: suppression sends and logs nothing.
	decider.EXPECT().Decide(gomock.Any(), "m-1", domain.CategoryPods).
		Return(domain.Decision{Allowed: false, Reason: "muted"}, nil)

	service := NewChannelService(sender, log, decider, func() time.Time { return channelNow })
	ref, err := service.SendPodAlert(context.Background(), "m-1", "+233550000101", "pod_1")
	if err != nil || ref != "" {
		t.Fatalf("SendPodAlert = %q, %v; want suppressed", ref, err)
	}
}

func TestPodAlertDeliveredWhenAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := NewMockSender(ctrl)
	log := NewMockDeliveryLog(ctrl)
	decider := NewMockPreferenceDecider(ctrl)
	decider.EXPECT().Decide(gomock.Any(), "m-1", domain.CategoryPods).
		Return(domain.Decision{Allowed: true, Reason: "allowed"}, nil)
	sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return("ref-2", nil)
	log.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)

	service := NewChannelService(sender, log, decider, func() time.Time { return channelNow })
	if _, err := service.SendPodAlert(context.Background(), "m-1", "+233550000101", "pod_1"); err != nil {
		t.Fatal(err)
	}
}

func TestDeliveryFailureLoggedAndReported(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := NewMockSender(ctrl)
	log := NewMockDeliveryLog(ctrl)
	sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return("", errors.New("provider down"))
	log.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, entry DeliveryEntry) error {
			if entry.Status != "failed" {
				t.Fatalf("entry = %#v, want failed status", entry)
			}
			return nil
		})

	service := NewChannelService(sender, log, nil, func() time.Time { return channelNow })
	if _, err := service.SendOtp(context.Background(), "+233550000101", "123456"); err != ErrDeliveryFailed {
		t.Fatalf("SendOtp = %v, want ErrDeliveryFailed", err)
	}
}
