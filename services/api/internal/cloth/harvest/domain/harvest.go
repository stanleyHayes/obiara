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

type Status string

const (
	StatusAwaiting  Status = "awaiting_pair"
	StatusReady     Status = "ready"
	StatusHandedOff Status = "handed_off"
	StatusAccepted  Status = "accepted"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
	StatusDeclined  Status = "declined"
	StatusExpired   Status = "expired"
)

const ReadyValidity = 7 * 24 * time.Hour

var (
	ErrInvalid         = errors.New("invalid harvest")
	ErrNotMember       = errors.New("harvest actor is not a member")
	ErrConsent         = errors.New("harvest consent is not current")
	ErrTransition      = errors.New("invalid harvest transition")
	ErrExpired         = errors.New("harvest expired")
	ErrStaleRevision   = errors.New("stale harvest revision")
	ErrCommandMismatch = errors.New("harvest command replay mismatch")
)

var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var digest = regexp.MustCompile(`^[a-f0-9]{64}$`)
var reason = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)

var formats = map[string]bool{"woven_band": true, "framed_cloth": true, "digital_archive": true}
var tokens = map[string]bool{
	"warp_even": true, "warp_ripple": true, "weft_close": true, "weft_open": true,
	"edge_soft": true, "edge_bound": true, "tone_warm": true, "tone_cool": true,
	"mark_sparse": true, "mark_dense": true, "finish_matte": true, "finish_lustre": true,
}

type Payload struct {
	RecipeKey        string   `bson:"recipeKey" json:"recipeKey"`
	RecipeVersion    string   `bson:"recipeVersion" json:"recipeVersion"`
	RenderSeed       string   `bson:"renderSeed" json:"renderSeed"`
	ProductionTokens []string `bson:"productionTokens" json:"productionTokens"`
	Format           string   `bson:"format" json:"format"`
	DeliveryRef      string   `bson:"deliveryRef" json:"deliveryRef"`
	PolicyVersion    string   `bson:"policyVersion" json:"policyVersion"`
}

