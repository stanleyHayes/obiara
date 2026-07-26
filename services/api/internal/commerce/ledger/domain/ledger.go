package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	MaxLines       = 32
	MaxMinor int64 = 1_000_000_000_000
)

var (
	ErrInvalid    = errors.New("invalid ledger posting")
	ErrUnbalanced = errors.New("unbalanced ledger posting")
	ErrOverflow   = errors.New("ledger amount overflow")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Currency string

const (
	CurrencyGHS Currency = "GHS"
	CurrencyUSD Currency = "USD"
)

type AccountClass string

const (
	ClassAsset     AccountClass = "asset"
	ClassLiability AccountClass = "liability"
	ClassEquity    AccountClass = "equity"
	ClassRevenue   AccountClass = "revenue"
	ClassExpense   AccountClass = "expense"
)

type Side string

const (
	SideDebit  Side = "debit"
	SideCredit Side = "credit"
)

type Purpose string

const (
	PurposeSaleSettlement    Purpose = "sale_settlement"
	PurposeRefundSettlement  Purpose = "refund_settlement"
	PurposeCatalogReceivable Purpose = "catalog_receivable"
)

type Line struct {
	AccountKey string       `bson:"accountKey"`
	Class      AccountClass `bson:"class"`
	Side       Side         `bson:"side"`
	Minor      int64        `bson:"minor"`
}
type Posting struct {
	id, commandID, fingerprint, referenceKey string
	purpose                                  Purpose
	currency                                 Currency
	lines                                    []Line
	postedAt                                 time.Time
}
type State struct {
	ID, CommandID, Fingerprint, ReferenceKey string
	Purpose                                  Purpose
	Currency                                 Currency
	Lines                                    []Line
	PostedAt                                 time.Time
}

func NewPosting(id, command, reference string, purpose Purpose, currency Currency, lines []Line, at time.Time) (Posting, error) {
	p := Posting{id: id, commandID: command, referenceKey: reference, purpose: purpose, currency: currency, lines: normalized(lines), postedAt: at.UTC()}
	if !opaque.MatchString(id) || !opaque.MatchString(command) || !key.MatchString(reference) || !validPurpose(purpose) || !validCurrency(currency) || at.IsZero() {
		return Posting{}, ErrInvalid
	}
	if e := validateLines(p.lines); e != nil {
		return Posting{}, e
	}
	p.fingerprint = fingerprint(p)
	return p, nil
}
func Rehydrate(s State) (Posting, error) {
	p := Posting{id: s.ID, commandID: s.CommandID, fingerprint: s.Fingerprint, referenceKey: s.ReferenceKey, purpose: s.Purpose, currency: s.Currency, lines: normalized(s.Lines), postedAt: s.PostedAt}
	if !opaque.MatchString(p.id) || !opaque.MatchString(p.commandID) || !key.MatchString(p.referenceKey) || !validPurpose(p.purpose) || !validCurrency(p.currency) || p.postedAt.IsZero() {
		return Posting{}, ErrInvalid
	}
	if e := validateLines(p.lines); e != nil {
		return Posting{}, e
	}
	if p.fingerprint != fingerprint(p) {
		return Posting{}, ErrInvalid
	}
	return p, nil
}
func validPurpose(p Purpose) bool {
	return p == PurposeSaleSettlement || p == PurposeRefundSettlement || p == PurposeCatalogReceivable
}
func validCurrency(c Currency) bool { return c == CurrencyGHS || c == CurrencyUSD }
func validClass(c AccountClass) bool {
	return c == ClassAsset || c == ClassLiability || c == ClassEquity || c == ClassRevenue || c == ClassExpense
}
func normalized(v []Line) []Line {
	x := append([]Line(nil), v...)
	slices.SortFunc(x, func(a, b Line) int { return strings.Compare(a.AccountKey, b.AccountKey) })
	return x
}
func validateLines(v []Line) error {
	if len(v) < 2 || len(v) > MaxLines {
		return ErrInvalid
	}
	var debit, credit int64
	for i, x := range v {
		if !key.MatchString(x.AccountKey) || !validClass(x.Class) || (x.Side != SideDebit && x.Side != SideCredit) || x.Minor <= 0 || x.Minor > MaxMinor || (i > 0 && v[i-1].AccountKey == x.AccountKey) {
			return ErrInvalid
		}
		var e error
		if x.Side == SideDebit {
			debit, e = add(debit, x.Minor)
		} else {
			credit, e = add(credit, x.Minor)
		}
		if e != nil {
			return e
		}
	}
	if debit != credit {
		return ErrUnbalanced
	}
	return nil
}
func add(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, ErrOverflow
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, ErrOverflow
	}
	return a + b, nil
}

type BookedLine struct {
	PostingID string
	Currency  Currency
	Line      Line
}

func RecomputeBalance(account string, class AccountClass, currency Currency, lines []BookedLine) (int64, error) {
	if !key.MatchString(account) || !validClass(class) || !validCurrency(currency) {
		return 0, ErrInvalid
	}
	var balance int64
	for _, b := range lines {
		if !opaque.MatchString(b.PostingID) || b.Currency != currency || b.Line.AccountKey != account || b.Line.Class != class || b.Line.Minor <= 0 || b.Line.Minor > MaxMinor {
			return 0, ErrInvalid
		}
		amount := b.Line.Minor
		if normalDebit(class) != (b.Line.Side == SideDebit) {
			amount = -amount
		}
		var e error
		balance, e = add(balance, amount)
		if e != nil {
			return 0, e
		}
	}
	return balance, nil
}
func normalDebit(c AccountClass) bool { return c == ClassAsset || c == ClassExpense }
func fingerprint(p Posting) string {
	var b strings.Builder
	b.WriteString(p.commandID + "\x00" + p.referenceKey + "\x00" + string(p.purpose) + "\x00" + string(p.currency))
	for _, x := range p.lines {
		b.WriteString("\x00" + x.AccountKey + "\x00" + string(x.Class) + "\x00" + string(x.Side) + "\x00" + strconv.FormatInt(x.Minor, 10))
	}
	s := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(s[:])
}
func (p Posting) ID() string           { return p.id }
func (p Posting) CommandID() string    { return p.commandID }
func (p Posting) Fingerprint() string  { return p.fingerprint }
func (p Posting) ReferenceKey() string { return p.referenceKey }
func (p Posting) Purpose() Purpose     { return p.purpose }
func (p Posting) Currency() Currency   { return p.currency }
func (p Posting) Lines() []Line        { return append([]Line(nil), p.lines...) }
func (p Posting) PostedAt() time.Time  { return p.postedAt }
