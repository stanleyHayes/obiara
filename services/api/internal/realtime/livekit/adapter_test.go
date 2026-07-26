package livekit

import (
	"context"
	"fmt"
	"github.com/livekit/protocol/auth"
	"github.com/stanleyHayes/obiara/services/api/internal/realtime/livekit/application"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func adapter(t *testing.T, host bool) (*Adapter, time.Time) {
	now := time.Now().UTC().Truncate(time.Second)
	a, e := New(Config{APIKey: "api-key-123", APISecret: strings.Repeat("s", 32), MaxTTL: 15 * time.Minute, AllowExplicitHost: host, ClockSkew: time.Second}, func() time.Time { return now })
	if e != nil {
		t.Fatal(e)
	}
	return a, now
}
func verify(t *testing.T, tok string) (*auth.ClaimGrants, time.Duration) {
	t.Helper()
	v, e := auth.ParseAPIToken(tok)
	if e != nil {
		t.Fatal(e)
	}
	registered, g, e := v.Verify(strings.Repeat("s", 32))
	if e != nil {
		t.Fatal(e)
	}
	return g, registered.ExpiresAt.Time.Sub(registered.IssuedAt.Time)
}
func TestRoleClaimsAreLeastPrivilege(t *testing.T) {
	cases := []struct {
		role           application.Role
		pub, sub, host bool
	}{{application.RoleListener, false, true, false}, {application.RoleSpeaker, true, true, false}, {application.RoleHost, true, true, true}}
	for _, tc := range cases {
		a, now := adapter(t, tc.host)
		got, e := a.Issue(context.Background(), application.JoinRequest{RoomKey: key(1), ParticipantKey: key(2), Role: tc.role, ServerTime: now, TTL: 5 * time.Minute, ExplicitHostContract: tc.host})
		if e != nil {
			t.Fatal(e)
		}
		g, ttl := verify(t, got.Signed)
		v := g.Video
		if g.Identity != key(2) || g.Name != key(2) || !v.RoomJoin || v.Room != key(1) || *v.CanPublish != tc.pub || *v.CanSubscribe != tc.sub || v.RoomCreate || v.RoomAdmin || v.RoomList || v.RoomRecord || v.Recorder || v.IngressAdmin || ttl != 5*time.Minute {
			t.Fatalf("%+v ttl=%v", g, ttl)
		}
	}
}
func TestRejectsRawIdentifiersTTLAndImplicitHost(t *testing.T) {
	a, now := adapter(t, false)
	base := application.JoinRequest{RoomKey: key(1), ParticipantKey: key(2), Role: application.RoleSpeaker, ServerTime: now, TTL: time.Minute}
	bad := []application.JoinRequest{base, base, base, base}
	bad[0].RoomKey = "room@example.com"
	bad[1].ParticipantKey = "+233555123456"
	bad[2].TTL = 16 * time.Minute
	bad[3].Role = application.RoleHost
	for _, r := range bad {
		if _, e := a.Issue(context.Background(), r); e != ErrDenied {
			t.Fatalf("%+v %v", r, e)
		}
	}
}
func TestConfigAndServerTimeFailClosed(t *testing.T) {
	if _, e := New(Config{}, time.Now); e != ErrInvalidConfig {
		t.Fatal(e)
	}
	a, now := adapter(t, false)
	r := application.JoinRequest{RoomKey: key(1), ParticipantKey: key(2), Role: application.RoleListener, ServerTime: now.Add(2 * time.Second), TTL: time.Minute}
	if _, e := a.Issue(context.Background(), r); e != ErrDenied {
		t.Fatal(e)
	}
}
func TestTokenContainsNoRawPersonalData(t *testing.T) {
	a, now := adapter(t, false)
	got, _ := a.Issue(context.Background(), application.JoinRequest{RoomKey: key(1), ParticipantKey: key(2), Role: application.RoleListener, ServerTime: now, TTL: time.Minute})
	for _, bad := range []string{"email", "phone", "name@example.com", "+233"} {
		if strings.Contains(got.Signed, bad) {
			t.Fatal(bad)
		}
	}
}
