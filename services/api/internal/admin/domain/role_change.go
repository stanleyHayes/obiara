package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrRoleChangeIDNeeded = errors.New("role change id is required")
	ErrRoleChangeReason   = errors.New("role change reason must be between 12 and 240 characters")
	ErrRoleChangeClosed   = errors.New("role change is already closed")
	ErrSameApprover       = errors.New("role change needs a distinct second approver")
)

type RoleChangeStatus string

const (
	RoleChangePending  RoleChangeStatus = "pending"
	RoleChangeApproved RoleChangeStatus = "approved"
)

// RoleChange is a durable four-eyes proposal for adding or removing the admin
// role. It records only bounded operational data and never caller-supplied
// evidence or member data.
type RoleChange struct {
	id            string
	targetID      string
	targetVersion int64
	roles         []Role
	reason        string
	proposerID    string
	approverID    string
	status        RoleChangeStatus
	createdAt     time.Time
	approvedAt    *time.Time
}

func NewRoleChange(id, targetID string, targetVersion int64, roles []Role, reason, proposerID string, now time.Time) (RoleChange, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(targetID) == "" || strings.TrimSpace(proposerID) == "" {
		return RoleChange{}, ErrRoleChangeIDNeeded
	}
	if err := validateRoles(roles); err != nil {
		return RoleChange{}, err
	}
	reason = strings.TrimSpace(reason)
	if len(reason) < 12 || len(reason) > 240 {
		return RoleChange{}, ErrRoleChangeReason
	}
	return RoleChange{
		id: id, targetID: targetID, targetVersion: targetVersion,
		roles: append([]Role(nil), roles...), reason: reason,
		proposerID: proposerID, status: RoleChangePending, createdAt: now.UTC(),
	}, nil
}

func ReconstituteRoleChange(id, targetID string, targetVersion int64, roles []Role, reason, proposerID, approverID string, status RoleChangeStatus, createdAt time.Time, approvedAt *time.Time) RoleChange {
	return RoleChange{id: id, targetID: targetID, targetVersion: targetVersion, roles: roles, reason: reason, proposerID: proposerID, approverID: approverID, status: status, createdAt: createdAt, approvedAt: approvedAt}
}

func (change *RoleChange) Approve(approverID string, now time.Time) error {
	if change.status != RoleChangePending {
		return ErrRoleChangeClosed
	}
	if strings.TrimSpace(approverID) == "" || approverID == change.proposerID {
		return ErrSameApprover
	}
	at := now.UTC()
	change.approverID = approverID
	change.status = RoleChangeApproved
	change.approvedAt = &at
	return nil
}

func (change RoleChange) ID() string               { return change.id }
func (change RoleChange) TargetID() string         { return change.targetID }
func (change RoleChange) TargetVersion() int64     { return change.targetVersion }
func (change RoleChange) Roles() []Role            { return append([]Role(nil), change.roles...) }
func (change RoleChange) Reason() string           { return change.reason }
func (change RoleChange) ProposerID() string       { return change.proposerID }
func (change RoleChange) ApproverID() string       { return change.approverID }
func (change RoleChange) Status() RoleChangeStatus { return change.status }
func (change RoleChange) CreatedAt() time.Time     { return change.createdAt }
func (change RoleChange) ApprovedAt() *time.Time   { return change.approvedAt }
