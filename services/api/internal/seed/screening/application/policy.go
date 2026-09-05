package application

import (
	"errors"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	MaxTextRunes       = 1000
	MaxTextBytes       = 4000
	MaxMedia           = 4
	MaxMediaBytes      = 25 * 1024 * 1024
	MaxMediaDurationMs = 90_000
	MaxReasonCodes     = 4
)

type Status string

const (
	StatusApproved    Status = "approved"
	StatusRejected    Status = "rejected"
	StatusUncertain   Status = "uncertain"
	StatusHumanReview Status = "human_review"
)

type ReasonCode string

const (
	ReasonClear               ReasonCode = "clear"
	ReasonContactExfiltration ReasonCode = "contact_exfiltration"
	ReasonPaymentRequest      ReasonCode = "payment_request"
	ReasonSexualContent       ReasonCode = "sexual_content"
	ReasonThreat              ReasonCode = "threat"
	ReasonUnsupportedLocale   ReasonCode = "unsupported_locale"
	ReasonUnsupportedMedia    ReasonCode = "unsupported_media"
	ReasonProviderFailure     ReasonCode = "provider_failure"
	ReasonUncertain           ReasonCode = "uncertain"
)

var (
	// ErrHumanReviewRequired is deliberately the sow context's own error
	// value rather than a new one with the same message. errors.Is compares
	// identity: two separate errors.New calls that read alike do not match,
	// and the sow service decides whether to hold a sow by testing exactly
	// this. A private copy here would send every reviewable sow down the
	// "service unavailable" path — the bug the held state was added to fix.
	ErrHumanReviewRequired = sowapplication.ErrHumanReviewRequired
	ErrInvalidInput        = errors.New("invalid sow screening input")
)

var localePattern = regexp.MustCompile(`^[a-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$`)
var opaquePattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

var allowedMIMEs = []string{
	"audio/mpeg",
	"audio/mp4",
	"audio/ogg",
	"audio/webm",
	"image/jpeg",
	"image/png",
}

var advisoryReasons = []ReasonCode{
	ReasonClear,
	ReasonContactExfiltration,
	ReasonPaymentRequest,
	ReasonSexualContent,
	ReasonThreat,
	ReasonUncertain,
}

type ScreeningInput struct {
	Text          string
	LocaleTag     string
	LocaleVersion uint64
	Media         []MediaMetadata
}

type Advisory struct {
	Status     Status
	Reasons    []ReasonCode
	Confidence int
}

type Adjudication struct {
	Status        Status
	Reason        ReasonCode
	Reference     string
	HumanReviewed bool
}

type ReviewCase struct {
	ID       string
	Input    ScreeningInput
	Reason   ReasonCode
	Advisory *Advisory
}

func normalizeText(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", ErrInvalidInput
	}
	normalized := strings.Join(strings.Fields(norm.NFKC.String(strings.TrimSpace(value))), " ")
	if normalized == "" || utf8.RuneCountInString(normalized) > MaxTextRunes || len([]byte(normalized)) > MaxTextBytes {
		return "", ErrInvalidInput
	}
	return normalized, nil
}

func validLocale(review LocaleReview) bool {
	return localePattern.MatchString(review.Tag) && review.Version > 0 && review.Reviewed && !review.ReviewedAt.IsZero()
}

func validMedia(metadata MediaMetadata) bool {
	return slices.Contains(allowedMIMEs, metadata.MIME) &&
		metadata.Bytes > 0 && metadata.Bytes <= MaxMediaBytes &&
		metadata.DurationMs >= 0 && metadata.DurationMs <= MaxMediaDurationMs
}

func validAdvisory(advisory Advisory) bool {
	if advisory.Status != StatusApproved && advisory.Status != StatusRejected && advisory.Status != StatusUncertain ||
		advisory.Confidence < 0 || advisory.Confidence > 100 ||
		len(advisory.Reasons) == 0 || len(advisory.Reasons) > MaxReasonCodes {
		return false
	}
	seen := make(map[ReasonCode]struct{}, len(advisory.Reasons))
	for _, reason := range advisory.Reasons {
		if !slices.Contains(advisoryReasons, reason) {
			return false
		}
		if _, duplicate := seen[reason]; duplicate {
			return false
		}
		seen[reason] = struct{}{}
	}
	return true
}

func validAdjudication(adjudication Adjudication) bool {
	if !adjudication.HumanReviewed || !opaquePattern.MatchString(adjudication.Reference) {
		return false
	}
	switch adjudication.Status {
	case StatusApproved:
		return adjudication.Reason == ReasonClear
	case StatusRejected:
		return slices.Contains([]ReasonCode{
			ReasonContactExfiltration,
			ReasonPaymentRequest,
			ReasonSexualContent,
			ReasonThreat,
		}, adjudication.Reason)
	default:
		return false
	}
}
