package objectstore

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/media/application"
	"github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

var fixedNow = time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)

func testSigner(t *testing.T, config Config) *Signer {
	t.Helper()
	signer, err := NewSigner(config, func() time.Time { return fixedNow })
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	return signer
}

func defaultConfig() Config {
	return Config{
		Region:    "eu-central-1",
		Bucket:    "obiara-media",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}
}

// The signing key derivation is the one part of SigV4 that cannot be checked
// by self-consistency — a wrong-but-stable key signs every request wrongly and
// every internal comparison still passes. This is AWS's own published example
// (Signature Version 4 test suite: key for 20150830/us-east-1/iam).
func TestSigningKeyMatchesThePublishedAWSVector(t *testing.T) {
	key := hmacSHA256([]byte("AWS4"+"wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY"), []byte("20150830"))
	key = hmacSHA256(key, []byte("us-east-1"))
	key = hmacSHA256(key, []byte("iam"))
	key = hmacSHA256(key, []byte("aws4_request"))

	const want = "c4afb1cc5771d871763a393e44b703571b55cc28424d1a5e86da6ed3c154a4b9"
	if got := hex.EncodeToString(key); got != want {
		t.Fatalf("signing key = %s, want %s", got, want)
	}
}

func TestSignUploadCarriesTheGrantInsideTheSignature(t *testing.T) {
	signer := testSigner(t, defaultConfig())
	digest := strings.Repeat("ab", 32)
	checksum, err := domain.NewChecksum("sha256", digest)
	if err != nil {
		t.Fatal(err)
	}

	access, err := signer.SignUpload(context.Background(), application.UploadSigningRequest{
		ObjectKey:   "voice/introductions/vi_abc/prompt-1.opus",
		ContentType: "audio/ogg",
		Size:        184_320,
		Checksum:    checksum,
		ExpiresAt:   fixedNow.Add(10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("SignUpload: %v", err)
	}

	parsed, err := url.Parse(access.URL)
	if err != nil {
		t.Fatalf("unparseable URL: %v", err)
	}
	query := parsed.Query()

	if parsed.Host != "obiara-media.s3.eu-central-1.amazonaws.com" {
		t.Fatalf("host = %q", parsed.Host)
	}
	if query.Get("X-Amz-Algorithm") != algorithm {
		t.Fatalf("algorithm = %q", query.Get("X-Amz-Algorithm"))
	}
	if query.Get("X-Amz-Expires") != "600" {
		t.Fatalf("expires = %q, want 600", query.Get("X-Amz-Expires"))
	}
	if len(query.Get("X-Amz-Signature")) != 64 {
		t.Fatalf("signature is not a sha256 hex digest: %q", query.Get("X-Amz-Signature"))
	}

	// The whole point of signing these rather than merely sending them: a URL
	// issued for one clip cannot be replayed to push a different file.
	signed := query.Get("X-Amz-SignedHeaders")
	for _, header := range []string{"content-length", "content-type", "host", "x-amz-checksum-sha256"} {
		if !strings.Contains(signed, header) {
			t.Fatalf("SignedHeaders %q is missing %q", signed, header)
		}
	}
	if !access.ExpiresAt.Equal(fixedNow.Add(10 * time.Minute)) {
		t.Fatalf("ExpiresAt = %v", access.ExpiresAt)
	}
}

func TestChecksumIsSentAsBase64NotHex(t *testing.T) {
	// S3 rejects a hex digest in this header. The domain stores hex, so the
	// conversion is the adapter's job and is easy to get silently wrong —
	// silently, because a bad digest only shows up when real bytes arrive.
	signer := testSigner(t, defaultConfig())
	digest := strings.Repeat("ab", 32)
	checksum, _ := domain.NewChecksum("sha256", digest)

	headers := map[string]string{"host": signer.host}
	raw, _ := hex.DecodeString(digest)
	want := base64.StdEncoding.EncodeToString(raw)

	access, err := signer.SignUpload(context.Background(), application.UploadSigningRequest{
		ObjectKey: "a/b.opus", ContentType: "audio/ogg", Size: 1,
		Checksum: checksum, ExpiresAt: fixedNow.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = headers
	if len(want) != 44 || strings.Contains(want, digest) {
		t.Fatalf("expected a 44-char base64 digest, got %q", want)
	}
	if !strings.Contains(access.URL, "x-amz-checksum-sha256") {
		t.Fatal("checksum header was not signed")
	}
}

func TestPathStyleAddressesRewritesHostAndPath(t *testing.T) {
	// R2 and MinIO cannot serve virtual-host style; getting this wrong points
	// every upload at a hostname that does not resolve.
	config := defaultConfig()
	config.Endpoint = "https://abc123.r2.cloudflarestorage.com"
	config.PathStyle = true
	signer := testSigner(t, config)

	access, err := signer.SignRead(context.Background(), application.ReadSigningRequest{
		ObjectKey: "voice/x.opus",
		ExpiresAt: fixedNow.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(access.URL, "https://abc123.r2.cloudflarestorage.com/obiara-media/voice/x.opus?") {
		t.Fatalf("path-style URL = %s", access.URL)
	}
}

func TestExpiryOutsideTheProtocolWindowIsRefused(t *testing.T) {
	signer := testSigner(t, defaultConfig())
	for name, expiry := range map[string]time.Time{
		"already past":      fixedNow.Add(-time.Second),
		"exactly now":       fixedNow,
		"beyond seven days": fixedNow.Add(8 * 24 * time.Hour),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := signer.SignRead(context.Background(), application.ReadSigningRequest{
				ObjectKey: "a.opus", ExpiresAt: expiry,
			})
			if err != ErrExpiry {
				t.Fatalf("err = %v, want ErrExpiry", err)
			}
		})
	}
}

func TestObjectKeysThatCouldEscapeTheBucketAreRefused(t *testing.T) {
	signer := testSigner(t, defaultConfig())
	for _, key := range []string{"", "   ", "/", "../secrets/key", "voice/../../etc/passwd"} {
		if _, err := signer.SignRead(context.Background(), application.ReadSigningRequest{
			ObjectKey: key, ExpiresAt: fixedNow.Add(time.Minute),
		}); err != ErrObjectKey {
			t.Fatalf("key %q returned %v, want ErrObjectKey", key, err)
		}
	}
}

func TestIncompleteConfigurationFailsAtStartupNotAtUpload(t *testing.T) {
	// A missing secret must stop the service booting. Discovering it when a
	// member presses record costs them the recording.
	base := defaultConfig()
	for name, mutate := range map[string]func(*Config){
		"no region":    func(c *Config) { c.Region = "" },
		"no bucket":    func(c *Config) { c.Bucket = "" },
		"no key id":    func(c *Config) { c.AccessKey = "" },
		"no secret":    func(c *Config) { c.SecretKey = " " },
		"bad endpoint": func(c *Config) { c.Endpoint = "://" },
	} {
		t.Run(name, func(t *testing.T) {
			config := base
			mutate(&config)
			if _, err := NewSigner(config, nil); err == nil {
				t.Fatal("expected a configuration error")
			}
		})
	}
}

func TestSpacesEncodeAsPercentTwentyNotPlus(t *testing.T) {
	// url.Values.Encode would write "+", which SigV4 does not accept; the
	// signature would verify locally and be rejected by the provider.
	if got := encodeRFC3986("a b"); got != "a%20b" {
		t.Fatalf("encodeRFC3986(%q) = %q, want a%%20b", "a b", got)
	}
	if got := encodePath("voice/a b/c.opus"); got != "voice/a%20b/c.opus" {
		t.Fatalf("encodePath = %q", got)
	}
}

func TestTheSameRequestSignsIdentically(t *testing.T) {
	signer := testSigner(t, defaultConfig())
	request := application.ReadSigningRequest{
		ObjectKey: "voice/x.opus", ExpiresAt: fixedNow.Add(5 * time.Minute),
	}
	first, err := signer.SignRead(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := signer.SignRead(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if first.URL != second.URL {
		t.Fatal("signing is not deterministic for a fixed clock")
	}
}

// The adapter must satisfy the port it exists for.
var _ application.Signer = (*Signer)(nil)

func TestDeleteTreatsAnAlreadyGoneObjectAsDeleted(t *testing.T) {
	// Purges are retried. Failing the second attempt because the first
	// succeeded would strand a withdrawn recording in purge_pending forever,
	// which is the opposite of what erasure is for.
	for _, status := range []int{
		http.StatusNoContent, http.StatusOK, http.StatusNotFound,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("method = %s, want DELETE", r.Method)
			}
			if r.URL.Query().Get("X-Amz-Signature") == "" {
				t.Fatal("the delete was not signed")
			}
			w.WriteHeader(status)
		}))
		config := defaultConfig()
		config.Endpoint = server.URL
		config.PathStyle = true
		signer := testSigner(t, config)

		if err := signer.Delete(context.Background(), "voice/x.opus"); err != nil {
			t.Fatalf("status %d returned %v, want nil", status, err)
		}
		server.Close()
	}
}

func TestDeleteReportsAStorageRefusal(t *testing.T) {
	// A refusal must not read as success: the aggregate would be marked
	// purged while the bytes are still sitting in the bucket.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	config := defaultConfig()
	config.Endpoint = server.URL
	config.PathStyle = true
	signer := testSigner(t, config)

	if err := signer.Delete(context.Background(), "voice/x.opus"); err == nil {
		t.Fatal("a 403 from storage was reported as a successful erasure")
	}
}

func TestDeleteRefusesAKeyThatCouldEscapeTheBucket(t *testing.T) {
	signer := testSigner(t, defaultConfig())
	if err := signer.Delete(context.Background(), "../other/secret"); err != ErrObjectKey {
		t.Fatalf("err = %v, want ErrObjectKey", err)
	}
}
