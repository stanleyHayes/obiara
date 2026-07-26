package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalid    = errors.New("invalid matching feature")
	ErrStale      = errors.New("stale matching feature grant")
	ErrMismatch   = errors.New("matching feature command mismatch")
	ErrTransition = errors.New("invalid matching feature transition")
)

var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Definition struct {
	Key         string    `bson:"key"`
	Version     uint64    `bson:"version"`
	Purpose     string    `bson:"purpose"`
	Optional    bool      `bson:"optional"`
	EffectiveAt time.Time `bson:"effectiveAt"`
}

func NewDefinition(key string, version uint64, purpose string, effectiveAt time.Time) (Definition, error) {
	d := Definition{Key: strings.TrimSpace(key), Version: version, Purpose: strings.TrimSpace(purpose), Optional: true, EffectiveAt: effectiveAt.UTC()}
	if !token.MatchString(d.Key) || !token.MatchString(d.Purpose) || version == 0 || effectiveAt.IsZero() {
		return Definition{}, ErrInvalid
	}
	return d, nil
}
func (d Definition) Active(at time.Time) bool {
	return d.Optional && !at.IsZero() && !at.UTC().Before(d.EffectiveAt)
}

type Status string

const (
	StatusGranted   Status = "granted"
	StatusWithdrawn Status = "withdrawn"
)

type Command struct {
	ID               string    `bson:"id"`
	ExpectedRevision uint64    `bson:"expectedRevision"`
	At               time.Time `bson:"at"`
}
type Event struct {
	Sequence uint64    `bson:"sequence"`
	Command  string    `bson:"command"`
	Action   string    `bson:"action"`
	At       time.Time `bson:"at"`
}
type Applied struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type Grant struct {
	memberKey, featureKey, purpose string
	featureVersion, grantVersion   uint64
	status                         Status
	grantedAt, withdrawnAt         time.Time
	revision                       uint64
	events                         []Event
	commands                       []Applied
}
type GrantState struct {
	MemberKey      string
	FeatureKey     string
	Purpose        string
	FeatureVersion uint64
	GrantVersion   uint64
	Status         Status
	GrantedAt      time.Time
	WithdrawnAt    time.Time
	Revision       uint64
	Events         []Event
	Commands       []Applied
}

