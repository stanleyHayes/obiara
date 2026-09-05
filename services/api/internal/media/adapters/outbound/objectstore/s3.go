// Package objectstore signs time-limited URLs for an S3-compatible bucket.
//
// It implements the media context's Signer port and nothing else: the service
// decides who may act, what the object is called and how long the grant lives,
// and this only turns that decision into a URL the storage provider accepts.
// No vendor type crosses the port.
//
// Signature Version 4 is computed here from the standard library rather than
// pulled in with a vendor SDK. The algorithm is a published specification and
// about a hundred lines of HMAC; an SDK would add a large dependency tree to a
// service whose release gate counts reachable advisories, and would tie the
// adapter to one provider. As written this signs for AWS S3, Cloudflare R2,
// Backblaze B2 and MinIO without changing a line.
package objectstore

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/media/application"
)

const (
	algorithm     = "AWS4-HMAC-SHA256"
	service       = "s3"
	terminator    = "aws4_request"
	unsignedBody  = "UNSIGNED-PAYLOAD"
	isoLayout     = "20060102T150405Z"
	dateLayout    = "20060102"
	maxPresignTTL = 7 * 24 * time.Hour // the protocol's own ceiling
	minPresignTTL = time.Second
)

var (
	ErrConfiguration = errors.New("object store configuration is incomplete")
	ErrExpiry        = errors.New("object store grant expiry is out of range")
	ErrObjectKey     = errors.New("object key is not usable")
)

// Config describes one bucket. Endpoint is the provider's host — leave it
// empty for AWS S3 proper, which is derived from the region.
type Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	// PathStyle addresses the bucket as a path segment rather than a
	// subdomain. R2 and MinIO require it; AWS S3 accepts either.
	PathStyle bool
}

type Signer struct {
	config Config
	host   string
	scheme string
	prefix string
	now    func() time.Time
	// client performs the one operation that cannot be handed to the browser.
	// Deleting is the server's own act; a presigned delete URL sent to a
	// client would be a standing licence to erase somebody's recording.
	client *http.Client
}

// NewSigner validates the configuration once, at startup, so a missing secret
// is a boot failure rather than a member's upload failing at the moment they
// press record.
func NewSigner(config Config, now func() time.Time) (*Signer, error) {
	config.Region = strings.TrimSpace(config.Region)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKey = strings.TrimSpace(config.AccessKey)
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	if config.Region == "" || config.Bucket == "" ||
		config.AccessKey == "" || config.SecretKey == "" {
		return nil, ErrConfiguration
	}
	if now == nil {
		now = time.Now
	}

	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://s3.%s.amazonaws.com", config.Region)
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" {
		return nil, ErrConfiguration
	}

	signer := &Signer{
		config: config, scheme: parsed.Scheme, now: now,
		client: &http.Client{Timeout: 20 * time.Second},
	}
	if config.PathStyle {
		signer.host = parsed.Host
		signer.prefix = "/" + config.Bucket
	} else {
		signer.host = config.Bucket + "." + parsed.Host
	}
	// A trailing path on the endpoint (MinIO behind a sub-path) is kept.
	signer.prefix = strings.TrimSuffix(parsed.Path, "/") + signer.prefix
	return signer, nil
}

func (signer *Signer) SignUpload(
	_ context.Context,
	request application.UploadSigningRequest,
) (application.SignedAccess, error) {
	extra := url.Values{}
	headers := map[string]string{"host": signer.host}

	// Signed, not merely sent: a header inside the signature cannot be
	// changed by whoever holds the URL. Content type and length are part of
	// the grant, so a URL issued for a 40-second Opus clip cannot be reused
	// to push an arbitrary file.
	if contentType := strings.TrimSpace(request.ContentType); contentType != "" {
		headers["content-type"] = contentType
	}
	if request.Size > 0 {
		headers["content-length"] = strconv.FormatInt(request.Size, 10)
	}
	// The digest the provider must verify before it accepts the bytes. S3
	// takes it base64-encoded; the domain holds it as hex.
	if value := request.Checksum.Value(); value != "" {
		raw, err := hex.DecodeString(value)
		if err != nil {
			return application.SignedAccess{}, ErrObjectKey
		}
		headers["x-amz-checksum-sha256"] = base64.StdEncoding.EncodeToString(raw)
	}

	return signer.presign("PUT", request.ObjectKey, request.ExpiresAt, headers, extra)
}

func (signer *Signer) SignRead(
	_ context.Context,
	request application.ReadSigningRequest,
) (application.SignedAccess, error) {
	return signer.presign(
		"GET", request.ObjectKey, request.ExpiresAt,
		map[string]string{"host": signer.host}, url.Values{},
	)
}

