//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

const integrationTimeout = 3 * time.Minute

type codeSenderStub struct {
	codes map[string]string
}

func (stub *codeSenderStub) SendMfaCode(_ context.Context, email, code string) error {
	stub.codes[email] = code
	return nil
}

func TestAdminAuthEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_admin_test")
	principals := mongodb.NewPrincipalRepository(database)
	challenges := mongodb.NewChallengeRepository(database)
	sessions := mongodb.NewSessionRepository(database)
	for name, ensure := range map[string]func(context.Context) error{
		"principals": principals.EnsureIndexes, "challenges": challenges.EnsureIndexes, "sessions": sessions.EnsureIndexes,
	} {
		if err := ensure(ctx); err != nil {
			t.Fatalf("ensure %s: %v", name, err)
		}
	}

	sender := &codeSenderStub{codes: map[string]string{}}
	ids := func() func() string {
		counter := 0
		return func() string {
			counter++
			return "adm_" + strings.Repeat("z", counter)
		}
	}()
	service := application.NewAdminService(principals, challenges, sessions, mongodb.NewAuditStore(database), sender, time.Now, ids)

	// Bootstrap the root admin directly (production bootstrap is a
	// controlled migration, not an API path).
	root, err := domain.NewPrincipal("adm_root", "root@example.test", []domain.Role{domain.RoleAdmin}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := principals.Create(ctx, root); err != nil {
		t.Fatal(err)
	}
	rootSession := domain.ReconstituteSession(
		"sess_root", root.ID(), root.Roles(), true, time.Now().Add(time.Hour), false, 2, time.Now(),
	)
	if err := sessions.Create(ctx, rootSession); err != nil {
		t.Fatal(err)
	}

	// Enroll a verifier through the privileged path; audited.
	verifier, err := service.Enroll(ctx, rootSession.ID(), "verifier@example.test", []domain.Role{domain.RoleVerifier})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	auditCount, err := database.Collection("admin_access").CountDocuments(ctx, bson.M{"action": "admin.enroll"})
	if err != nil || auditCount != 1 {
		t.Fatalf("enroll audits = %d, want 1", auditCount)
	}

	// A verifier cannot enroll (least privilege, FR-801).
	verifierSession := domain.ReconstituteSession(
		"sess_verifier", verifier.ID(), verifier.Roles(), true, time.Now().Add(time.Hour), false, 2, time.Now(),
	)
	if err := sessions.Create(ctx, verifierSession); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Enroll(ctx, verifierSession.ID(), "sneaky@example.test", []domain.Role{domain.RoleAdmin}); err != application.ErrNotAdmin {
		t.Fatalf("verifier enroll = %v, want ErrNotAdmin", err)
	}

	// Admin-role grants are durable proposals and require a different
	// stepped-up administrator. Approval updates the proposal, principal and
	// immutable audit in one transaction.
	secondAdmin, err := domain.NewPrincipal("adm_second", "second@example.test", []domain.Role{domain.RoleAdmin}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := principals.Create(ctx, secondAdmin); err != nil {
		t.Fatal(err)
	}
	secondSession := domain.ReconstituteSession(
		"sess_second", secondAdmin.ID(), secondAdmin.Roles(), true, time.Now().Add(time.Hour), false, 2, time.Now(),
	)
	if err := sessions.Create(ctx, secondSession); err != nil {
		t.Fatal(err)
	}
	change, err := service.ProposeAdminRoleChange(
		ctx, rootSession.ID(), verifier.ID(),
		[]domain.Role{domain.RoleVerifier, domain.RoleAdmin},
		"Add verified command-centre coverage",
	)
	if err != nil {
		t.Fatalf("propose role change: %v", err)
	}
	if _, err := service.ApproveAdminRoleChange(ctx, rootSession.ID(), change.ID()); err != domain.ErrSameApprover {
		t.Fatalf("self approval = %v, want ErrSameApprover", err)
	}
	updatedVerifier, err := service.ApproveAdminRoleChange(ctx, secondSession.ID(), change.ID())
	if err != nil {
		t.Fatalf("approve role change: %v", err)
	}
	if !updatedVerifier.HasRole(domain.RoleAdmin) {
		t.Fatal("approved verifier must hold admin role")
	}
	roleAuditCount, err := database.Collection("admin_access").CountDocuments(
		ctx,
		bson.M{"target": change.ID()},
	)
	if err != nil || roleAuditCount != 2 {
		t.Fatalf("role-change audits = %d, want proposal + approval", roleAuditCount)
	}

	// Login: code delivered by email bridge, session issued.
	if err := service.StartLogin(ctx, "verifier@example.test"); err != nil {
		t.Fatal(err)
	}
	code := sender.codes["verifier@example.test"]
	if len(code) != 6 {
		t.Fatalf("no code delivered: %v", sender.codes)
	}
	session, err := service.CompleteLogin(ctx, "verifier@example.test", code)
	if err != nil {
		t.Fatalf("complete login: %v", err)
	}
	if !session.Active(time.Now()) || session.SteppedUp() {
		t.Fatalf("session = %#v", session)
	}

	// Step-up: new code, session flagged, audited.
	if err := service.StepUpStart(ctx, session.ID()); err != nil {
		t.Fatal(err)
	}
	stepCode := sender.codes["verifier@example.test"]
	if stepCode == code {
		t.Fatal("step-up must mint a fresh code")
	}
	stepped, err := service.StepUpComplete(ctx, session.ID(), stepCode)
	if err != nil {
		t.Fatalf("step-up: %v", err)
	}
	if !stepped.SteppedUp() {
		t.Fatal("session must be stepped up")
	}
	stepAudits, err := database.Collection("admin_access").CountDocuments(ctx, bson.M{"action": "admin.step_up"})
	if err != nil || stepAudits != 1 {
		t.Fatalf("step-up audits = %d, want 1", stepAudits)
	}

	// Wrong code increments attempts without a session.
	if err := service.StartLogin(ctx, "verifier@example.test"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteLogin(ctx, "verifier@example.test", "000000"); err != domain.ErrMfaMismatch {
		t.Fatalf("wrong code = %v, want mismatch", err)
	}
}
