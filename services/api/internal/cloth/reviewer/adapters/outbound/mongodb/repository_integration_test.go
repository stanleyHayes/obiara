//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	reviewermongo "github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func key(number int) string { return fmt.Sprintf("%064x", number) }

func TestOTPSingleUseExpiryRevokeAndPrivacy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(context.Background())
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	database := client.Database("cloth_reviewer_test")
	repository := reviewermongo.NewRepository(database)
	if err = repository.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	pending := createReview(t, "review-1", "create-1", now, now.Add(10*time.Minute), now.Add(24*time.Hour))
	if err = repository.Create(ctx, pending); err != nil {
		t.Fatal(err)
	}
	if _, err = pending.Redeem(key(4), key(99), key(7), now.Add(time.Minute), command("wrong-otp", 1, now.Add(time.Minute))); !errors.Is(err, domain.ErrCredential) {
		t.Fatalf("wrong OTP=%v", err)
	}

	redeemed, err := pending.Redeem(key(4), key(5), key(7), now.Add(time.Minute), command("redeem", 1, now.Add(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	competing, err := pending.Redeem(key(4), key(5), key(8), now.Add(time.Minute), command("competing", 1, now.Add(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		command string
		review  domain.Review
		err     error
	}
	results := make(chan result, 2)
	go func() { results <- result{"redeem", redeemed, repository.Append(ctx, redeemed, 1, "redeem")} }()
	go func() { results <- result{"competing", competing, repository.Append(ctx, competing, 1, "competing")} }()
	successes, conflicts := 0, 0
	var winner result
	for range 2 {
		result := <-results
		switch {
		case result.err == nil:
			successes++
			winner = result
		case errors.Is(result.err, application.ErrOptimisticConflict):
			conflicts++
		default:
			t.Fatal(result.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("single use successes=%d conflicts=%d", successes, conflicts)
	}
	if err = repository.Append(ctx, winner.review, 1, winner.command); !errors.Is(err, application.ErrCommandApplied) {
		t.Fatalf("redemption replay=%v", err)
	}
	saved, err := repository.Find(ctx, "review-1")
	if err != nil {
		t.Fatal(err)
	}
	projection, err := saved.Project(saved.SessionDigest(), now.Add(2*time.Minute))
	if err != nil || projection.WatermarkRef != key(6) || len(projection.QuestionRefs) != 2 || len(projection.MaterialRefs) != 1 {
		t.Fatalf("projection=%+v error=%v", projection, err)
	}
	revoked, err := saved.Revoke(now.Add(3*time.Minute), command("revoke", 2, now.Add(3*time.Minute)))
	if err != nil || repository.Append(ctx, revoked, 2, "revoke") != nil {
		t.Fatalf("revoke=%v", err)
	}
	saved, err = repository.Find(ctx, "review-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = saved.Project(saved.SessionDigest(), now.Add(3*time.Minute)); !errors.Is(err, domain.ErrRevoked) {
		t.Fatalf("access after revoke=%v", err)
	}

	expiring := createReviewWithCredentials(t, "review-2", "create-2", key(14), key(15), now, now.Add(5*time.Minute), now.Add(time.Hour))
	if err = repository.Create(ctx, expiring); err != nil {
		t.Fatal(err)
	}
	if _, err = expiring.Redeem(key(14), key(15), key(9), now.Add(5*time.Minute), command("expired-otp", 1, now.Add(5*time.Minute))); !errors.Is(err, domain.ErrExpired) {
		t.Fatalf("expired OTP=%v", err)
	}
	inviteExpired := createReviewWithCredentials(t, "review-3", "create-3", key(24), key(25), now, now.Add(10*time.Minute), now.Add(10*time.Minute))
	if err = repository.Create(ctx, inviteExpired); err != nil {
		t.Fatal(err)
	}
	if _, err = inviteExpired.Redeem(key(24), key(25), key(10), now.Add(10*time.Minute), command("expired-invite", 1, now.Add(10*time.Minute))); !errors.Is(err, domain.ErrExpired) {
		t.Fatalf("expired invite=%v", err)
	}

	var raw bson.M
	if err = database.Collection("cloth_reviewer_access").FindOne(ctx, bson.M{"_id": "review-1"}).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	encoded, err := bson.MarshalExtJSON(raw, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"raw-invite-token", "123456", "raw-session-token", "member-private",
		"reviewer-private", "raw question body", "raw material body", "bearer",
	} {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(forbidden)) {
			t.Fatalf("stored reviewer access leaked %q: %s", forbidden, encoded)
		}
	}
	if _, exists := raw["sessionDigest"]; !exists {
		t.Fatalf("redeemed session digest is missing: %s", encoded)
	}
}

func createReview(t *testing.T, id, commandID string, now, otpExpiry, inviteExpiry time.Time) domain.Review {
	return createReviewWithCredentials(t, id, commandID, key(4), key(5), now, otpExpiry, inviteExpiry)
}

func createReviewWithCredentials(t *testing.T, id, commandID, inviteDigest, otpDigest string, now, otpExpiry, inviteExpiry time.Time) domain.Review {
	t.Helper()
	review, err := domain.Create(domain.State{
		ID: id, Members: []string{key(1), key(2)}, ReviewerKey: key(3),
		InviteDigest: inviteDigest, OTPDigest: otpDigest, WatermarkRef: key(6),
		QuestionRefs: []string{"question-1", "question-2"}, MaterialRefs: []string{"material-1"},
		OTPExpiresAt: otpExpiry, InviteExpiresAt: inviteExpiry,
	}, now, command(commandID, 0, now))
	if err != nil {
		t.Fatal(err)
	}
	return review
}

func command(id string, revision uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ExpectedRevision: revision, At: at}
}
