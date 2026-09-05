package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var ErrDeviceBlocklistFailed = errors.New("device blocklist write failed")

// ActionLog is the append-only action audit store.
type ActionLog interface {
	Append(context.Context, domain.ActionRecord) error
	CountForSubject(context.Context, string) (int, error)
	// AppliedCommand reports whether this operator request already took
	// effect. Checked before the ladder rather than after, because a retry
	// recomputes priors against a log that now includes the first attempt,
	// and the same action would be refused as off-ladder.
	AppliedCommand(ctx context.Context, commandID string) (bool, error)
}

// IdentityEnforcement applies account status effects (identity context
// port; the bridge runs at composition).
type IdentityEnforcement interface {
	Suspend(ctx context.Context, accountID string, until time.Time) error
	Block(ctx context.Context, accountID string) error
}

// SessionRevoker revokes all sessions for an account on enforcement.
type SessionRevoker interface {
	RevokeMemberSessions(ctx context.Context, memberID string) error
}

// DeviceBlocklister records device/biometric blocklist entries for Tier-A
// actions (Doc 09 §2: bans propagate to device and biometric blocklists).
type DeviceBlocklister interface {
	Blocklist(ctx context.Context, subjectID, reason string, at time.Time) error
}

// ActionService applies the T&S action ladder with propagated controls
// (E12-S04): the ladder check runs against the case tier and the subject's
// prior actions, then account status, sessions and device blocklists move
// together, and the action is immutably logged.
type ActionService struct {
	cases    CaseRepository
	actions  ActionLog
	identity IdentityEnforcement
	sessions SessionRevoker
	devices  DeviceBlocklister
	now      func() time.Time
	newID    func() string
}

func NewActionService(cases CaseRepository, actions ActionLog, identity IdentityEnforcement, sessions SessionRevoker, devices DeviceBlocklister, now func() time.Time, newID func() string) ActionService {
	return ActionService{cases: cases, actions: actions, identity: identity, sessions: sessions, devices: devices, now: now, newID: newID}
}

// Apply executes an action on an in-review case.
func (service ActionService) Apply(ctx context.Context, caseID string, action domain.Action, actorID, commandID string) error {
	applied, err := service.actions.AppliedCommand(ctx, commandID)
	if err != nil {
		return err
	}
	if applied {
		// The same decision, sent twice. Taking it once is the whole point.
		return nil
	}
	safetyCase, err := service.cases.FindByID(ctx, caseID)
	if err != nil {
		return err
	}

	priors, err := service.actions.CountForSubject(ctx, safetyCase.SubjectID())
	if err != nil {
		return err
	}
	if err := domain.CheckLadder(safetyCase.Tier(), action, priors); err != nil {
		return err
	}

	// Propagate the account/device controls before logging; a failed
	// propagation leaves no partial action record.
	now := service.now()
	switch {
	case action == domain.ActionBan:
		if err := service.identity.Block(ctx, safetyCase.SubjectID()); err != nil {
			return err
		}
		if err := service.devices.Blocklist(ctx, safetyCase.SubjectID(), "ban:"+caseID, now); err != nil {
			return ErrDeviceBlocklistFailed
		}
	case domain.IsSuspension(action):
		duration, _ := domain.SuspensionDuration(action)
		if err := service.identity.Suspend(ctx, safetyCase.SubjectID(), now.Add(duration)); err != nil {
			return err
		}
	}

	// Warnings change no account state; every non-warning revokes sessions.
	if action != domain.ActionWarning {
		if err := service.sessions.RevokeMemberSessions(ctx, safetyCase.SubjectID()); err != nil {
			return err
		}
	}

	if err := service.actions.Append(ctx, domain.ActionRecord{
		ID:        service.newID(),
		CommandID: commandID,
		CaseID:    safetyCase.ID(),
		SubjectID: safetyCase.SubjectID(),
		Action:    action,
		ActorID:   actorID,
		Priors:    priors,
		CreatedAt: now.UTC(),
	}); err != nil {
		return err
	}

	if err := safetyCase.Resolve(string(action), actorID, service.now()); err != nil {
		return err
	}
	return service.cases.Update(ctx, safetyCase)
}
