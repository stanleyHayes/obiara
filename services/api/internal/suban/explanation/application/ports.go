package application

import (
	"context"
	suban "github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Authority interface {
	RequireSelf(context.Context, string, string) error
	RequireAppealReviewer(context.Context, string) error
}
type EventSource interface {
	ListForSubject(context.Context, string) ([]suban.Event, error)
	FindForSubject(context.Context, string, string) (suban.Event, error)
}
type AppealRepository interface {
	Create(context.Context, domain.Appeal) error
	Find(context.Context, string) (domain.Appeal, error)
	Save(context.Context, domain.Appeal, uint64, string) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
