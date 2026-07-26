// Package inappsender delivers to the in-app inbox through the router.
package inappsender

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	inappdomain "github.com/stanleyHayes/obiara/internal/notifications/inapp/domain"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/application"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
)

// InAppStore is the inbox persistence surface.
type InAppStore interface {
	Add(context.Context, inappdomain.Notification) error
}

type Sender struct {
	store InAppStore
	now   func() time.Time
}

func New(store InAppStore, now func() time.Time) *Sender {
	return &Sender{store: store, now: now}
}

func (sender *Sender) Channel() domain.Channel { return domain.ChannelInApp }

func (sender *Sender) Send(ctx context.Context, outbound application.Outbound) (string, error) {
	notification, err := inappdomain.NewNotification(newID(), outbound.MemberID, domain.CategoryFor(outbound.Purpose), outbound.Reference, sender.now())
	if err != nil {
		return "", err
	}
	if err := sender.store.Add(ctx, notification); err != nil {
		return "", err
	}
	return "inapp:" + notification.ID(), nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "inapp_" + base64.RawURLEncoding.EncodeToString(id)
}