// Delete removes the object for good.
//
// It signs a DELETE the same way every other request is signed and then makes
// the call itself, rather than handing a URL to anyone. Erasure is what makes
// "withdraw this recording" true, so it must not depend on a client choosing
// to follow through.
//
// A missing object counts as deleted. Retrying a partial purge is normal, and
// failing the second attempt because the first succeeded would strand the
// aggregate in purge_pending forever.
func (signer *Signer) Delete(ctx context.Context, objectKey string) error {
	access, err := signer.presign(
		"DELETE", objectKey, signer.now().UTC().Add(time.Minute),
		map[string]string{"host": signer.host}, url.Values{},
	)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, access.URL, nil)
	if err != nil {
		return err
	}
	response, err := signer.client.Do(request)
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	}()
	// S3 answers 204 for a delete and 404 when it was already gone; both mean
	// the object is not there any more, which is the whole ask.
	if response.StatusCode == http.StatusNoContent ||
		response.StatusCode == http.StatusOK ||
		response.StatusCode == http.StatusNotFound {
		return nil
	}
	return fmt.Errorf("delete object: storage answered %d", response.StatusCode)
}

func (signer *Signer) presign(
	method, objectKey string,
	expiresAt time.Time,
	headers map[string]string,
	extra url.Values,
) (application.SignedAccess, error) {
	objectKey = strings.TrimPrefix(strings.TrimSpace(objectKey), "/")
	if objectKey == "" || strings.Contains(objectKey, "..") {
		return application.SignedAccess{}, ErrObjectKey
	}

	now := signer.now().UTC()
	ttl := expiresAt.UTC().Sub(now)
	if ttl < minPresignTTL || ttl > maxPresignTTL {
		return application.SignedAccess{}, ErrExpiry
	}

	stamp := now.Format(isoLayout)
	day := now.Format(dateLayout)
	scope := strings.Join([]string{day, signer.config.Region, service, terminator}, "/")

	signedHeaders, canonicalHeaders := canonicalizeHeaders(headers)

	query := url.Values{}
	for key, values := range extra {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	query.Set("X-Amz-Algorithm", algorithm)
	query.Set("X-Amz-Credential", signer.config.AccessKey+"/"+scope)
	query.Set("X-Amz-Date", stamp)
	query.Set("X-Amz-Expires", strconv.Itoa(int(ttl.Seconds())))
	query.Set("X-Amz-SignedHeaders", signedHeaders)

	canonicalURI := signer.prefix + "/" + encodePath(objectKey)
	canonicalRequest := strings.Join([]string{
		method,
		canonicalURI,
		encodeQuery(query),
		canonicalHeaders,
		signedHeaders,
		unsignedBody,
	}, "\n")

	stringToSign := strings.Join([]string{
		algorithm,
		stamp,
		scope,
		hex.EncodeToString(sha256Sum([]byte(canonicalRequest))),
	}, "\n")

	signature := hex.EncodeToString(
		hmacSHA256(signer.signingKey(day), []byte(stringToSign)),
	)
	query.Set("X-Amz-Signature", signature)

	return application.SignedAccess{
		URL:       signer.scheme + "://" + signer.host + canonicalURI + "?" + encodeQuery(query),
		ExpiresAt: now.Add(ttl),
	}, nil
}

// signingKey derives the date/region/service-scoped key. Deriving it per
// request rather than caching keeps the secret out of long-lived state for
// the sake of a few microseconds of HMAC.
func (signer *Signer) signingKey(day string) []byte {
	key := hmacSHA256([]byte("AWS4"+signer.config.SecretKey), []byte(day))
	key = hmacSHA256(key, []byte(signer.config.Region))
	key = hmacSHA256(key, []byte(service))
	return hmacSHA256(key, []byte(terminator))
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func sha256Sum(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

// canonicalizeHeaders lower-cases names, trims values and sorts by name, which
// is what the signature is computed over.
func canonicalizeHeaders(headers map[string]string) (string, string) {
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, strings.ToLower(strings.TrimSpace(name)))
	}
	sort.Strings(names)

	var canonical strings.Builder
	for _, name := range names {
		canonical.WriteString(name)
		canonical.WriteString(":")
		canonical.WriteString(strings.TrimSpace(headers[name]))
		canonical.WriteString("\n")
	}
	return strings.Join(names, ";"), canonical.String()
}

// encodePath percent-encodes each segment. Slashes separate segments and stay
// literal; everything else follows RFC 3986, which is stricter than Go's
// url.PathEscape (a space must be %20, never +).
func encodePath(key string) string {
	segments := strings.Split(key, "/")
	for index, segment := range segments {
		segments[index] = encodeRFC3986(segment)
	}
	return strings.Join(segments, "/")
}

// encodeQuery sorts by key and encodes to the same rule, because Go's
// url.Values.Encode leaves characters AWS expects escaped.
func encodeQuery(query url.Values) string {
	keys := make([]string, 0, len(query))
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		values := append([]string(nil), query[key]...)
		sort.Strings(values)
		for _, value := range values {
			parts = append(parts, encodeRFC3986(key)+"="+encodeRFC3986(value))
		}
	}
	return strings.Join(parts, "&")
}

func encodeRFC3986(value string) string {
	const unreserved = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~"
	var encoded strings.Builder
	for _, b := range []byte(value) {
		if strings.IndexByte(unreserved, b) >= 0 {
			encoded.WriteByte(b)
			continue
		}
		encoded.WriteString(fmt.Sprintf("%%%02X", b))
	}
	return encoded.String()
}
