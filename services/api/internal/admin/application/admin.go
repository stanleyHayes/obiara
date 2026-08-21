// Package application runs admin enrollment, MFA login and step-up
// (E16-S01). Enrollment is a privileged action and always audited
// (FR-801; agent_plan.md §4).
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

var (
	ErrPrincipalNotFound  = errors.New("admin principal not found")
	ErrPrincipalExists    = errors.New("admin principal already exists")
	ErrChallengeNotFound  = errors.New("mfa challenge not found")
	ErrSessionNotFound    = errors.New("admin session not found")
	ErrNotAdmin           = errors.New("actor lacks the admin role for enrollment")
	ErrStepUpRequired     = errors.New("a recent MFA step-up is required")
	ErrSelfSuspension     = errors.New("an admin cannot suspend their own principal")
	ErrLastAdmin          = errors.New("the last active admin cannot be removed or suspended")
	ErrFourEyesRequired   = errors.New("admin role changes require a distinct second approver")
	ErrPrincipalConflict  = errors.New("admin principal changed concurrently")
	ErrRoleChangeNotFound = errors.New("admin role change not found")
)

// PrincipalRepository persists admin principals. Privileged mutations are
// audit-atomic: the mutation and its immutable audit entry commit in one
// transaction, and last-active-admin guards are enforced inside that same
// transaction so concurrent stepped-up admins cannot race past them.
type PrincipalRepository interface {
	FindByEmail(context.Context, string) (domain.Principal, error)
	FindByID(context.Context, string) (domain.Principal, error)
	List(context.Context) ([]domain.Principal, error)
	// CreateWithAudit creates the principal and appends the audit entry
	// (actorID, action, at) atomically.
	CreateWithAudit(ctx context.Context, principal domain.Principal, actorID, action string, at time.Time) error
	// UpdateWithAudit persists the principal mutation and the audit entry
	// atomically. When guardLastAdmin is true the transaction aborts with
	// ErrLastAdmin if the mutation would remove the last active admin.
	UpdateWithAudit(ctx context.Context, principal domain.Principal, guardLastAdmin bool, actorID, action string, at time.Time) error
	// CreateRoleChange records the proposal and its audit entry atomically;
	// guardLastAdmin enforces the last-active-admin invariant inside the
	// same transaction.
	CreateRoleChange(ctx context.Context, change domain.RoleChange, guardLastAdmin bool) error
	FindRoleChange(context.Context, string) (domain.RoleChange, error)
	ListPendingRoleChanges(context.Context) ([]domain.RoleChange, error)
	ApproveRoleChange(context.Context, domain.RoleChange, domain.Principal) error
}

// ChallengeRepository persists MFA challenges.
type ChallengeRepository interface {
	Create(context.Context, domain.Challenge) error
	LatestForPrincipal(context.Context, string) (domain.Challenge, error)
	Update(context.Context, domain.Challenge) error
}

// SessionRepository persists admin sessions.
type SessionRepository interface {
	Create(context.Context, domain.Session) error
	FindByID(context.Context, string) (domain.Session, error)
	Update(context.Context, domain.Session) error
}

// AccessAudit is the immutable admin-access audit store (FR-801).
type AccessAudit interface {
	Append(ctx context.Context, actorID, action, target string, at time.Time) error
}

// CodeSender delivers MFA codes (email channel bridge).
type CodeSender interface {
	SendMfaCode(ctx context.Context, email, code string) error
}

// AdminService runs admin auth.
type AdminService struct {
	principals PrincipalRepository
	challenges ChallengeRepository
	sessions   SessionRepository
	audit      AccessAudit
	sender     CodeSender
	now        func() time.Time
	newID      func() string
}

func NewAdminService(principals PrincipalRepository, challenges ChallengeRepository, sessions SessionRepository, audit AccessAudit, sender CodeSender, now func() time.Time, newID func() string) AdminService {
	return AdminService{principals: principals, challenges: challenges, sessions: sessions, audit: audit, sender: sender, now: now, newID: newID}
}

// Enroll creates a principal. The actor must hold the admin role, and the
// enrollment is always audited.
func (service AdminService) Enroll(ctx context.Context, sessionID, email string, roles []domain.Role) (domain.Principal, error) {
	_, actor, err := service.requireSteppedUpAdmin(ctx, sessionID)
	if err != nil {
		return domain.Principal{}, err
	}

	principal, err := domain.NewPrincipal(service.newID(), email, roles, service.now())
	if err != nil {
		return domain.Principal{}, err
	}
	if err := service.principals.CreateWithAudit(ctx, principal, actor.ID(), "admin.enroll", service.now().UTC()); err != nil {
		return domain.Principal{}, err
	}
	return principal, nil
}

