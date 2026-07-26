package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewChecksum(t *testing.T) {
	t.Parallel()
	valid := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	for _, test := range []struct {
		name      string
		algorithm string
		value     string
		wantErr   bool
	}{
		{name: "sha256", algorithm: " SHA256 ", value: valid},
		{name: "unknown algorithm", algorithm: "md5", value: valid, wantErr: true},
		{name: "wrong length", algorithm: "sha256", value: "aa", wantErr: true},
		{name: "non hexadecimal", algorithm: "sha256", value: valid[:63] + "z", wantErr: true},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			checksum, err := NewChecksum(test.algorithm, test.value)
			if (err != nil) != test.wantErr {
				t.Fatalf("NewChecksum() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && (checksum.Algorithm() != "sha256" || checksum.Value() != valid) {
				t.Fatalf("unexpected checksum: %s:%s", checksum.Algorithm(), checksum.Value())
			}
		})
	}
}

func TestAssetRetentionAndLegalHold(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

	for _, test := range []struct {
		name      string
		retention Retention
		deleteAt  time.Time
		wantErr   error
	}{
		{name: "unretained", deleteAt: now},
		{name: "retention active", retention: NewRetention(now.Add(time.Hour), false), deleteAt: now, wantErr: ErrRetentionActive},
		{name: "retention elapsed", retention: NewRetention(now.Add(time.Hour), false), deleteAt: now.Add(2 * time.Hour)},
		{name: "legal hold prevents deletion", retention: NewRetention(time.Time{}, true), deleteAt: now, wantErr: ErrLegalHoldActive},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			asset := validAsset(t, now, time.Time{}, test.retention)
			deleted, err := asset.MarkDeleted(test.deleteAt)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("MarkDeleted() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				if asset.IsDeleted() {
					t.Fatal("original asset was mutated")
				}
				return
			}
			if !deleted.IsDeleted() || asset.IsDeleted() {
				t.Fatal("deletion must produce a new value without mutating original")
			}
		})
	}
}

func TestAssetAvailability(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	expiring := validAsset(t, now, now.Add(time.Hour), Retention{})
	if !expiring.AvailableAt(now) {
		t.Fatal("asset should be available before expiry")
	}
	if expiring.AvailableAt(now.Add(time.Hour)) {
		t.Fatal("asset must be unavailable at expiry")
	}
	deleted, err := expiring.MarkDeleted(now)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.AvailableAt(now) {
		t.Fatal("deleted asset must be unavailable")
	}
}

func TestAssetMetadataIsImmutable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	asset := validAsset(t, now, time.Time{}, Retention{})
	checksum := asset.Checksum()
	checksum.value = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	retention := asset.Retention()
	retention.legalHold = true

	if asset.Checksum().Value() == checksum.Value() || asset.Retention().LegalHold() {
		t.Fatal("returned metadata must not mutate the asset")
	}
}

func validAsset(t *testing.T, createdAt, expiresAt time.Time, retention Retention) Asset {
	t.Helper()
	checksum, err := NewChecksum("sha256", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	asset, err := NewAsset(NewAssetParams{
		ID: "asset-1", ObjectKey: "owners/member-1/asset-1", OwnerID: "member-1",
		ContentType: "image/jpeg", Size: 128, Checksum: checksum,
		CreatedAt: createdAt, ExpiresAt: expiresAt, Retention: retention,
	})
	if err != nil {
		t.Fatal(err)
	}
	return asset
}
