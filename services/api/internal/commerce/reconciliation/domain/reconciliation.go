package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const MaxMinor int64 = 1_000_000_000_000

var (
	ErrInvalid  = errors.New("invalid reconciliation fact")
	ErrMismatch = errors.New("ledger proof mismatch")
	opaque      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	key         = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type Currency string

const (
	CurrencyGHS Currency = "GHS"
	CurrencyUSD Currency = "USD"
)

type ProviderStatus string

const (
	StatusSettled ProviderStatus = "settled"
	StatusFailed  ProviderStatus = "failed"
)

type Outcome string

const (
	OutcomeReconciled Outcome = "reconciled"
	OutcomeException  Outcome = "exception"
	OutcomeNotDue     Outcome = "not_due"
)

type ExceptionCode string

const (
	ExceptionLedgerMissing    ExceptionCode = "ledger_missing"
	ExceptionReference        ExceptionCode = "reference_mismatch"
	ExceptionCurrency         ExceptionCode = "currency_mismatch"
	ExceptionAmount           ExceptionCode = "amount_mismatch"
	ExceptionUnbalancedLedger ExceptionCode = "ledger_unbalanced"
)

type StatementFact struct {
	id, providerKey, eventKey, referenceKey, ledgerCommand, fingerprint string
	currency                                                            Currency
	status                                                              ProviderStatus
	minor                                                               int64
	occurredAt, receivedAt                                              time.Time
}

type FactState struct {
	ID, ProviderKey, EventKey, ReferenceKey, LedgerCommand, Fingerprint string
	Currency                                                            Currency
	Status                                                              ProviderStatus
	Minor                                                               int64
	OccurredAt, ReceivedAt                                              time.Time
}

func NewFact(id, providerKey, eventKey, referenceKey, command string, currency Currency, status ProviderStatus, minor int64, occurredAt, receivedAt time.Time) (StatementFact, error) {
	f := StatementFact{id: id, providerKey: providerKey, eventKey: eventKey, referenceKey: referenceKey, ledgerCommand: command, currency: currency, status: status, minor: minor, occurredAt: occurredAt.UTC(), receivedAt: receivedAt.UTC()}
	if err := validateFact(f); err != nil {
		return StatementFact{}, err
	}
	f.fingerprint = factFingerprint(f)
	return f, nil
}

func RehydrateFact(s FactState) (StatementFact, error) {
	f := StatementFact{id: s.ID, providerKey: s.ProviderKey, eventKey: s.EventKey, referenceKey: s.ReferenceKey, ledgerCommand: s.LedgerCommand, fingerprint: s.Fingerprint, currency: s.Currency, status: s.Status, minor: s.Minor, occurredAt: s.OccurredAt.UTC(), receivedAt: s.ReceivedAt.UTC()}
	if err := validateFact(f); err != nil || f.fingerprint != factFingerprint(f) {
		return StatementFact{}, ErrInvalid
	}
	return f, nil
}

func validateFact(f StatementFact) error {
	if !opaque.MatchString(f.id) || !key.MatchString(f.providerKey) || !key.MatchString(f.eventKey) || !key.MatchString(f.referenceKey) || !opaque.MatchString(f.ledgerCommand) ||
		(f.currency != CurrencyGHS && f.currency != CurrencyUSD) || (f.status != StatusSettled && f.status != StatusFailed) ||
		f.minor <= 0 || f.minor > MaxMinor || f.occurredAt.IsZero() || f.receivedAt.IsZero() || f.occurredAt.After(f.receivedAt.Add(5*time.Minute)) {
		return ErrInvalid
	}
	return nil
}

func factFingerprint(f StatementFact) string {
	return digest(strings.Join([]string{f.providerKey, f.eventKey, f.referenceKey, f.ledgerCommand, string(f.currency), string(f.status), strconv.FormatInt(f.minor, 10), f.occurredAt.Format(time.RFC3339Nano)}, "\x00"))
}

type LedgerProof struct {
	CommandID, ReferenceKey string
	Currency                Currency
	Minor                   int64
	Balanced                bool
}

type Decision struct {
	outcome   Outcome
	exception ExceptionCode
}

func Compare(f StatementFact, p LedgerProof, found bool) Decision {
	if f.status == StatusFailed {
		return Decision{outcome: OutcomeNotDue}
	}
	if !found {
		return Decision{outcome: OutcomeException, exception: ExceptionLedgerMissing}
	}
	if p.CommandID != f.ledgerCommand || p.ReferenceKey != f.referenceKey {
		return Decision{outcome: OutcomeException, exception: ExceptionReference}
	}
	if p.Currency != f.currency {
		return Decision{outcome: OutcomeException, exception: ExceptionCurrency}
	}
	if p.Minor != f.minor {
		return Decision{outcome: OutcomeException, exception: ExceptionAmount}
	}
	if !p.Balanced {
		return Decision{outcome: OutcomeException, exception: ExceptionUnbalancedLedger}
	}
	return Decision{outcome: OutcomeReconciled}
}

type Audit struct {
	id, factID, fingerprint string
	outcome                 Outcome
	exception               ExceptionCode
	recordedAt              time.Time
}

type AuditState struct {
	ID, FactID, Fingerprint string
	Outcome                 Outcome
	Exception               ExceptionCode
	RecordedAt              time.Time
}

func NewAudit(id string, fact StatementFact, d Decision, at time.Time) (Audit, error) {
	a := Audit{id: id, factID: fact.id, outcome: d.outcome, exception: d.exception, recordedAt: at.UTC()}
	if !opaque.MatchString(id) || at.IsZero() || (d.outcome != OutcomeReconciled && d.outcome != OutcomeException && d.outcome != OutcomeNotDue) ||
		(d.outcome == OutcomeException) != (d.exception != "") {
		return Audit{}, ErrInvalid
	}
	a.fingerprint = digest(strings.Join([]string{a.factID, string(a.outcome), string(a.exception)}, "\x00"))
	return a, nil
}

func RehydrateAudit(s AuditState) (Audit, error) {
	a := Audit{id: s.ID, factID: s.FactID, fingerprint: s.Fingerprint, outcome: s.Outcome, exception: s.Exception, recordedAt: s.RecordedAt.UTC()}
	if !opaque.MatchString(a.id) || !opaque.MatchString(a.factID) || a.recordedAt.IsZero() || (a.outcome == OutcomeException) != (a.exception != "") ||
		a.fingerprint != digest(strings.Join([]string{a.factID, string(a.outcome), string(a.exception)}, "\x00")) {
		return Audit{}, ErrInvalid
	}
	return a, nil
}

type Checkpoint struct {
	id, day, fingerprint        string
	total, reconciled, excepted int
	completedAt                 time.Time
}

type CheckpointState struct {
	ID, Day, Fingerprint        string
	Total, Reconciled, Excepted int
	CompletedAt                 time.Time
}

func NewCheckpoint(id, day string, total, reconciled, excepted int, at time.Time) (Checkpoint, error) {
	c := Checkpoint{id: id, day: day, total: total, reconciled: reconciled, excepted: excepted, completedAt: at.UTC()}
	if !opaque.MatchString(id) || !validDay(day) || total < 0 || reconciled < 0 || excepted < 0 || reconciled+excepted > total || at.IsZero() {
		return Checkpoint{}, ErrInvalid
	}
	c.fingerprint = checkpointFingerprint(c)
	return c, nil
}

func RehydrateCheckpoint(s CheckpointState) (Checkpoint, error) {
	c := Checkpoint{id: s.ID, day: s.Day, fingerprint: s.Fingerprint, total: s.Total, reconciled: s.Reconciled, excepted: s.Excepted, completedAt: s.CompletedAt.UTC()}
	if !opaque.MatchString(c.id) || !validDay(c.day) || c.total < 0 || c.reconciled < 0 || c.excepted < 0 || c.reconciled+c.excepted > c.total || c.completedAt.IsZero() || c.fingerprint != checkpointFingerprint(c) {
		return Checkpoint{}, ErrInvalid
	}
	return c, nil
}

func checkpointFingerprint(c Checkpoint) string {
	return digest(strings.Join([]string{c.day, strconv.Itoa(c.total), strconv.Itoa(c.reconciled), strconv.Itoa(c.excepted)}, "\x00"))
}
func validDay(day string) bool {
	t, e := time.Parse("2006-01-02", day)
	return e == nil && t.Format("2006-01-02") == day
}
func digest(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }

func (f StatementFact) ID() string             { return f.id }
func (f StatementFact) ProviderKey() string    { return f.providerKey }
func (f StatementFact) EventKey() string       { return f.eventKey }
func (f StatementFact) ReferenceKey() string   { return f.referenceKey }
func (f StatementFact) LedgerCommand() string  { return f.ledgerCommand }
func (f StatementFact) Currency() Currency     { return f.currency }
func (f StatementFact) Status() ProviderStatus { return f.status }
func (f StatementFact) Minor() int64           { return f.minor }
func (f StatementFact) OccurredAt() time.Time  { return f.occurredAt }
func (f StatementFact) ReceivedAt() time.Time  { return f.receivedAt }
func (f StatementFact) Fingerprint() string    { return f.fingerprint }
func (d Decision) Outcome() Outcome            { return d.outcome }
func (d Decision) Exception() ExceptionCode    { return d.exception }
func (a Audit) ID() string                     { return a.id }
func (a Audit) FactID() string                 { return a.factID }
func (a Audit) Outcome() Outcome               { return a.outcome }
func (a Audit) Exception() ExceptionCode       { return a.exception }
func (a Audit) Fingerprint() string            { return a.fingerprint }
func (a Audit) RecordedAt() time.Time          { return a.recordedAt }
func (c Checkpoint) ID() string                { return c.id }
func (c Checkpoint) Day() string               { return c.day }
func (c Checkpoint) Total() int                { return c.total }
func (c Checkpoint) Reconciled() int           { return c.reconciled }
func (c Checkpoint) Excepted() int             { return c.excepted }
func (c Checkpoint) Fingerprint() string       { return c.fingerprint }
func (c Checkpoint) CompletedAt() time.Time    { return c.completedAt }
