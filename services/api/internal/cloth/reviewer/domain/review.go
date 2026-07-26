package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusRedeemed Status = "redeemed"
	StatusRevoked  Status = "revoked"
)

const (
	MaxOTPValidity    = 10 * time.Minute
	MaxInviteValidity = 24 * time.Hour
	MaxQuestions      = 24
	MaxMaterials      = 12
)

var (
	ErrInvalid         = errors.New("invalid reviewer access")
	ErrExpired         = errors.New("reviewer access expired")
	ErrRedeemed        = errors.New("reviewer access already redeemed")
	ErrRevoked         = errors.New("reviewer access revoked")
	ErrCredential      = errors.New("invalid reviewer credential")
	ErrStaleRevision   = errors.New("stale reviewer revision")
	ErrCommandMismatch = errors.New("reviewer command replay mismatch")
)

var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var digest = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Command struct {
	ID               string
	ExpectedRevision uint64
	At               time.Time
}

type AppliedCommand struct {
	ID, Fingerprint string
	Revision        uint64
}

type Projection struct {
	ID, WatermarkRef string
	QuestionRefs     []string
	MaterialRefs     []string
	ExpiresAt        time.Time
}

type Review struct {
	id, reviewerKey, inviteDigest, otpDigest, sessionDigest, watermarkRef string
	members, questionRefs, materialRefs                                   []string
	status                                                                Status
	otpExpiresAt, inviteExpiresAt, redeemedAt, revokedAt                  time.Time
	revision                                                              uint64
	commands                                                              []AppliedCommand
}

type State struct {
	ID, ReviewerKey, InviteDigest, OTPDigest, SessionDigest, WatermarkRef string
	Members, QuestionRefs, MaterialRefs                                   []string
	Status                                                                Status
	OTPExpiresAt, InviteExpiresAt, RedeemedAt, RevokedAt                  time.Time
	Revision                                                              uint64
	Commands                                                              []AppliedCommand
}

func Create(state State, now time.Time, command Command) (Review, error) {
	review := fromState(state)
	if !review.validStatic() || state.Status != "" || state.Revision != 0 || len(state.Commands) != 0 ||
		now.IsZero() || command.At.IsZero() || command.ExpectedRevision != 0 ||
		!state.OTPExpiresAt.After(now) || state.OTPExpiresAt.After(now.Add(MaxOTPValidity)) ||
		!state.InviteExpiresAt.After(now) || state.InviteExpiresAt.After(now.Add(MaxInviteValidity)) ||
		state.OTPExpiresAt.After(state.InviteExpiresAt) {
		return Review{}, ErrInvalid
	}
	review.status = StatusPending
	if err := review.apply(command, "create"); err != nil {
		return Review{}, err
	}
	return review, nil
}

func Rehydrate(state State) (Review, error) {
	review := fromState(state)
	if !review.validStatic() || review.revision == 0 || len(review.commands) != int(review.revision) {
		return Review{}, ErrInvalid
	}
	switch review.status {
	case StatusPending:
		if review.sessionDigest != "" || !review.redeemedAt.IsZero() || !review.revokedAt.IsZero() {
			return Review{}, ErrInvalid
		}
	case StatusRedeemed:
		if !digest.MatchString(review.sessionDigest) || review.redeemedAt.IsZero() || !review.revokedAt.IsZero() {
			return Review{}, ErrInvalid
		}
	case StatusRevoked:
		if review.revokedAt.IsZero() {
			return Review{}, ErrInvalid
		}
	default:
		return Review{}, ErrInvalid
	}
	seen := map[string]bool{}
	for index, command := range review.commands {
		if !opaque.MatchString(command.ID) || !digest.MatchString(command.Fingerprint) ||
			command.Revision != uint64(index+1) || seen[command.ID] {
			return Review{}, ErrInvalid
		}
		seen[command.ID] = true
	}
	return review, nil
}

func (review Review) Redeem(inviteDigest, otpDigest, sessionDigest string, now time.Time, command Command) (Review, error) {
	if replay, err := review.replay(command, "redeem", sessionDigest); replay || err != nil {
		return review, err
	}
	if review.status == StatusRevoked {
		return Review{}, ErrRevoked
	}
	if review.status == StatusRedeemed {
		return Review{}, ErrRedeemed
	}
	if !now.Before(review.inviteExpiresAt) || !now.Before(review.otpExpiresAt) {
		return Review{}, ErrExpired
	}
	if inviteDigest != review.inviteDigest || otpDigest != review.otpDigest || !digest.MatchString(sessionDigest) {
		return Review{}, ErrCredential
	}
	review.status, review.sessionDigest, review.redeemedAt = StatusRedeemed, sessionDigest, now.UTC()
	if err := review.apply(command, "redeem", sessionDigest); err != nil {
		return Review{}, err
	}
	return review, nil
}

func (review Review) Revoke(now time.Time, command Command) (Review, error) {
	if replay, err := review.replay(command, "revoke"); replay || err != nil {
		return review, err
	}
	if review.status == StatusRevoked {
		return Review{}, ErrRevoked
	}
	if now.IsZero() {
		return Review{}, ErrInvalid
	}
	review.status, review.revokedAt = StatusRevoked, now.UTC()
	if err := review.apply(command, "revoke"); err != nil {
		return Review{}, err
	}
	return review, nil
}

