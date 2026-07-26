//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	camongo "github.com/stanleyHayes/obiara/services/api/internal/communityaudit/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/application"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/domain"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
	"time"
)

type auth struct{}

func (auth) Authorize(_ context.Context, a string, _ application.Capability) error {
	if a != "admin:1" {
		return errors.New("denied")
	}
	return nil
}

type mfa bool

func (m mfa) Recent(context.Context, string, time.Time) bool { return bool(m) }
func TestAuthorizationRedactionAuditAndRawIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	ct, e := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if e != nil {
		t.Fatal(e)
	}
	defer ct.Terminate(context.Background())
	u, _ := ct.ConnectionString(ctx)
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	cl, e := apimongo.Connect(ctx, u+sep+"directConnection=true")
	if e != nil {
		t.Fatal(e)
	}
	defer cl.Disconnect(context.Background())
	db := cl.Database("community_audit_test")
	r := camongo.New(db)
	if e = r.EnsureIndexes(ctx); e != nil {
		t.Fatal(e)
	}
	n := time.Now().UTC()
	c, _ := domain.New("case:1", domain.KindCircle, "circle:opaque", "evidence:opaque", "legitimacy.review", n)
	if e = r.Seed(ctx, c); e != nil {
		t.Fatal(e)
	}
	k, _ := privacy.New([]byte(strings.Repeat("k", 32)))
	s := application.New(auth{}, mfa(false), k, r, func() time.Time { return n })
	if _, e = s.Evidence(ctx, "admin:1", "case:1", "legitimacy_review", "corr:raw"); !errors.Is(e, application.ErrMFA) {
		t.Fatal("MFA bypass")
	}
	s = application.New(auth{}, mfa(true), k, r, func() time.Time { return n })
	q, e := s.Queue(ctx, "admin:1", 10)
	if e != nil || len(q) != 1 {
		t.Fatal(q, e)
	}
	if _, e = s.Decide(ctx, "admin:1", "case:1", "decision:1", "evidence_verified", "corr:raw", true, 1); e != nil {
		t.Fatal(e)
	}
	var d bson.M
	if e = db.Collection("community_audit_cases").FindOne(ctx, bson.M{"_id": "case:1"}).Decode(&d); e != nil {
		t.Fatal(e)
	}
	raw, _ := bson.MarshalExtJSON(d, false, false)
	for _, x := range []string{"admin:1", "corr:raw"} {
		if strings.Contains(string(raw), x) {
			t.Fatalf("raw id leaked %s", raw)
		}
	}
}
