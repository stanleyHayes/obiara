package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	notificationdomain "github.com/stanleyHayes/obiara/internal/notifications/domain"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
)

var routerNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func outbound(purpose domain.Purpose) Outbound {
	return Outbound{MemberID: "m-1", Phone: "+233550000101", Purpose: purpose, Template: "t", Reference: "ref-1"}
}

func TestDeliverFirstChannelWins(t *testing.T) {
	ctrl := gomock.NewController(t)
	push := NewMockChannelSender(ctrl)
	inapp := NewMockChannelSender(ctrl)
	decider := NewMockPreferenceDecider(ctrl)

	push.EXPECT().Channel().Return(domain.ChannelPush).AnyTimes()
	inapp.EXPECT().Channel().Return(domain.ChannelInApp).AnyTimes()
	decider.EXPECT().Decide(gomock.Any(), "m-1", notificationdomain.CategoryRitual).
		Return(notificationdomain.Decision{Allowed: true}, nil)
	push.EXPECT().Send(gomock.Any(), gomock.Any()).Return("push-ref", nil)
	// in-app must not be consulted after push succeeds.

	router := NewRouter([]ChannelSender{push, inapp}, decider, func() time.Time { return routerNow })
	result, err := router.Deliver(context.Background(), outbound(domain.PurposeRitual))
	if err != nil || result.Channel != domain.ChannelPush {
		t.Fatalf("result = %#v, %v", result, err)
	}
}

func TestDeliverFallsBack(t *testing.T) {
	ctrl := gomock.NewController(t)
	sms := NewMockChannelSender(ctrl)
	whatsapp := NewMockChannelSender(ctrl)

	sms.EXPECT().Channel().Return(domain.ChannelSMS).AnyTimes()
	whatsapp.EXPECT().Channel().Return(domain.ChannelWhatsApp).AnyTimes()
	sms.EXPECT().Send(gomock.Any(), gomock.Any()).Return("", errors.New("sms down"))
	whatsapp.EXPECT().Send(gomock.Any(), gomock.Any()).Return("wa-ref", nil)

	// OTP bypasses preferences entirely (nil decider).
	router := NewRouter([]ChannelSender{sms, whatsapp}, nil, func() time.Time { return routerNow })
	result, err := router.Deliver(context.Background(), outbound(domain.PurposeOtp))
	if err != nil || result.Channel != domain.ChannelWhatsApp || result.ProviderRef != "wa-ref" {
		t.Fatalf("result = %#v, %v, want WhatsApp fallback", result, err)
	}
}

func TestDeliverAllChannelsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	sms := NewMockChannelSender(ctrl)
	whatsapp := NewMockChannelSender(ctrl)

	sms.EXPECT().Channel().Return(domain.ChannelSMS).AnyTimes()
	whatsapp.EXPECT().Channel().Return(domain.ChannelWhatsApp).AnyTimes()
	sms.EXPECT().Send(gomock.Any(), gomock.Any()).Return("", errors.New("down"))
	whatsapp.EXPECT().Send(gomock.Any(), gomock.Any()).Return("", errors.New("down"))

	router := NewRouter([]ChannelSender{sms, whatsapp}, nil, func() time.Time { return routerNow })
	if _, err := router.Deliver(context.Background(), outbound(domain.PurposeOtp)); err != ErrAllChannelsFailed {
		t.Fatalf("deliver = %v, want ErrAllChannelsFailed", err)
	}
}

func TestSuppressedByPreferences(t *testing.T) {
	ctrl := gomock.NewController(t)
	push := NewMockChannelSender(ctrl)
	decider := NewMockPreferenceDecider(ctrl)
	// No Send expectation: suppression delivers nothing.
	push.EXPECT().Channel().Return(domain.ChannelPush).AnyTimes()
	decider.EXPECT().Decide(gomock.Any(), "m-1", notificationdomain.CategoryPods).
		Return(notificationdomain.Decision{Allowed: false, Reason: "muted"}, nil)

	router := NewRouter([]ChannelSender{push}, decider, func() time.Time { return routerNow })
	result, err := router.Deliver(context.Background(), outbound(domain.PurposePods))
	if err != nil || result.Channel != "" {
		t.Fatalf("result = %#v, want suppressed", result)
	}
}

func TestMissingChannelSkipped(t *testing.T) {
	ctrl := gomock.NewController(t)
	inapp := NewMockChannelSender(ctrl)
	inapp.EXPECT().Channel().Return(domain.ChannelInApp).AnyTimes()
	inapp.EXPECT().Send(gomock.Any(), gomock.Any()).Return("inapp:1", nil)

	// Only in-app registered: pods ladder skips push and whatsapp.
	router := NewRouter([]ChannelSender{inapp}, nil, func() time.Time { return routerNow })
	result, err := router.Deliver(context.Background(), outbound(domain.PurposePods))
	if err != nil || result.Channel != domain.ChannelInApp {
		t.Fatalf("result = %#v, %v", result, err)
	}
}
