package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidUnits    = errors.New("allowance units must be positive")
	ErrInsufficient    = errors.New("insufficient weekly seed allowance")
	ErrCommandConflict = errors.New("command id was already used with different input")
	ErrInvalidLedger   = errors.New("invalid allowance ledger")
)

type EntryKind string

const (
	EntryIssuance EntryKind = "issuance"
	EntrySpend    EntryKind = "spend"
	EntryRenewal  EntryKind = "renewal"
)

// Entry is immutable audit evidence. Units is always a positive quantity;
// BalanceAfter records the server-authoritative result.
type Entry struct {
	Sequence     int64
	Kind         EntryKind
	CommandID    string
	Fingerprint  string
	Units        int64
	BalanceAfter int64
	WeekStart    time.Time
	OccurredAt   time.Time
	ActorKey     string
}

type Ledger struct {
	subjectKey      string
	weeklyAllowance int64
	balance         int64
	weekStart       time.Time
	version         int64
	entries         []Entry
}

func Issue(subjectKey string, weeklyAllowance int64, weekStart, now time.Time, commandID, fingerprint string) (Ledger, error) {
	if subjectKey == "" || commandID == "" || fingerprint == "" || weeklyAllowance <= 0 {
		return Ledger{}, ErrInvalidLedger
	}
	l := Ledger{subjectKey: subjectKey, weeklyAllowance: weeklyAllowance, weekStart: weekStart.UTC()}
	l.append(EntryIssuance, commandID, fingerprint, weeklyAllowance, weeklyAllowance, now, "system")
	return l, nil
}

func Rehydrate(subjectKey string, weeklyAllowance, balance int64, weekStart time.Time, version int64, entries []Entry) (Ledger, error) {
	if subjectKey == "" || weeklyAllowance <= 0 || balance < 0 || version < 1 || int64(len(entries)) != version {
		return Ledger{}, ErrInvalidLedger
	}
	copyEntries := append([]Entry(nil), entries...)
	for i, entry := range copyEntries {
		if entry.Sequence != int64(i+1) || entry.Units <= 0 || entry.BalanceAfter < 0 {
			return Ledger{}, ErrInvalidLedger
		}
	}
	if len(copyEntries) == 0 || copyEntries[len(copyEntries)-1].BalanceAfter != balance {
		return Ledger{}, ErrInvalidLedger
	}
	return Ledger{subjectKey: subjectKey, weeklyAllowance: weeklyAllowance, balance: balance, weekStart: weekStart.UTC(), version: version, entries: copyEntries}, nil
}

func (l Ledger) Spend(units int64, weekStart, now time.Time, commandID, fingerprint string) (Ledger, bool, error) {
	if units <= 0 {
		return l, false, ErrInvalidUnits
	}
	if replay, conflict := l.command(commandID, fingerprint); replay || conflict {
		if conflict {
			return l, false, ErrCommandConflict
		}
		return l, false, nil
	}
	next := l.clone()
	if weekStart.UTC().After(next.weekStart) {
		next.weekStart = weekStart.UTC()
		next.append(EntryRenewal, "renew:"+weekStart.UTC().Format(time.RFC3339), "weekly-renewal:"+weekStart.UTC().Format(time.RFC3339), next.weeklyAllowance, next.weeklyAllowance, now, "system")
	}
	if units > next.balance {
		return l, false, ErrInsufficient
	}
	next.append(EntrySpend, commandID, fingerprint, units, next.balance-units, now, next.subjectKey)
	return next, true, nil
}

func (l Ledger) Renew(weekStart, now time.Time, commandID, fingerprint string) (Ledger, bool, error) {
	if replay, conflict := l.command(commandID, fingerprint); replay || conflict {
		if conflict {
			return l, false, ErrCommandConflict
		}
		return l, false, nil
	}
	if !weekStart.UTC().After(l.weekStart) {
		return l, false, nil
	}
	next := l.clone()
	next.weekStart = weekStart.UTC()
	next.append(EntryRenewal, commandID, fingerprint, next.weeklyAllowance, next.weeklyAllowance, now, "system")
	return next, true, nil
}

func (l *Ledger) append(kind EntryKind, commandID, fingerprint string, units, balance int64, now time.Time, actor string) {
	l.version++
	l.balance = balance
	l.entries = append(l.entries, Entry{
		Sequence: l.version, Kind: kind, CommandID: commandID, Fingerprint: fingerprint,
		Units: units, BalanceAfter: balance, WeekStart: l.weekStart, OccurredAt: now.UTC(), ActorKey: actor,
	})
}

func (l Ledger) command(id, fingerprint string) (bool, bool) {
	if id == "" || fingerprint == "" {
		return false, true
	}
	for _, e := range l.entries {
		if e.CommandID == id {
			return e.Fingerprint == fingerprint, e.Fingerprint != fingerprint
		}
	}
	return false, false
}

func (l Ledger) clone() Ledger          { l.entries = append([]Entry(nil), l.entries...); return l }
func (l Ledger) SubjectKey() string     { return l.subjectKey }
func (l Ledger) WeeklyAllowance() int64 { return l.weeklyAllowance }
func (l Ledger) Balance() int64         { return l.balance }
func (l Ledger) WeekStart() time.Time   { return l.weekStart }
func (l Ledger) Version() int64         { return l.version }
func (l Ledger) Entries() []Entry       { return append([]Entry(nil), l.entries...) }
func (l Ledger) EntriesAfter(version int64) ([]Entry, error) {
	if version < 0 || version > int64(len(l.entries)) {
		return nil, fmt.Errorf("%w: invalid version", ErrInvalidLedger)
	}
	return append([]Entry(nil), l.entries[version:]...), nil
}
