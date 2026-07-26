package application

import (
	"context"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Role string

const (
	RoleListener Role = "listener"
	RoleSpeaker  Role = "speaker"
	RoleHost     Role = "host"
)

type JoinRequest struct {
	RoomKey, ParticipantKey string
	Role                    Role
	ServerTime              time.Time
	TTL                     time.Duration
	ExplicitHostContract    bool
}
type JoinToken struct {
	Signed    string
	ExpiresAt time.Time
}
type TokenIssuer interface {
	Issue(context.Context, JoinRequest) (JoinToken, error)
}
type Service struct{ issuer TokenIssuer }

func NewService(i TokenIssuer) Service { return Service{i} }
func (s Service) Issue(ctx context.Context, r JoinRequest) (JoinToken, error) {
	return s.issuer.Issue(ctx, r)
}
