package media

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

type assetStub struct {
	assets map[string]mediadomain.Asset
	err    error
}

func (stub assetStub) FindByID(_ context.Context, id string) (mediadomain.Asset, error) {
	if stub.err != nil {
		return mediadomain.Asset{}, stub.err
	}
	asset, found := stub.assets[id]
	if !found {
		return mediadomain.Asset{}, errors.New("not found")
	}
	return asset, nil
}

func asset(t *testing.T, id, owner string) mediadomain.Asset {
	t.Helper()
	checksum, err := mediadomain.NewChecksum("sha256", strings.Repeat("a", 64))
	if err != nil {
		t.Fatal(err)
	}
	built, err := mediadomain.NewAsset(mediadomain.NewAssetParams{
		ID: id, ObjectKey: "objects/" + id, OwnerID: owner,
		ContentType: "audio/ogg", Size: 2048, Checksum: checksum,
		Duration: 45 * time.Second, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return built
}

func TestOnlyYourOwnRecordingsAreYours(t *testing.T) {
	stub := assetStub{assets: map[string]mediadomain.Asset{
		"mine":   asset(t, "mine", "member-1"),
		"theirs": asset(t, "theirs", "member-2"),
	}}
	ownership := NewOwnership(stub)

	owned, err := ownership.OwnedBy(context.Background(), "member-1", []string{"mine"})
	if err != nil || !owned {
		t.Fatalf("owned = %v, err = %v", owned, err)
	}
	if owned, _ := ownership.OwnedBy(context.Background(), "member-1", []string{"theirs"}); owned {
		t.Fatal("another member's recording was treated as this member's")
	}
	// One borrowed recording among their own is still a borrowed recording.
	if owned, _ := ownership.OwnedBy(context.Background(), "member-1", []string{"mine", "theirs"}); owned {
		t.Fatal("a mixed set passed as entirely this member's")
	}
}

func TestARecordingNobodyCanAccountForIsNotYours(t *testing.T) {
	// Treating an unreadable asset as somebody's own would let a member
	// attach a reference to a recording nothing can describe, which is the
	// case the check exists for. It answers "no", not "broken": reporting a
	// fault would refuse the sow with an outage when the truthful answer is
	// a refusal.
	ownership := NewOwnership(assetStub{err: errors.New("mongo unavailable")})
	owned, err := ownership.OwnedBy(context.Background(), "member-1", []string{"anything"})
	if err != nil {
		t.Fatalf("an unreadable asset reported a fault: %v", err)
	}
	if owned {
		t.Fatal("an unreadable asset was treated as owned")
	}
}

func TestAnUncomposedOwnershipCheckIsAFaultNotAPass(t *testing.T) {
	if _, err := (Ownership{}).OwnedBy(context.Background(), "member-1", []string{"ref"}); err == nil {
		t.Fatal("an uncomposed check answered instead of failing")
	}
}

func TestNoRecordingsIsTriviallyYourOwn(t *testing.T) {
	owned, err := NewOwnership(assetStub{}).OwnedBy(context.Background(), "member-1", nil)
	if err != nil || !owned {
		t.Fatalf("owned = %v, err = %v", owned, err)
	}
}
