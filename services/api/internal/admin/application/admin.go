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
	ErrPrincipalNotFound = errors.New("admin principal not found")
	ErrPrincipalExists   = errors.New("admin principal already exists")
	ErrChallengeNotFound = errors.New("mfa challenge not found")
	ErrSessionNotFound   = errors.New("admin session not found")
	ErrNotAdmin          = errors.New("actor lacks the admin role for enrollment")
)

// PrincipalRepository persists admin principals.
type PrincipalRepository interface {
	Create(context.Context, domain.Principal) error
	FindByEmail(context.Context, string) (domain.Principal, error)
	FindByID(context.Context, string) (domain.Principal, error)
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
func (service AdminService) Enroll(ctx context.Context, actorID, email string, roles []domain.Role) (domain.Principal, error) {
	actor, err := service.principals.FindByID(ctx, actorID)
	if err != nil {
		return domain.Principal{}, err
	}
	if !actor.HasRole(domain.RoleAdmin) || actor.Status() != domain.StatusActive {
		return domain.Principal{}, ErrNotAdmin
	}

	principal, err := domain.NewPrincipal(service.newID(), email, roles, service.now())
	if err != nil {
		return domain.Principal{}, err
	}
	if err := service.principals.Create(ctx, principal); err != nil {
		return domain.Principal{}, err
	}
	if err := service.audit.Append(ctx, actorID, "admin.enroll", principal.ID(), service.now().UTC()); err != nil {
		return domain.Principal{}, err
	}
	return principal, nil
}

// StartLogin mints and sends an MFA code for an active principal.
func (service AdminService) StartLogin(ctx context.Context, email string) error {
	principal, err := service.principals.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if principal.Status() != domain.StatusActive {
		return ErrPrincipalNotFound
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
