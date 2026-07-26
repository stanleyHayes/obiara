package livekit

import (
	"context"
	"errors"
	"github.com/livekit/protocol/auth"
	"github.com/stanleyHayes/obiara/services/api/internal/realtime/livekit/application"
	"regexp"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid livekit configuration")
	ErrDenied        = errors.New("livekit token unavailable")
	opaquePattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type Config struct {
	APIKey, APISecret string
	MaxTTL            time.Duration
	AllowExplicitHost bool
	ClockSkew         time.Duration
}
type Adapter struct {
	cfg Config
	now func() time.Time
}

func New(cfg Config, now func() time.Time) (*Adapter, error) {
	if len(cfg.APIKey) < 8 || len(cfg.APISecret) < 32 || cfg.MaxTTL <= 0 || cfg.MaxTTL > time.Hour || cfg.ClockSkew < 0 || cfg.ClockSkew > time.Minute || now == nil {
		return nil, ErrInvalidConfig
	}
	return &Adapter{cfg, now}, nil
}
func (a *Adapter) Issue(_ context.Context, r application.JoinRequest) (application.JoinToken, error) {
	if a == nil || !opaquePattern.MatchString(r.RoomKey) || !opaquePattern.MatchString(r.ParticipantKey) || r.ServerTime.IsZero() || r.TTL <= 0 || r.TTL > a.cfg.MaxTTL {
		return application.JoinToken{}, ErrDenied
	}
	now := a.now().UTC()
	delta := r.ServerTime.UTC().Sub(now)
	if delta > a.cfg.ClockSkew || delta < -a.cfg.ClockSkew {
		return application.JoinToken{}, ErrDenied
	}
	pub, sub := false, false
	switch r.Role {
	case application.RoleListener:
		sub = true
	case application.RoleSpeaker:
		pub = true
		sub = true
	case application.RoleHost:
		if !a.cfg.AllowExplicitHost || !r.ExplicitHostContract {
			return application.JoinToken{}, ErrDenied
		}
		pub = true
		sub = true
	default:
		return application.JoinToken{}, ErrDenied
	}
	no := false
	grant := &auth.VideoGrant{RoomJoin: true, Room: r.RoomKey, CanPublish: &pub, CanSubscribe: &sub, CanPublishData: &pub, CanUpdateOwnMetadata: &no, CanSubscribeMetrics: &no, CanManageAgentSession: &no}
	signed, err := auth.NewAccessToken(a.cfg.APIKey, a.cfg.APISecret).SetIdentity(r.ParticipantKey).SetName(r.ParticipantKey).SetVideoGrant(grant).SetValidFor(r.TTL).ToJWT()
	if err != nil {
		return application.JoinToken{}, ErrDenied
	}
	return application.JoinToken{Signed: signed, ExpiresAt: now.Add(r.TTL)}, nil
}
