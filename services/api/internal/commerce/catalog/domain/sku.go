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

const MaxPriceMinor int64 = 100_000_000

var (
	ErrInvalid    = errors.New("invalid catalog sku")
	ErrStale      = errors.New("stale catalog command")
	ErrMismatch   = errors.New("catalog command mismatch")
	ErrTransition = errors.New("invalid catalog transition")
)
var slug = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Kind string

const (
	KindPhysicalGood   Kind = "physical_good"
	KindEventTicket    Kind = "event_ticket"
	KindDigitalService Kind = "digital_service"
)

type Currency string

const (
	CurrencyGHS Currency = "GHS"
	CurrencyUSD Currency = "USD"
)

type Price struct {
	Currency Currency `bson:"currency"`
	Minor    int64    `bson:"minor"`
}

func NewPrice(currency Currency, minor int64) (Price, error) {
	p := Price{currency, minor}
	if !validPrice(p) {
		return Price{}, ErrInvalid
	}
	return p, nil
}
func validPrice(p Price) bool {
	return (p.Currency == CurrencyGHS || p.Currency == CurrencyUSD) && p.Minor >= 1 && p.Minor <= MaxPriceMinor
}

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusRetired   Status = "retired"
)

type Command struct {
	ID               string    `bson:"id"`
	ExpectedRevision uint64    `bson:"expectedRevision"`
	At               time.Time `bson:"at"`
}
type Event struct {
	Sequence  uint64    `bson:"sequence"`
	CommandID string    `bson:"commandId"`
	Action    string    `bson:"action"`
	At        time.Time `bson:"at"`
}
type Applied struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type SKU struct {
	id, skuKey, titleRef   string
	version                uint64
	kind                   Kind
	price                  Price
	status                 Status
	publishedAt, retiredAt time.Time
	revision               uint64
	events                 []Event
	commands               []Applied
}
type State struct {
	ID, SKUKey, TitleRef   string
	Version                uint64
	Kind                   Kind
	Price                  Price
	Status                 Status
	PublishedAt, RetiredAt time.Time
	Revision               uint64
	Events                 []Event
	Commands               []Applied
}

func Create(id, sku, title string, version uint64, kind Kind, price Price, c Command) (SKU, error) {
	s := SKU{id: id, skuKey: sku, titleRef: title, version: version, kind: kind, price: price, status: StatusDraft}
	if !opaque.MatchString(id) || !slug.MatchString(sku) || !key.MatchString(title) || version == 0 || !validKind(kind) || !validPrice(price) || c.ExpectedRevision != 0 {
		return SKU{}, ErrInvalid
	}
	if e := s.apply(c, "create"); e != nil {
		return SKU{}, e
	}
	return s, nil
}
func NextVersion(prior SKU, id string, price Price, c Command) (SKU, error) {
	if prior.status != StatusPublished && prior.status != StatusRetired {
		return SKU{}, ErrTransition
	}
	return Create(id, prior.skuKey, prior.titleRef, prior.version+1, prior.kind, price, c)
}
func Rehydrate(x State) (SKU, error) {
	s := SKU{id: x.ID, skuKey: x.SKUKey, titleRef: x.TitleRef, version: x.Version, kind: x.Kind, price: x.Price, status: x.Status, publishedAt: x.PublishedAt, retiredAt: x.RetiredAt, revision: x.Revision, events: append([]Event(nil), x.Events...), commands: append([]Applied(nil), x.Commands...)}
	if !opaque.MatchString(s.id) || !slug.MatchString(s.skuKey) || !key.MatchString(s.titleRef) || s.version == 0 || !validKind(s.kind) || !validPrice(s.price) || !validState(s) || s.revision == 0 || len(s.events) != int(s.revision) || len(s.commands) != int(s.revision) {
		return SKU{}, ErrInvalid
	}
	return s, nil
}
func validKind(k Kind) bool {
	return k == KindPhysicalGood || k == KindEventTicket || k == KindDigitalService
}
func validState(s SKU) bool {
	switch s.status {
	case StatusDraft:
		return s.publishedAt.IsZero() && s.retiredAt.IsZero()
	case StatusPublished:
		return !s.publishedAt.IsZero() && s.retiredAt.IsZero()
	case StatusRetired:
		return !s.publishedAt.IsZero() && !s.retiredAt.IsZero() && !s.retiredAt.Before(s.publishedAt)
	default:
		return false
	}
}
func (s SKU) Publish(c Command) (SKU, error) {
	if replay, e := s.replay(c, "publish"); replay || e != nil {
		return s, e
	}
	if s.status != StatusDraft {
		return SKU{}, ErrTransition
	}
	s = s.clone()
	s.status = StatusPublished
	s.publishedAt = c.At.UTC()
	return s, s.apply(c, "publish")
}
func (s SKU) Retire(c Command) (SKU, error) {
	if replay, e := s.replay(c, "retire"); replay || e != nil {
		return s, e
	}
	if s.status != StatusPublished || c.At.Before(s.publishedAt) {
		return SKU{}, ErrTransition
	}
	s = s.clone()
	s.status = StatusRetired
	s.retiredAt = c.At.UTC()
	return s, s.apply(c, "retire")
}
func (s SKU) clone() SKU {
	s.events = append([]Event(nil), s.events...)
	s.commands = append([]Applied(nil), s.commands...)
	return s
}
func (s *SKU) apply(c Command, a string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != s.revision {
		return ErrStale
	}
	f := fingerprint(s.id, c, a)
	s.revision++
	s.events = append(s.events, Event{s.revision, c.ID, a, c.At.UTC()})
	s.commands = append(s.commands, Applied{c.ID, f, s.revision})
	return nil
}
func (s SKU) replay(c Command, a string) (bool, error) {
	f := fingerprint(s.id, c, a)
	for _, x := range s.commands {
		if x.ID == c.ID {
			if x.Fingerprint != f {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func fingerprint(id string, c Command, a string) string {
	x := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(x[:])
}
func (s SKU) ID() string             { return s.id }
func (s SKU) SKUKey() string         { return s.skuKey }
func (s SKU) TitleRef() string       { return s.titleRef }
func (s SKU) Version() uint64        { return s.version }
func (s SKU) Kind() Kind             { return s.kind }
func (s SKU) Price() Price           { return s.price }
func (s SKU) Status() Status         { return s.status }
func (s SKU) PublishedAt() time.Time { return s.publishedAt }
func (s SKU) RetiredAt() time.Time   { return s.retiredAt }
func (s SKU) Revision() uint64       { return s.revision }
func (s SKU) Events() []Event        { return append([]Event(nil), s.events...) }
func (s SKU) Commands() []Applied    { return append([]Applied(nil), s.commands...) }

var _ = strings.Compare