func (review Review) Project(sessionDigest string, now time.Time) (Projection, error) {
	if review.status == StatusRevoked {
		return Projection{}, ErrRevoked
	}
	if review.status != StatusRedeemed || sessionDigest != review.sessionDigest {
		return Projection{}, ErrCredential
	}
	if !now.Before(review.inviteExpiresAt) {
		return Projection{}, ErrExpired
	}
	return Projection{
		ID: review.id, WatermarkRef: review.watermarkRef,
		QuestionRefs: append([]string(nil), review.questionRefs...),
		MaterialRefs: append([]string(nil), review.materialRefs...),
		ExpiresAt:    review.inviteExpiresAt,
	}, nil
}

func (review *Review) apply(command Command, action string, values ...string) error {
	if !opaque.MatchString(command.ID) || command.At.IsZero() || command.ExpectedRevision != review.revision {
		return ErrStaleRevision
	}
	fingerprint := fingerprint(review.id, command, action, values...)
	review.revision++
	review.commands = append(review.commands, AppliedCommand{ID: command.ID, Fingerprint: fingerprint, Revision: review.revision})
	return nil
}

func (review Review) replay(command Command, action string, values ...string) (bool, error) {
	want := fingerprint(review.id, command, action, values...)
	for _, applied := range review.commands {
		if applied.ID == command.ID {
			if applied.Fingerprint != want {
				return false, ErrCommandMismatch
			}
			return true, nil
		}
	}
	return false, nil
}

func fromState(state State) Review {
	return Review{
		id: state.ID, reviewerKey: state.ReviewerKey, inviteDigest: state.InviteDigest,
		otpDigest: state.OTPDigest, sessionDigest: state.SessionDigest, watermarkRef: state.WatermarkRef,
		members: append([]string(nil), state.Members...), questionRefs: append([]string(nil), state.QuestionRefs...),
		materialRefs: append([]string(nil), state.MaterialRefs...), status: state.Status,
		otpExpiresAt: state.OTPExpiresAt.UTC(), inviteExpiresAt: state.InviteExpiresAt.UTC(),
		redeemedAt: state.RedeemedAt.UTC(), revokedAt: state.RevokedAt.UTC(),
		revision: state.Revision, commands: append([]AppliedCommand(nil), state.Commands...),
	}
}

func (review Review) validStatic() bool {
	members := append([]string(nil), review.members...)
	slices.Sort(members)
	if !opaque.MatchString(review.id) || len(members) != 2 || members[0] == members[1] ||
		!slices.Equal(members, review.members) ||
		!digest.MatchString(members[0]) || !digest.MatchString(members[1]) ||
		!digest.MatchString(review.reviewerKey) || !digest.MatchString(review.inviteDigest) ||
		!digest.MatchString(review.otpDigest) || !digest.MatchString(review.watermarkRef) ||
		len(review.questionRefs) == 0 || len(review.questionRefs) > MaxQuestions ||
		len(review.materialRefs) > MaxMaterials || review.otpExpiresAt.IsZero() ||
		review.inviteExpiresAt.IsZero() {
		return false
	}
	review.members = members
	return validRefs(review.questionRefs) && validRefs(review.materialRefs)
}

func validRefs(refs []string) bool {
	seen := map[string]bool{}
	for _, ref := range refs {
		if !opaque.MatchString(ref) || seen[ref] {
			return false
		}
		seen[ref] = true
	}
	return true
}

func fingerprint(id string, command Command, action string, values ...string) string {
	value := id + "\x00" + command.ID + "\x00" + action + "\x00" +
		strconv.FormatUint(command.ExpectedRevision, 10) + "\x00" + command.At.UTC().Format(time.RFC3339Nano)
	for _, item := range values {
		value += "\x00" + item
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (review Review) ID() string                 { return review.id }
func (review Review) Members() []string          { return append([]string(nil), review.members...) }
func (review Review) ReviewerKey() string        { return review.reviewerKey }
func (review Review) InviteDigest() string       { return review.inviteDigest }
func (review Review) OTPDigest() string          { return review.otpDigest }
func (review Review) SessionDigest() string      { return review.sessionDigest }
func (review Review) WatermarkRef() string       { return review.watermarkRef }
func (review Review) QuestionRefs() []string     { return append([]string(nil), review.questionRefs...) }
func (review Review) MaterialRefs() []string     { return append([]string(nil), review.materialRefs...) }
func (review Review) Status() Status             { return review.status }
func (review Review) OTPExpiresAt() time.Time    { return review.otpExpiresAt }
func (review Review) InviteExpiresAt() time.Time { return review.inviteExpiresAt }
func (review Review) RedeemedAt() time.Time      { return review.redeemedAt }
func (review Review) RevokedAt() time.Time       { return review.revokedAt }
func (review Review) Revision() uint64           { return review.revision }
func (review Review) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), review.commands...)
}
func (review Review) HasMember(member string) bool    { return slices.Contains(review.members, member) }
func (review Review) IsReviewer(reviewer string) bool { return review.reviewerKey == reviewer }
func (review Review) HasCommand(commandID string) bool {
	for _, command := range review.commands {
		if command.ID == commandID {
			return true
		}
	}
	return false
}