type Envelope struct {
	HandoffID        string    `json:"handoffId"`
	RecipeKey        string    `json:"recipeKey"`
	RecipeVersion    string    `json:"recipeVersion"`
	RenderSeed       string    `json:"renderSeed"`
	ProductionTokens []string  `json:"productionTokens"`
	Format           string    `json:"format"`
	DeliveryRef      string    `json:"deliveryRef"`
	PolicyVersion    string    `json:"policyVersion"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

type Command struct {
	ID, ActorKey     string
	ExpectedRevision uint64
	At               time.Time
}

type Event struct {
	Sequence                  uint64
	CommandID, Action, Reason string
	At                        time.Time
}

type Applied struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type Harvest struct {
	id, handoffID      string
	members            []string
	payload            Payload
	approvals          []string
	status             Status
	readyAt, expiresAt time.Time
	revision           uint64
	events             []Event
	commands           []Applied
}

type State struct {
	ID, HandoffID      string
	Members            []string
	Payload            Payload
	Approvals          []string
	Status             Status
	ReadyAt, ExpiresAt time.Time
	Revision           uint64
	Events             []Event
	Commands           []Applied
}

func Create(id string, members []string, payload Payload, command Command) (Harvest, error) {
	members = sorted(members)
	h := Harvest{id: id, members: members, payload: clonePayload(payload), status: StatusAwaiting}
	if !opaque.MatchString(id) || !validPair(members) || !validPayload(payload) || command.ExpectedRevision != 0 {
		return Harvest{}, ErrInvalid
	}
	if err := h.apply(command, "create", "", payloadHash(payload)); err != nil {
		return Harvest{}, err
	}
	return h, nil
}

func Rehydrate(state State) (Harvest, error) {
	h := Harvest{id: state.ID, handoffID: state.HandoffID, members: append([]string(nil), state.Members...),
		payload: clonePayload(state.Payload), approvals: append([]string(nil), state.Approvals...), status: state.Status,
		readyAt: state.ReadyAt.UTC(), expiresAt: state.ExpiresAt.UTC(), revision: state.Revision,
		events: append([]Event(nil), state.Events...), commands: append([]Applied(nil), state.Commands...)}
	if !opaque.MatchString(h.id) || !validPair(h.members) || !slices.IsSorted(h.members) || !validPayload(h.payload) ||
		len(h.events) != int(h.revision) || len(h.commands) != int(h.revision) || h.revision == 0 {
		return Harvest{}, ErrInvalid
	}
	if (h.status == StatusReady || h.status == StatusHandedOff || h.status == StatusAccepted) &&
		(len(h.approvals) != 2 || h.readyAt.IsZero() || h.expiresAt.IsZero()) {
		return Harvest{}, ErrInvalid
	}
	if h.handoffID != "" && !opaque.MatchString(h.handoffID) {
		return Harvest{}, ErrInvalid
	}
	return h, nil
}

func (h Harvest) Revise(payload Payload, command Command) (Harvest, error) {
	if replay, err := h.replay(command, "revise", "", payloadHash(payload)); replay || err != nil {
		return h, err
	}
	if h.status != StatusAwaiting && h.status != StatusReady || !validPayload(payload) {
		return Harvest{}, ErrTransition
	}
	h.payload, h.approvals, h.status = clonePayload(payload), nil, StatusAwaiting
	h.readyAt, h.expiresAt = time.Time{}, time.Time{}
	return h, h.apply(command, "revise", "", payloadHash(payload))
}

func (h Harvest) Approve(command Command) (Harvest, error) {
	if replay, err := h.replay(command, "approve", "", payloadHash(h.payload)); replay || err != nil {
		return h, err
	}
	if h.status != StatusAwaiting || !slices.Contains(h.members, command.ActorKey) {
		return Harvest{}, ErrNotMember
	}
	if slices.Contains(h.approvals, command.ActorKey) {
		return Harvest{}, ErrConsent
	}
	h.approvals = append(append([]string(nil), h.approvals...), command.ActorKey)
	slices.Sort(h.approvals)
	if len(h.approvals) == 2 {
		h.status, h.readyAt, h.expiresAt = StatusReady, command.At.UTC(), command.At.UTC().Add(ReadyValidity)
	}
	return h, h.apply(command, "approve", "", payloadHash(h.payload))
}

func (h Harvest) Handoff(handoffID string, command Command) (Harvest, error) {
	if replay, err := h.replay(command, "handoff", handoffID, payloadHash(h.payload)); replay || err != nil {
		return h, err
	}
	if h.status != StatusReady || len(h.approvals) != 2 {
		return Harvest{}, ErrConsent
	}
	if !command.At.Before(h.expiresAt) {
		return Harvest{}, ErrExpired
	}
	if !opaque.MatchString(handoffID) {
		return Harvest{}, ErrInvalid
	}
	h.status, h.handoffID = StatusHandedOff, handoffID
	return h, h.apply(command, "handoff", "", handoffID, payloadHash(h.payload))
}

func (h Harvest) Callback(next Status, code string, command Command) (Harvest, error) {
	action := "callback:" + string(next)
	if replay, err := h.replay(command, action, code); replay || err != nil {
		return h, err
	}
	if !reason.MatchString(code) {
		return Harvest{}, ErrInvalid
	}
	valid := next == StatusAccepted && h.status == StatusHandedOff ||
		next == StatusDeclined && h.status == StatusHandedOff ||
		next == StatusCompleted && h.status == StatusAccepted
	if !valid {
		return Harvest{}, ErrTransition
	}
	h.status = next
	return h, h.apply(command, action, code)
}

func (h Harvest) Cancel(command Command) (Harvest, error) {
	if replay, err := h.replay(command, "cancel", ""); replay || err != nil {
		return h, err
	}
	if !slices.Contains(h.members, command.ActorKey) || h.status == StatusAccepted || h.status == StatusCompleted ||
		h.status == StatusCancelled || h.status == StatusDeclined || h.status == StatusExpired {
		return Harvest{}, ErrTransition
	}
	h.status = StatusCancelled
	return h, h.apply(command, "cancel", "")
}

func (h Harvest) Expire(now time.Time, command Command) (Harvest, error) {
	if h.status != StatusReady || now.Before(h.expiresAt) {
		return Harvest{}, ErrTransition
	}
	h.status = StatusExpired
	return h, h.apply(command, "expire", "")
}

func (h Harvest) ProviderEnvelope(now time.Time) (Envelope, error) {
	if h.status != StatusHandedOff && h.status != StatusAccepted || !now.Before(h.expiresAt) {
		return Envelope{}, ErrTransition
	}
	return Envelope{HandoffID: h.handoffID, RecipeKey: h.payload.RecipeKey, RecipeVersion: h.payload.RecipeVersion,
		RenderSeed: h.payload.RenderSeed, ProductionTokens: append([]string(nil), h.payload.ProductionTokens...),
		Format: h.payload.Format, DeliveryRef: h.payload.DeliveryRef, PolicyVersion: h.payload.PolicyVersion,
		ExpiresAt: h.expiresAt}, nil
}

func (h *Harvest) apply(command Command, action, reasonCode string, values ...string) error {
	if !opaque.MatchString(command.ID) || command.At.IsZero() || command.ExpectedRevision != h.revision {
		return ErrStaleRevision
	}
	value := fingerprint(h.id, command, action, reasonCode, values...)
	h.revision++
	// Clone-on-append: Harvest methods use value receivers, so the slices
	// share backing arrays with the source aggregate. Appending into spare
	// capacity would corrupt every copy derived from the same state.
	h.events = append(append([]Event(nil), h.events...), Event{Sequence: h.revision, CommandID: command.ID, Action: action, Reason: reasonCode, At: command.At.UTC()})
	h.commands = append(append([]Applied(nil), h.commands...), Applied{ID: command.ID, Fingerprint: value, Revision: h.revision})
	return nil
}

func (h Harvest) replay(command Command, action, reasonCode string, values ...string) (bool, error) {
	want := fingerprint(h.id, command, action, reasonCode, values...)
	for _, applied := range h.commands {
		if applied.ID == command.ID {
			if applied.Fingerprint != want {
				return false, ErrCommandMismatch
			}
			return true, nil
		}
	}
	return false, nil
}

func validPayload(p Payload) bool {
	if !digest.MatchString(p.RecipeKey) || !opaque.MatchString(p.RecipeVersion) || !digest.MatchString(p.RenderSeed) ||
		!formats[p.Format] || !digest.MatchString(p.DeliveryRef) || !opaque.MatchString(p.PolicyVersion) ||
		len(p.ProductionTokens) != 6 {
		return false
	}
	seen := map[string]bool{}
	for _, token := range p.ProductionTokens {
		if !tokens[token] || seen[token] {
			return false
		}
		seen[token] = true
	}
	return true
}
func validPair(m []string) bool {
	return len(m) == 2 && m[0] != m[1] && digest.MatchString(m[0]) && digest.MatchString(m[1])
}
func sorted(v []string) []string { out := append([]string(nil), v...); slices.Sort(out); return out }
func clonePayload(p Payload) Payload {
	p.ProductionTokens = append([]string(nil), p.ProductionTokens...)
	return p
}
func payloadHash(p Payload) string {
	sum := sha256.Sum256([]byte(p.RecipeKey + "\x00" + p.RecipeVersion + "\x00" + p.RenderSeed + "\x00" +
		strings.Join(p.ProductionTokens, "\x00") + "\x00" + p.Format + "\x00" + p.DeliveryRef + "\x00" + p.PolicyVersion))
	return hex.EncodeToString(sum[:])
}
func fingerprint(id string, c Command, action, code string, values ...string) string {
	value := id + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + action + "\x00" + code + "\x00" +
		strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano) + "\x00" + strings.Join(values, "\x00")
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (h Harvest) ID() string              { return h.id }
func (h Harvest) HandoffID() string       { return h.handoffID }
func (h Harvest) Members() []string       { return append([]string(nil), h.members...) }
func (h Harvest) Payload() Payload        { return clonePayload(h.payload) }
func (h Harvest) Approvals() []string     { return append([]string(nil), h.approvals...) }
func (h Harvest) Status() Status          { return h.status }
func (h Harvest) ReadyAt() time.Time      { return h.readyAt }
func (h Harvest) ExpiresAt() time.Time    { return h.expiresAt }
func (h Harvest) Revision() uint64        { return h.revision }
func (h Harvest) Events() []Event         { return append([]Event(nil), h.events...) }
func (h Harvest) Commands() []Applied     { return append([]Applied(nil), h.commands...) }
func (h Harvest) HasMember(k string) bool { return slices.Contains(h.members, k) }
