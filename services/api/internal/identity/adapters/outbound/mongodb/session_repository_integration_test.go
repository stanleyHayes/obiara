//go:build integration

package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	apimongo "github.com/stanleyHayes/obiara/services/api/internal/platform/mongo"
)

const integrationTimeout = 3 * time.Minute

func startMongo(t *testing.T) (context.Context, *mongo.Database) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	t.Cleanup(cancel)

	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("read Testcontainer connection string: %v", err)
	}
	// See platform/mongo integration tests: directConnection keeps the
	// driver on the host-mapped port of the single-node replica set.
	// ConnectionString already carries ?replicaSet=..., so append with the
	// right separator.
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	return ctx, client.Database("obiara_identity_test")
}

func newTestSession(t *testing.T, id, memberID, deviceID string) domain.Session {
	t.Helper()
	now := time.Now()
	access, err := domain.IssueAccessToken(id, now)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := domain.IssueRefreshToken(id, now)
	if err != nil {
		t.Fatal(err)
	}
	session, err := domain.Start(id, memberID, deviceID, access, refresh, now)
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func TestSessionRepositoryRoundTripAndConcurrency(t *testing.T) {
	ctx, database := startMongo(t)
	repository := NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}

	session := newTestSession(t, "sess_rt", "member-1", "device-1")
	if err := repository.Create(ctx, session); err != nil {
		t.Fatalf("create: %v", err)
	}

	loaded, err := repository.FindByID(ctx, "sess_rt")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if loaded.MemberID() != "member-1" || loaded.RefreshTokenHash() != session.RefreshTokenHash() {
		t.Fatalf("round trip mismatch: %#v", loaded)
	}

	// Rotate and persist; a second update with the stale version must fail.
	now := time.Now()
	access, _ := domain.IssueAccessToken("sess_rt", now)
	refresh, _ := domain.IssueRefreshToken("sess_rt", now)
	if err := loaded.Rotate(access, refresh, now); err != nil {
		t.Fatal(err)
	}
	if err := repository.Update(ctx, loaded); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := repository.Update(ctx, loaded); !errors.Is(err, application.ErrStaleSession) {
		t.Fatalf("stale update error = %v, want ErrStaleSession", err)
	}

	reloaded, err := repository.FindByID(ctx, "sess_rt")
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Version() != 2 {
		t.Fatalf("version = %d, want 2", reloaded.Version())
	}
}

func TestSessionRepositoryRevokeScopes(t *testing.T) {
	ctx, database := startMongo(t)
	repository := NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}

	for _, session := range []domain.Session{
		newTestSession(t, "sess_a1", "member-a", "device-1"),
		newTestSession(t, "sess_a2", "member-a", "device-2"),
		newTestSession(t, "sess_b1", "member-b", "device-1"),
	} {
		if err := repository.Create(ctx, session); err != nil {
			t.Fatalf("create %s: %v", session.ID(), err)
		}
	}

	if err := repository.RevokeAllForDevice(ctx, "device-1", time.Now()); err != nil {
		t.Fatalf("revoke device: %v", err)
	}
	for id, want := range map[string]domain.Status{
		"sess_a1": domain.StatusRevoked,
		"sess_b1": domain.StatusRevoked,
		"sess_a2": domain.StatusActive,
	} {
		session, err := repository.FindByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if session.Status() != want {
			t.Fatalf("%s status = %q, want %q", id, session.Status(), want)
		}
	}

	if err := repository.RevokeAllForMember(ctx, "member-a", time.Now()); err != nil {
		t.Fatalf("revoke member: %v", err)
	}
	session, err := repository.FindByID(ctx, "sess_a2")
	if err != nil {
		t.Fatal(err)
	}
	if session.Status() != domain.StatusRevoked {
		t.Fatalf("member revoke missed sess_a2: %q", session.Status())
	}

	if _, err := repository.FindByID(ctx, "sess_missing"); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("missing session error = %v, want ErrSessionNotFound", err)
	}
}
