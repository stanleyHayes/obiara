package domain

import (
	"errors"
	"strings"
	"time"
)

// Payout is money leaving the platform.
//
// It is the first outbound money this product has ever sent, and it is
// deliberately slow: an affiliate requests, an operator approves with a
// step-up, and only then is a transfer initiated. A scheduled job paying
// everybody automatically would put real money in real hands with nobody in
// the loop, and the first bug would be the expensive kind.
type PayoutStatus string

const (
	PayoutRequested PayoutStatus = "requested"
	PayoutApproved  PayoutStatus = "approved"
	PayoutSent      PayoutStatus = "sent"
	PayoutRefused   PayoutStatus = "refused"
	PayoutFailed    PayoutStatus = "failed"
)

var (
	ErrInvalidPayout    = errors.New("invalid payout")
	ErrPayoutTransition = errors.New("invalid payout transition")
	ErrBelowMinimum     = errors.New("that is below the payout minimum")
	ErrWithholdingUnset = errors.New("a withholding rate must be configured before anything is paid")
)

// Payout is one request to be paid.
//
// Gross, withholding and net are all stored rather than derived at read time.
// A rate that changes must not silently restate what somebody was already
// paid, and a filing has to be able to show the number that was actually
// withheld on the day.
type Payout struct {
	id, affiliateID  string
	grossPesewas     int64
	withholdingBasis int64
	withheldPesewas  int64
	netPesewas       int64
	status           PayoutStatus
	// approverKey is the operator who approved it, keyed. A payout has a name
	// against it, and that name is a digest like every other actor in an
	// audit trail here.
	approverKey  string
	reference    string
	transferCode string
	requestedAt  time.Time
	decidedAt    *time.Time
}

// RequestPayout opens a request.
//
// withholdingBasisPoints is the rate in hundredths of a percent — 750 is 7.5%.
// Basis points rather than a float because a tax rate multiplied by money must
// not be a floating-point operation, and because "7.5%" written as 0.075 is a
// number no computer stores exactly.
//
// A rate of zero is refused rather than treated as "no withholding": the
// difference between a deliberate zero and an unset config is exactly the
// difference between a decision and a mistake, and this is a tax question.
func RequestPayout(
	id, affiliateID string, grossPesewas, minimumPesewas int64,
	withholdingBasisPoints int64, at time.Time,
) (Payout, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(id)) ||
		!opaquePattern.MatchString(strings.TrimSpace(affiliateID)) ||
		grossPesewas <= 0 || at.IsZero() {
		return Payout{}, ErrInvalidPayout
	}
	if withholdingBasisPoints <= 0 || withholdingBasisPoints >= 10_000 {
		return Payout{}, ErrWithholdingUnset
	}
	if grossPesewas < minimumPesewas {
		return Payout{}, ErrBelowMinimum
	}
	// Integer arithmetic, rounding the withholding down so a member of the
	// public is never over-withheld by a rounding decision nobody made.
	withheld := grossPesewas * withholdingBasisPoints / 10_000
	if withheld >= grossPesewas {
		return Payout{}, ErrInvalidPayout
	}
	return Payout{
		id: strings.TrimSpace(id), affiliateID: strings.TrimSpace(affiliateID),
		grossPesewas: grossPesewas, withholdingBasis: withholdingBasisPoints,
		withheldPesewas: withheld, netPesewas: grossPesewas - withheld,
		status: PayoutRequested, requestedAt: at.UTC(),
	}, nil
}

// Approve records an operator's decision to pay.
func (payout Payout) Approve(approverKey string, at time.Time) (Payout, error) {
	if payout.status != PayoutRequested {
		return Payout{}, ErrPayoutTransition
	}
	if !keyPattern.MatchString(approverKey) || at.IsZero() {
		return Payout{}, ErrInvalidPayout
	}
	next := payout
	next.status = PayoutApproved
	next.approverKey = approverKey
	decided := at.UTC()
	next.decidedAt = &decided
	return next, nil
}

// Refuse records an operator declining it.
func (payout Payout) Refuse(approverKey string, at time.Time) (Payout, error) {
	if payout.status != PayoutRequested {
		return Payout{}, ErrPayoutTransition
	}
	if !keyPattern.MatchString(approverKey) || at.IsZero() {
		return Payout{}, ErrInvalidPayout
	}
	next := payout
	next.status = PayoutRefused
	next.approverKey = approverKey
	decided := at.UTC()
	next.decidedAt = &decided
	return next, nil
}

// Sent records that the transfer was accepted by the processor.
func (payout Payout) Sent(reference, transferCode string) (Payout, error) {
	if payout.status != PayoutApproved {
		return Payout{}, ErrPayoutTransition
	}
	if strings.TrimSpace(reference) == "" || strings.TrimSpace(transferCode) == "" {
		return Payout{}, ErrInvalidPayout
	}
	next := payout
	next.status = PayoutSent
	next.reference = strings.TrimSpace(reference)
	next.transferCode = strings.TrimSpace(transferCode)
	return next, nil
}

// Failed records a transfer the processor would not accept. An approved
// payout that failed stays failed rather than returning to requested: whether
// to try again is a decision somebody makes, not a state machine's to assume.
func (payout Payout) Failed() (Payout, error) {
	if payout.status != PayoutApproved {
		return Payout{}, ErrPayoutTransition
	}
	next := payout
	next.status = PayoutFailed
	return next, nil
}

func (payout Payout) ID() string              { return payout.id }
func (payout Payout) AffiliateID() string     { return payout.affiliateID }
func (payout Payout) GrossPesewas() int64     { return payout.grossPesewas }
func (payout Payout) WithheldPesewas() int64  { return payout.withheldPesewas }
func (payout Payout) NetPesewas() int64       { return payout.netPesewas }
func (payout Payout) WithholdingBasis() int64 { return payout.withholdingBasis }
func (payout Payout) Status() PayoutStatus    { return payout.status }
func (payout Payout) ApproverKey() string     { return payout.approverKey }
func (payout Payout) Reference() string       { return payout.reference }
func (payout Payout) TransferCode() string    { return payout.transferCode }
func (payout Payout) RequestedAt() time.Time  { return payout.requestedAt }
func (payout Payout) DecidedAt() *time.Time   { return payout.decidedAt }

// RehydratePayout rebuilds a payout from storage.
func RehydratePayout(
	id, affiliateID string, gross, basis, withheld, net int64,
	status PayoutStatus, approverKey, reference, transferCode string,
	requestedAt time.Time, decidedAt *time.Time,
) (Payout, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(id)) || gross <= 0 ||
		withheld < 0 || net <= 0 || withheld+net != gross {
		return Payout{}, ErrInvalidPayout
	}
	return Payout{
		id: id, affiliateID: affiliateID, grossPesewas: gross, withholdingBasis: basis,
		withheldPesewas: withheld, netPesewas: net, status: status,
		approverKey: approverKey, reference: reference, transferCode: transferCode,
		requestedAt: requestedAt.UTC(), decidedAt: decidedAt,
	}, nil
}
