//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	roommongo "github.com/stanleyHayes/obiara/services/api/internal/circle/room/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"strings"
	"testing"
	"time"
)

type auth struct{}

func (auth) Authorize(_ context.Context, d application.Decision) error {
	if d.ActorID == "outsider" {
		return errors.New("denied")
	}
	if d.Capability == application.CapabilityHost && d.ActorID != "host:1" {
		return errors.New("denied")
	}
	return nil
}

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "entry:" + string(rune('0'+i.n)) }
func TestAuthorizationReplayAndRetention(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	c, e := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Terminate(context.Background())
	u, _ := c.ConnectionString(ctx)
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	cl, e := apimongo.Connect(ctx, u+sep+"directConnection=true")
	if e != nil {
		t.Fatal(e)
	}
	defer cl.Disconnect(context.Background())
	repo := roommongo.NewRepository(cl.Database("room_test"))
	if e = repo.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	k, _ := privacy.New([]byte(strings.Repeat("k", 32)))
	now := time.Now().UTC()
	id := &ids{}
	s := application.NewService(auth{}, repo, k, id, func() time.Time { return now })
	m, _ := domain.NewMediaRef("asset:1", "transcript:1", "audio/ogg", time.Minute)
	q := application.Create{CommandID: "cmd:1", CircleID: "circle:1", ActorID: "member:1", Media: m, Retention: time.Hour}
	first, e := s.Voice(ctx, q)
	if e != nil {
		t.Fatal(e)
	}
	again, e := s.Voice(ctx, q)
	if e != nil || again.ID() != first.ID() {
		t.Fatal("replay failed", e)
	}
	if _, e = s.List(ctx, "circle:1", "outsider", 10); !errors.Is(e, application.ErrDenied) {
		t.Fatal("read auth bypass")
	}
	items, e := s.List(ctx, "circle:1", "member:1", 10)
	if e != nil || len(items) != 1 {
		t.Fatal(items, e)
	}
	now = now.Add(2 * time.Hour)
	items, e = s.List(ctx, "circle:1", "member:1", 10)
	if e != nil || len(items) != 0 {
		t.Fatal("expired entry visible")
	}
}
