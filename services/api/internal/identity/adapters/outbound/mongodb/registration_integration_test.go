//go:build integration

package mongodb_test

import (
	"context"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

const integrationTimeout = 3 * time.Minute

func TestOtpRegistrationEndToEnd(t *testing.T) {
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

	database := client.Database("obiara_registration_test")
	sessionRepository := mongodb.NewRepository(database)
	challengeRepository := mongodb.NewOtpChallengeRepository(database)
	accountRepository := mongodb.NewAccountRepository(database)
	for name, ensure := range map[string]func(context.Context) error{
		"sessions":   sessionRepository.EnsureIndexes,
		"challenges": challengeRepository.EnsureIndexes,
		"accounts":   accountRepository.EnsureIndexes,
	} {
		if err := ensure(ctx); err != nil {
			t.Fatalf("ensure %s indexes: %v", name, err)
		}
	}

	ids := func() string { return "id_" + strings.ToLower(domain.HashToken(time.Now().String())[:16]) }
	sender := simulator.NewSender()
	sessions := application.NewSessionService(sessionRepository, time.Now, ids)
	registration := application.NewRegistrationService(challengeRepository, accountRepository, sender, sessions, time.Now, ids)

	phone := "+233550000101"

	// Request → code delivered via the simulator sender.
	request, err := registration.RequestOtp(ctx, phone)
	if err != nil {
		t.Fatalf("request otp: %v", err)
	}
	if request.ChallengeID == "" {
		t.Fatal("challenge id missing")
	}
	code, ok := sender.LastCode(phone)
	if !ok || len(code) != 6 {
		t.Fatalf("simulator code missing for %s", phone)
	}

	// Immediate resend is rate limited.
	if _, err := registration.RequestOtp(ctx, phone); err != domain.ErrOtpRateLimited {
		t.Fatalf("immediate resend = %v, want rate limited", err)
	}

	// Wrong code consumes an attempt and does not verify.
	if _, err := registration.VerifyOtp(ctx, phone, "000000", "device-1"); err != domain.ErrOtpMismatch {
		t.Fatalf("wrong code = %v, want mismatch", err)
	}

	// Correct code: account created, session issued and persisted.
	issued, err := registration.VerifyOtp(ctx, phone, code, "device-1")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if issued.AccessToken == "" || issued.RefreshToken == "" {
		t.Fatal("tokens missing")
	}
	account, err := accountRepository.FindByPhone(ctx, phone)
	if err != nil {
		t.Fatalf("account not created: %v", err)
	}
	if account.ID() != issued.Session.MemberID() {
		t.Fatalf("session member %q != account %q", issued.Session.MemberID(), account.ID())
	}
	if _, err := sessionRepository.FindByID(ctx, issued.Session.ID()); err != nil {
		t.Fatalf("session not persisted: %v", err)
	}

	// The same code cannot verify twice.
	if _, err := registration.VerifyOtp(ctx, phone, code, "device-1"); err != domain.ErrOtpConsumed {
		t.Fatalf("reused code = %v, want consumed", err)
	}
}
