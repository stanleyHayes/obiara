package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/domain"
	"time"
)

var (
	ErrDenied      = errors.New("circle room access denied")
	ErrNotFound    = errors.New("circle room entry not found")
	ErrConflict    = errors.New("circle room conflict")
	ErrUnavailable = errors.New("circle room unavailable")
)

type Capability string

const (
	CapabilityRead Capability = "read"
	CapabilityPost Capability = "post"
	CapabilityHost Capability = "host"
)

type Decision struct {
	CircleID, ActorID string
	Capability        Capability
}
type Authorizer interface {
	Authorize(context.Context, Decision) error
}
type Repository interface {
	Create(context.Context, domain.Entry) (domain.Entry, bool, error)
	Find(context.Context, string) (domain.Entry, error)
	Delete(context.Context, domain.Entry, uint64, string) error
	List(context.Context, string, time.Time, int) ([]domain.Entry, error)
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDs interface{ NewID() string }