// ListPrincipals returns the bounded operator directory to an authenticated admin.
func (service AdminService) ListPrincipals(ctx context.Context, sessionID string) ([]domain.Principal, error) {
	_, actor, err := service.Authenticate(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if !actor.HasRole(domain.RoleAdmin) {
		return nil, ErrNotAdmin
	}
	return service.principals.List(ctx)
}

// ChangeStatus suspends or reactivates an operator after MFA step-up.
func (service AdminService) ChangeStatus(ctx context.Context, sessionID, targetID string, status domain.Status) (domain.Principal, error) {
	_, actor, err := service.requireSteppedUpAdmin(ctx, sessionID)
	if err != nil {
		return domain.Principal{}, err
	}
	target, err := service.principals.FindByID(ctx, targetID)
	if err != nil {
		return domain.Principal{}, err
	}
	guardLastAdmin := false
	switch status {
	case domain.StatusSuspended:
		if actor.ID() == target.ID() {
			return domain.Principal{}, ErrSelfSuspension
		}
		// The last-active-admin check runs inside the mutation transaction so
		// concurrent stepped-up admins cannot both suspend the last admins.
		guardLastAdmin = target.HasRole(domain.RoleAdmin) && target.Status() == domain.StatusActive
		target.Suspend()
	case domain.StatusActive:
		target.Reactivate()
	default:
		return domain.Principal{}, domain.ErrInvalidStatus
	}
	if err := service.principals.UpdateWithAudit(ctx, target, guardLastAdmin, actor.ID(), "admin.principal.status."+string(status), service.now().UTC()); err != nil {
		return domain.Principal{}, err
	}
	return target, nil
}

// ChangeRoles applies non-admin role changes. Admin grants/revocations are
// deliberately routed to the separate four-eyes proposal flow.
func (service AdminService) ChangeRoles(ctx context.Context, sessionID, targetID string, roles []domain.Role) (domain.Principal, error) {
	_, actor, err := service.requireSteppedUpAdmin(ctx, sessionID)
	if err != nil {
		return domain.Principal{}, err
	}
	target, err := service.principals.FindByID(ctx, targetID)
	if err != nil {
		return domain.Principal{}, err
	}
	hasAdmin := false
	for _, role := range roles {
		hasAdmin = hasAdmin || role == domain.RoleAdmin
	}
	if hasAdmin != target.HasRole(domain.RoleAdmin) {
		return domain.Principal{}, ErrFourEyesRequired
	}
	if err := target.ReplaceRoles(roles); err != nil {
		return domain.Principal{}, err
	}
	if err := service.principals.UpdateWithAudit(ctx, target, false, actor.ID(), "admin.principal.roles", service.now().UTC()); err != nil {
		return domain.Principal{}, err
	}
	return target, nil
}

func (service AdminService) ProposeAdminRoleChange(ctx context.Context, sessionID, targetID string, roles []domain.Role, reason string) (domain.RoleChange, error) {
	_, actor, err := service.requireSteppedUpAdmin(ctx, sessionID)
	if err != nil {
		return domain.RoleChange{}, err
	}
	target, err := service.principals.FindByID(ctx, targetID)
	if err != nil {
		return domain.RoleChange{}, err
	}
	hasAdmin := false
	for _, role := range roles {
		hasAdmin = hasAdmin || role == domain.RoleAdmin
	}
	if hasAdmin == target.HasRole(domain.RoleAdmin) {
		return domain.RoleChange{}, ErrFourEyesRequired
	}
	// Revoking the admin role from an active admin enforces the
	// last-active-admin invariant inside the proposal transaction.
	guardLastAdmin := target.HasRole(domain.RoleAdmin) && !hasAdmin && target.Status() == domain.StatusActive
	change, err := domain.NewRoleChange(service.newID(), target.ID(), target.Version(), roles, reason, actor.ID(), service.now())
	if err != nil {
		return domain.RoleChange{}, err
	}
	if err := service.principals.CreateRoleChange(ctx, change, guardLastAdmin); err != nil {
		return domain.RoleChange{}, err
	}
	return change, nil
}

func (service AdminService) ListPendingRoleChanges(ctx context.Context, sessionID string) ([]domain.RoleChange, error) {
	_, actor, err := service.Authenticate(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if !actor.HasRole(domain.RoleAdmin) {
		return nil, ErrNotAdmin
	}
	return service.principals.ListPendingRoleChanges(ctx)
}

func (service AdminService) ApproveAdminRoleChange(ctx context.Context, sessionID, changeID string) (domain.Principal, error) {
	_, actor, err := service.requireSteppedUpAdmin(ctx, sessionID)
	if err != nil {
		return domain.Principal{}, err
	}
	change, err := service.principals.FindRoleChange(ctx, changeID)
	if err != nil {
		return domain.Principal{}, err
	}
	target, err := service.principals.FindByID(ctx, change.TargetID())
	if err != nil {
		return domain.Principal{}, err
	}
	if target.Version() != change.TargetVersion() {
		return domain.Principal{}, ErrPrincipalConflict
	}
	if err := change.Approve(actor.ID(), service.now()); err != nil {
		return domain.Principal{}, err
	}
	if err := target.ReplaceRoles(change.Roles()); err != nil {
		return domain.Principal{}, err
	}
	if err := service.principals.ApproveRoleChange(ctx, change, target); err != nil {
		return domain.Principal{}, err
	}
	return target, nil
}

func (service AdminService) requireSteppedUpAdmin(ctx context.Context, sessionID string) (domain.Session, domain.Principal, error) {
	session, actor, err := service.Authenticate(ctx, sessionID)
	if err != nil {
		return domain.Session{}, domain.Principal{}, err
	}
	if !actor.HasRole(domain.RoleAdmin) {
		return domain.Session{}, domain.Principal{}, ErrNotAdmin
	}
	if !session.SteppedUp() {
		return domain.Session{}, domain.Principal{}, ErrStepUpRequired
	}
	return session, actor, nil
}

// StartLogin mints and sends an MFA code for an active principal that has
// presented the correct password.
//
// Unknown emails, suspended principals and wrong passwords all no-op
// successfully: the endpoint is unauthenticated, so distinguishing them
// would turn it into an operator directory and a password oracle. Transport
// errors still surface. Callers therefore learn nothing from a success
// beyond "if that was a real operator with the right password, a code is on
// its way".
//
// A principal with no password digest keeps the code-only flow, so
// operators enrolled before passwords existed are not locked out.
func (service AdminService) StartLogin(ctx context.Context, email, password string) error {
	principal, err := service.principals.FindByEmail(ctx, email)
	if errors.Is(err, ErrPrincipalNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if principal.Status() != domain.StatusActive {
		return nil
	}
	if principal.HasPassword() && !principal.VerifyPassword(password) {
		return nil
	}
	return service.mintAndSend(ctx, principal.ID(), principal.Email())
}

func (service AdminService) mintAndSend(ctx context.Context, principalID, email string) error {
	code, err := domain.GenerateCode()
	if err != nil {
		return err
	}
	if err := service.challenges.Create(ctx, domain.NewChallenge(service.newID(), principalID, code, service.now())); err != nil {
		return err
	}
	return service.sender.SendMfaCode(ctx, email, code)
}

// CompleteLogin verifies a code and issues an admin session.
func (service AdminService) CompleteLogin(ctx context.Context, email, code string) (domain.Session, error) {
	principal, err := service.principals.FindByEmail(ctx, email)
	if err != nil {
		return domain.Session{}, err
	}
	challenge, err := service.challenges.LatestForPrincipal(ctx, principal.ID())
	if err != nil {
		return domain.Session{}, err
	}
	if err := challenge.Verify(code, service.now()); err != nil {
		_ = service.challenges.Update(ctx, challenge)
		return domain.Session{}, err
	}
	if err := service.challenges.Update(ctx, challenge); err != nil {
		return domain.Session{}, err
	}

	session := domain.NewSession(service.newID(), principal.ID(), principal.Roles(), service.now())
	if err := service.sessions.Create(ctx, session); err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

// StepUpStart mints and sends a step-up code for an active session.
func (service AdminService) StepUpStart(ctx context.Context, sessionID string) error {
	session, err := service.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if !session.Active(service.now()) {
		return domain.ErrSessionClosed
	}
	principal, err := service.principals.FindByID(ctx, session.PrincipalID())
	if err != nil {
		return err
	}
	return service.mintAndSend(ctx, principal.ID(), principal.Email())
}

// StepUpComplete verifies the step-up code and flags the session.
func (service AdminService) StepUpComplete(ctx context.Context, sessionID, code string) (domain.Session, error) {
	session, err := service.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return domain.Session{}, err
	}
	challenge, err := service.challenges.LatestForPrincipal(ctx, session.PrincipalID())
	if err != nil {
		return domain.Session{}, err
	}
	if err := challenge.Verify(code, service.now()); err != nil {
		_ = service.challenges.Update(ctx, challenge)
		return domain.Session{}, err
	}
	if err := service.challenges.Update(ctx, challenge); err != nil {
		return domain.Session{}, err
	}
	if err := session.MarkSteppedUp(service.now()); err != nil {
		return domain.Session{}, err
	}
	if err := service.sessions.Update(ctx, session); err != nil {
		return domain.Session{}, err
	}
	if err := service.audit.Append(ctx, session.PrincipalID(), "admin.step_up", session.ID(), service.now().UTC()); err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

// Authenticate resolves a short-lived bearer session and its active principal.
// Transport adapters use this instead of trusting caller-supplied roles.
func (service AdminService) Authenticate(ctx context.Context, sessionID string) (domain.Session, domain.Principal, error) {
	session, err := service.sessions.FindByID(ctx, sessionID)
	if err != nil || !session.Active(service.now()) {
		return domain.Session{}, domain.Principal{}, ErrSessionNotFound
	}
	principal, err := service.principals.FindByID(ctx, session.PrincipalID())
	if err != nil || principal.Status() != domain.StatusActive {
		return domain.Session{}, domain.Principal{}, ErrPrincipalNotFound
	}
	return session, principal, nil
}
