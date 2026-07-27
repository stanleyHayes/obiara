// Package domain models in-app calls (E09-S09; FR-304: calls never reveal
// phone numbers; call metadata retention follows the Doc 09 table). Call
// participants are opaque member/session identifiers only — no contact
// detail ever enters this aggregate.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

// CallStatus is the call lifecycle.
type CallStatus string

const (
	StatusRinging CallStatus = "ringing"
	StatusEnded   CallStatus = "ended"
)

var (
	ErrCallIDRequired     = errors.New("call id is required")
	ErrParticipantMissing = errors.New("both call participants are required")
	ErrNotRoomMember      = errors.New("both participants must be members of the room")
	ErrCallNotOpen        = errors.New("call is not open")
)

// CallTTL bounds issued join tokens (E09: bounded sessions).
const CallTTL = 30 * time.Minute

// Call is one in-app call between two room members.
type Call struct {
	id           string
	roomID       string
	participants [2]string
	status       CallStatus
	version      int64
	createdAt    time.Time
	endedAt      *time.Time
}

// OpaqueKey derives a 64-hex opaque identifier from a part and value,
// matching the realtime adapter's identifier contract while keeping the
// underlying member/call ids unguessable.
func OpaqueKey(part, value string) string {
	sum := sha256.Sum256([]byte("obiara." + part + ":" + value))
	return hex.EncodeToString(sum[:])
}

// NewCall opens a call. RoomKey is the opaque LiveKit room identifier.
func NewCall(id, roomID string, participants [2]string, now time.Time) (Call, error) {
	if id == "" {
		return Call{}, ErrCallIDRequired
	}
	if participants[0] == "" || participants[1] == "" || participants[0] == participants[1] {
		return Call{}, ErrParticipantMissing
	}
	return Call{
		id:           id,
		roomID:       roomID,
		participants: participants,
		status:       StatusRinging,
		version:      1,
		createdAt:    now.UTC(),
	}, nil
}

// ReconstituteCall rebuilds a stored call without policy checks.
func ReconstituteCall(id, roomID string, participants [2]string, status CallStatus, version int64, createdAt time.Time, endedAt *time.Time) Call {
	return Call{id: id, roomID: roomID, participants: participants, status: status, version: version, createdAt: createdAt, endedAt: endedAt}
}

// End closes the call.
func (call *Call) End(now time.Time) error {
	if call.status != StatusRinging {
		return ErrCallNotOpen
	}
	call.status = StatusEnded
	ended := now.UTC()
	call.endedAt = &ended
	call.version++
	return nil
}

// LiveKitRoomKey is the opaque room key handed to the realtime adapter
// (its contract requires 64-hex identifiers).
func (call Call) LiveKitRoomKey() string {
	return OpaqueKey("call", call.id)
}

func (call Call) ID() string              { return call.id }
func (call Call) RoomID() string          { return call.roomID }
func (call Call) Participants() [2]string { return call.participants }
func (call Call) Status() CallStatus      { return call.status }
func (call Call) Version() int64          { return call.version }
func (call Call) CreatedAt() time.Time    { return call.createdAt }
func (call Call) EndedAt() *time.Time     { return call.endedAt }
