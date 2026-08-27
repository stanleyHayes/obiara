package apihttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	adminapp "github.com/stanleyHayes/obiara/services/api/internal/admin/application"
	verificationadmin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

// AdminNotificationSeenMarks stores, per operator, the moment they last
// acknowledged their inbox.
//
// One timestamp rather than a record per notification: the inbox is derived
// from live queues, so there is nothing durable to mark individually, and a
// watermark cannot drift out of step with the desks the way a duplicated
// notification row would.
type AdminNotificationSeenMarks interface {
	SeenAt(ctx context.Context, principalID string) (time.Time, error)
	MarkSeen(ctx context.Context, principalID string, at time.Time) error
}

// RegisterAdminNotificationRoutes adds the operator inbox.
//
// The inbox is a projection of queues that already exist, never a second
// copy of them. Nothing is written when work arrives, so a case resolved on
// its own desk cannot leave a stale notification behind — the next read
// simply does not find it.
//
// Only sources that enforce their own authorization are aggregated. The
// verification queue is scoped by principal and pending role changes are
// scoped by session, so a source an operator may not see returns forbidden
// and is omitted. Deriving access from roles here instead would mean
// inventing an access policy in a projection, which is exactly where one
// should not live.
func RegisterAdminNotificationRoutes(
	mux *http.ServeMux,
	admin Admin,
	verification AdminVerification,
	marks AdminNotificationSeenMarks,
	resolve AdminPrincipalResolver,
	now func() time.Time,
) {
	mux.Handle("GET /v1/admin/notifications", adminNotificationsHandler(admin, verification, marks, resolve))
	mux.Handle("POST /v1/admin/notifications/seen", adminNotificationsSeenHandler(admin, marks, now))
}

type adminNotificationItem struct {
	Key      string    `json:"key"`
	Title    string    `json:"title"`
	Detail   string    `json:"detail"`
	Count    int       `json:"count"`
	Href     string    `json:"href"`
	LatestAt time.Time `json:"latestAt"`
	Unread   bool      `json:"unread"`
}

type adminNotificationsResponse struct {
	UnreadCount int                     `json:"unreadCount"`
	SeenAt      *time.Time              `json:"seenAt"`
	Items       []adminNotificationItem `json:"items"`
}

type adminSeenResponse struct {
	SeenAt time.Time `json:"seenAt"`
}

func adminNotificationsHandler(
	admin Admin,
	verification AdminVerification,
	marks AdminNotificationSeenMarks,
	resolve AdminPrincipalResolver,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, actor, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}

		seenAt, err := marks.SeenAt(r.Context(), actor.ID())
		if err != nil {
			writeAdminError(w, r, err)
			return
		}

		items := make([]adminNotificationItem, 0, 2)

		if principal, resolveErr := resolve(r); resolveErr == nil && verification != nil {
			// A limit is required by the port; it bounds the projection
			// rather than the truth, so the count is reported as "at least
			// this many" by the desk itself when it is reached.
			cases, queueErr := verification.ListQueue(r.Context(), principal, adminNotificationScan)
			switch {
			case queueErr == nil && len(cases) > 0:
				latest := cases[0].SubmittedAt
				for _, item := range cases {
					if item.SubmittedAt.After(latest) {
						latest = item.SubmittedAt
					}
				}
				items = append(items, adminNotificationItem{
					Key:      "verification_queue",
					Title:    "Verifications awaiting review",
					Detail:   "Identity cases the provider could not decide.",
					Count:    len(cases),
					Href:     "/verification",
					LatestAt: latest.UTC(),
				})
			case queueErr != nil && !errors.Is(queueErr, verificationadmin.ErrForbidden):
				writeAdminError(w, r, queueErr)
				return
			}
		}

		changes, changesErr := admin.ListPendingRoleChanges(r.Context(), session.ID())
		switch {
		case changesErr == nil && len(changes) > 0:
			latest := changes[0].CreatedAt()
			for _, change := range changes {
				if change.CreatedAt().After(latest) {
					latest = change.CreatedAt()
				}
			}
			items = append(items, adminNotificationItem{
				Key:   "role_change_approvals",
				Title: "Admin-role changes awaiting a second approver",
				// Worth naming plainly: four-eyes proposals block until a
				// distinct administrator acts, and nothing else in the
				// console tells anyone one is waiting.
				Detail:   "A distinct administrator must approve these before they take effect.",
				Count:    len(changes),
				Href:     "/operators",
				LatestAt: latest.UTC(),
			})
		case changesErr != nil && !errors.Is(changesErr, adminapp.ErrNotAdmin):
			writeAdminError(w, r, changesErr)
			return
		}

		response := adminNotificationsResponse{Items: items}
		if !seenAt.IsZero() {
			marked := seenAt.UTC()
			response.SeenAt = &marked
		}
		for index := range response.Items {
			// An operator who has never acknowledged the inbox sees
			// everything as unread, which is the truthful reading of "not
			// yet seen" rather than a special first-run case.
			response.Items[index].Unread = seenAt.IsZero() ||
				response.Items[index].LatestAt.After(seenAt.UTC())
			if response.Items[index].Unread {
				response.UnreadCount++
			}
		}
		writeSuccess(w, r, http.StatusOK, response)
	})
}

// adminNotificationScan bounds how far the projection reads into a queue.
const adminNotificationScan = 50

func adminNotificationsSeenHandler(admin Admin, marks AdminNotificationSeenMarks, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, actor, ok := authenticatedAdmin(w, r, admin)
		if !ok {
			return
		}
		at := now().UTC()
		if err := marks.MarkSeen(r.Context(), actor.ID(), at); err != nil {
			writeAdminError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, adminSeenResponse{SeenAt: at})
	})
}
