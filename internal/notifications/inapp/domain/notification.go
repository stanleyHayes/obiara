// Package domain models the in-app notification inbox (E13-S03). Every
// routed notification can land here as the always-available channel.
package domain

import (
	"errors"
	"time"

	notificationdomain "github.com/stanleyHayes/obiara/internal/notifications/domain"
)

var ErrNotificationInvalid = errors.New("invalid in-app notification")

// Notification is one inbox entry. Reference is opaque (pod, fire, ritual
// or case id); content renders from the localization registry at read
// time, never from free text at write time.
type Notification struct {
	id        string
	memberID  string
	category  notificationdomain.Category
	reference string
	createdAt time.Time
	readAt    *time.Time
}

func NewNotification(id, memberID string, category notificationdomain.Category, reference string, now time.Time) (Notification, error) {
	if id == "" || memberID == "" || reference == "" {
		return Notification{}, ErrNotificationInvalid
	}
	return Notification{id: id, memberID: memberID, category: category, reference: reference, createdAt: now.UTC()}, nil
}

// Reconstitute rebuilds a stored notification without policy checks.
func Reconstitute(id, memberID string, category notificationdomain.Category, reference string, createdAt time.Time, readAt *time.Time) Notification {
	return Notification{id: id, memberID: memberID, category: category, reference: reference, createdAt: createdAt, readAt: readAt}
}

func (notification *Notification) MarkRead(now time.Time) {
	if notification.readAt == nil {
		read := now.UTC()
		notification.readAt = &read
	}
}

func (n Notification) ID() string                            { return n.id }
func (n Notification) MemberID() string                      { return n.memberID }
func (n Notification) Category() notificationdomain.Category { return n.category }
func (n Notification) Reference() string                     { return n.reference }
func (n Notification) CreatedAt() time.Time                  { return n.createdAt }
func (n Notification) ReadAt() *time.Time                    { return n.readAt }