func GrantFeature(memberKey string, d Definition, grantVersion uint64, c Command) (Grant, error) {
	g := Grant{memberKey: memberKey, featureKey: d.Key, purpose: d.Purpose, featureVersion: d.Version, grantVersion: grantVersion, status: StatusGranted, grantedAt: c.At.UTC()}
	if !opaque.MatchString(memberKey) || !d.Active(c.At) || grantVersion == 0 || c.ExpectedRevision != 0 {
		return Grant{}, ErrInvalid
	}
	if err := g.apply(c, "grant"); err != nil {
		return Grant{}, err
	}
	return g, nil
}
func RehydrateGrant(s GrantState) (Grant, error) {
	g := Grant{memberKey: s.MemberKey, featureKey: s.FeatureKey, purpose: s.Purpose, featureVersion: s.FeatureVersion, grantVersion: s.GrantVersion, status: s.Status, grantedAt: s.GrantedAt, withdrawnAt: s.WithdrawnAt, revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]Applied(nil), s.Commands...)}
	if !opaque.MatchString(g.memberKey) || !token.MatchString(g.featureKey) || !token.MatchString(g.purpose) || g.featureVersion == 0 || g.grantVersion == 0 || g.grantedAt.IsZero() || g.revision == 0 || len(g.events) != int(g.revision) || len(g.commands) != int(g.revision) || (g.status != StatusGranted && g.status != StatusWithdrawn) {
		return Grant{}, ErrInvalid
	}
	return g, nil
}
func (g Grant) Withdraw(c Command) (Grant, error) {
	if replay, err := g.replay(c, "withdraw"); replay || err != nil {
		return g, err
	}
	if g.status != StatusGranted || c.At.Before(g.grantedAt) {
		return Grant{}, ErrTransition
	}
	g.events, g.commands = append([]Event(nil), g.events...), append([]Applied(nil), g.commands...)
	g.status, g.withdrawnAt = StatusWithdrawn, c.At.UTC()
	return g, g.apply(c, "withdraw")
}
func (g Grant) Regrant(d Definition, grantVersion uint64, c Command) (Grant, error) {
	action := "regrant:" + d.Purpose + ":" + strconv.FormatUint(d.Version, 10) + ":" + strconv.FormatUint(grantVersion, 10)
	if replay, err := g.replay(c, action); replay || err != nil {
		return g, err
	}
	if g.status != StatusWithdrawn || !d.Active(c.At) || d.Key != g.featureKey || grantVersion <= g.grantVersion || c.At.Before(g.withdrawnAt) {
		return Grant{}, ErrTransition
	}
	g.events, g.commands = append([]Event(nil), g.events...), append([]Applied(nil), g.commands...)
	g.purpose, g.featureVersion, g.grantVersion = d.Purpose, d.Version, grantVersion
	g.status, g.grantedAt, g.withdrawnAt = StatusGranted, c.At.UTC(), time.Time{}
	return g, g.apply(c, action)
}
func (g Grant) Effective(d Definition, at time.Time) bool {
	return g.status == StatusGranted && d.Active(at) && g.featureKey == d.Key && g.purpose == d.Purpose && g.featureVersion == d.Version && !at.UTC().Before(g.grantedAt)
}
func (g *Grant) apply(c Command, action string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != g.revision {
		return ErrStale
	}
	fp := fingerprint(g.memberKey, g.featureKey, c, action)
	g.revision++
	g.events = append(g.events, Event{Sequence: g.revision, Command: c.ID, Action: action, At: c.At.UTC()})
	g.commands = append(g.commands, Applied{ID: c.ID, Fingerprint: fp, Revision: g.revision})
	return nil
}
func (g Grant) replay(c Command, action string) (bool, error) {
	fp := fingerprint(g.memberKey, g.featureKey, c, action)
	for _, a := range g.commands {
		if a.ID == c.ID {
			if a.Fingerprint != fp {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func fingerprint(member, feature string, c Command, action string) string {
	sum := sha256.Sum256([]byte(member + "\x00" + feature + "\x00" + c.ID + "\x00" + action + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(sum[:])
}

type ConsentRef struct {
	MemberKey    string `bson:"memberKey"`
	GrantVersion uint64 `bson:"grantVersion"`
}
type EnabledFeature struct {
	Key            string       `bson:"key"`
	FeatureVersion uint64       `bson:"featureVersion"`
	Purpose        string       `bson:"purpose"`
	Consents       []ConsentRef `bson:"consents"`
}
type Decision struct {
	ID          string           `bson:"id"`
	Pair        []string         `bson:"pair"`
	Features    []EnabledFeature `bson:"features"`
	EvaluatedAt time.Time        `bson:"evaluatedAt"`
}

func NewDecision(id, first, second string, features []EnabledFeature, at time.Time) (Decision, error) {
	pair := []string{first, second}
	slices.Sort(pair)
	f := append([]EnabledFeature(nil), features...)
	slices.SortFunc(f, func(a, b EnabledFeature) int { return strings.Compare(a.Key, b.Key) })
	d := Decision{ID: id, Pair: pair, Features: f, EvaluatedAt: at.UTC()}
	if !opaque.MatchString(id) || first == second || !opaque.MatchString(first) || !opaque.MatchString(second) || at.IsZero() || !validFeatures(f, pair) {
		return Decision{}, ErrInvalid
	}
	return d, nil
}
func validFeatures(fs []EnabledFeature, pair []string) bool {
	for i, f := range fs {
		if !token.MatchString(f.Key) || !token.MatchString(f.Purpose) || f.FeatureVersion == 0 || len(f.Consents) != 2 || (i > 0 && fs[i-1].Key == f.Key) {
			return false
		}
		c := append([]ConsentRef(nil), f.Consents...)
		slices.SortFunc(c, func(a, b ConsentRef) int { return strings.Compare(a.MemberKey, b.MemberKey) })
		if c[0].MemberKey != pair[0] || c[1].MemberKey != pair[1] || c[0].GrantVersion == 0 || c[1].GrantVersion == 0 {
			return false
		}
	}
	return true
}

func (g Grant) MemberKey() string      { return g.memberKey }
func (g Grant) FeatureKey() string     { return g.featureKey }
func (g Grant) Purpose() string        { return g.purpose }
func (g Grant) FeatureVersion() uint64 { return g.featureVersion }
func (g Grant) GrantVersion() uint64   { return g.grantVersion }
func (g Grant) Status() Status         { return g.status }
func (g Grant) GrantedAt() time.Time   { return g.grantedAt }
func (g Grant) WithdrawnAt() time.Time { return g.withdrawnAt }
func (g Grant) Revision() uint64       { return g.revision }
func (g Grant) Events() []Event        { return append([]Event(nil), g.events...) }
func (g Grant) Commands() []Applied    { return append([]Applied(nil), g.commands...) }
