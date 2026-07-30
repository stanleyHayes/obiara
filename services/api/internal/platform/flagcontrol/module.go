package flagcontrol

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	flagmongo "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/adapters/outbound/mongodb"
	flagruntime "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/adapters/outbound/runtime"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/application"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flags"
)

type AdminAuthenticator interface {
	Authenticate(context.Context, string) (admindomain.Session, admindomain.Principal, error)
}

type ActorKeyer interface {
	Key(namespace, value string) (string, error)
}

type authority struct {
	admin AdminAuthenticator
	keyer ActorKeyer
}

func (a authority) RequireSteppedController(ctx context.Context, sessionID string, _ domain.Capability) (string, error) {
	session, principal, err := a.admin.Authenticate(ctx, strings.TrimSpace(sessionID))
	if err != nil || !session.SteppedUp() || !principal.HasRole(admindomain.RoleAdmin) {
		return "", application.ErrConflict
	}
	return a.keyer.Key("flag-controller", principal.ID())
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "flag_" + base64.RawURLEncoding.EncodeToString(value)
}

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

type Module struct {
	Controls ControlService
	Repo     *flagmongo.Repository
	Flags    *flags.Registry
}

type ControlService struct {
	service application.Service
	ctx     context.Context
}

func (service ControlService) Propose(ctx context.Context, command application.ProposeCommand) (domain.Proposal, error) {
	proposal, err := service.service.Propose(ctx, command)
	if err == nil {
		service.schedule(proposal)
	}
	return proposal, err
}

func (service ControlService) Approve(ctx context.Context, id, session string) (domain.Proposal, error) {
	return service.service.Approve(ctx, id, session)
}

func (service ControlService) Apply(ctx context.Context, id, session string) (domain.Proposal, error) {
	return service.service.Apply(ctx, id, session)
}

func (service ControlService) Expire(ctx context.Context, id string) (domain.Proposal, error) {
	return service.service.Expire(ctx, id)
}

func (service ControlService) schedule(proposal domain.Proposal) {
	delay := time.Until(proposal.ExpiresAt())
	if delay < 0 {
		delay = 0
	}
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-service.ctx.Done():
			return
		case <-timer.C:
			_, _ = service.service.Expire(context.WithoutCancel(service.ctx), proposal.ID())
		}
	}()
}

func NewModule(ctx context.Context, database *mongo.Database, admin AdminAuthenticator, keyer ActorKeyer, environment domain.Environment, getenv func(string) string) (Module, error) {
	configuration, err := flags.ParseEnvironment(getenv)
	if err != nil {
		return Module{}, err
	}
	registry := flags.New(configuration, nil, time.Now)
	repository := flagmongo.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	runtimeRegistry := flagruntime.NewRegistry(registry, environment, domain.MarketGH)
	service := application.New(
		repository,
		authority{admin: admin, keyer: keyer},
		runtimeRegistry,
		idSource{},
		clock{},
	)
	controls := ControlService{service: service, ctx: ctx}
	module := Module{Controls: controls, Repo: repository, Flags: registry}
	active, err := repository.ListActive(ctx, 100)
	if err != nil {
		return Module{}, err
	}
	now := time.Now().UTC()
	// Repository returns newest first for the desk. Replay oldest to newest so
	// the latest retained applied proposal wins for each capability.
	for index := len(active) - 1; index >= 0; index-- {
		proposal := active[index]
		if !now.Before(proposal.ExpiresAt()) {
			if _, err := service.Expire(ctx, proposal.ID()); err != nil {
				return Module{}, err
			}
			continue
		}
		if proposal.Status() == domain.StatusApplied {
			change, err := proposal.AppliedChange(now)
			if err != nil {
				return Module{}, err
			}
			if err := runtimeRegistry.Apply(ctx, proposal.State().Environment, proposal.State().Market, change); err != nil {
				return Module{}, err
			}
		}
		controls.schedule(proposal)
	}
	return module, nil
}
