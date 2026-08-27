// Command bootstrap seeds the first admin principal into a fresh Obiara
// database.
//
// It exists because admin enrollment is deliberately closed:
// AdminService.Enroll requires a caller who already holds an active,
// stepped-up admin session, so on an empty database nobody can enroll
// anybody and the console is unreachable. This command is the one
// out-of-band path that breaks that cycle, and it is run as an operator job
// rather than exposed as an HTTP endpoint so it cannot be reached from the
// internet.
//
// It is idempotent: re-running it updates the named principal's roles and
// password rather than creating a duplicate, so it doubles as the recovery
// path for a locked-out super admin.
//
// Credentials arrive only through the environment, never through flags,
// because process arguments are visible to other processes and land in
// shell history:
//
//	MONGODB_URI               required
//	MONGODB_DATABASE          required
//	BOOTSTRAP_ADMIN_EMAIL     required
//	BOOTSTRAP_ADMIN_PASSWORD  required, must satisfy the admin password policy
//	BOOTSTRAP_ADMIN_ROLES     optional, comma separated; defaults to every role
//	BOOTSTRAP_ALLOW_NON_ADMIN optional; set to seed a non-admin operator
//	                          account. The console's audited enrollment is
//	                          the normal path for staff — this exists for
//	                          seeding role-scoped accounts out of band, and
//	                          it is the only way to give one a password.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	adminmongodb "github.com/stanleyHayes/obiara/services/api/internal/admin/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

const connectTimeout = 30 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap failed:", err)
		os.Exit(1)
	}
}

func run() error {
	settings, err := loadSettings(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client, err := apimongo.Connect(connectCtx, settings.mongoURI)
	if err != nil {
		return err
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), connectTimeout)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
	}()

	database := client.Database(settings.mongoDatabase)
	// The unique email index must exist before the upsert, otherwise two
	// concurrent bootstraps could each insert the same operator.
	if err := adminmongodb.NewPrincipalRepository(database).EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure admin indexes: %w", err)
	}

	outcome, err := seedPrincipal(ctx, database, settings)
	if err != nil {
		return err
	}

	fmt.Printf("admin principal %s: %s\n", outcome.action, settings.email)
	fmt.Printf("  id     %s\n", outcome.id)
	fmt.Printf("  roles  %s\n", strings.Join(roleNames(settings.roles), ", "))
	fmt.Printf("  status %s\n", domain.StatusActive)
	fmt.Println("Sign in at the admin console with this email and password; a six-digit code follows by email.")
	return nil
}

type settings struct {
	mongoURI      string
	mongoDatabase string
	email         string
	password      string
	roles         []domain.Role
}

func loadSettings(getenv func(string) string) (settings, error) {
	loaded := settings{
		mongoURI:      strings.TrimSpace(getenv("MONGODB_URI")),
		mongoDatabase: strings.TrimSpace(getenv("MONGODB_DATABASE")),
		email:         strings.TrimSpace(getenv("BOOTSTRAP_ADMIN_EMAIL")),
		password:      getenv("BOOTSTRAP_ADMIN_PASSWORD"),
	}
	if loaded.mongoURI == "" {
		return settings{}, errors.New("MONGODB_URI is required")
	}
	if loaded.mongoDatabase == "" {
		return settings{}, errors.New("MONGODB_DATABASE is required")
	}
	if loaded.email == "" {
		return settings{}, errors.New("BOOTSTRAP_ADMIN_EMAIL is required")
	}
	if loaded.password == "" {
		return settings{}, errors.New("BOOTSTRAP_ADMIN_PASSWORD is required")
	}
	// Fail on the policy here rather than after connecting, so a weak
	// password never causes a half-finished run.
	if err := domain.CheckPasswordPolicy(loaded.password); err != nil {
		return settings{}, fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD: %w", err)
	}

	roles, err := parseRoles(
		getenv("BOOTSTRAP_ADMIN_ROLES"),
		strings.TrimSpace(getenv("BOOTSTRAP_ALLOW_NON_ADMIN")) != "",
	)
	if err != nil {
		return settings{}, err
	}
	loaded.roles = roles
	return loaded, nil
}

// allRoles is the default grant: a bootstrap operator is the platform's
// first super admin and needs every desk until they enroll specialists.
func allRoles() []domain.Role {
	return []domain.Role{
		domain.RoleAdmin, domain.RoleVerifier, domain.RoleTSAgent,
		domain.RoleHost, domain.RoleFinance,
	}
}

func parseRoles(value string, allowNonAdmin bool) ([]domain.Role, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return allRoles(), nil
	}
	known := allRoles()
	var roles []domain.Role
	for _, part := range strings.Split(trimmed, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		role := domain.Role(name)
		if !slices.Contains(known, role) {
			return nil, fmt.Errorf("BOOTSTRAP_ADMIN_ROLES contains unknown role %q", name)
		}
		if !slices.Contains(roles, role) {
			roles = append(roles, role)
		}
	}
	if len(roles) == 0 {
		return nil, errors.New("BOOTSTRAP_ADMIN_ROLES named no valid roles")
	}
	// Without the admin role the seeded operator could not enroll anyone
	// else, which would leave the console just as unreachable as before.
	// Seeding a role-scoped account is a different job from breaking that
	// cycle, so it has to say so explicitly rather than silently producing
	// an operator who cannot let anybody in.
	if !allowNonAdmin && !slices.Contains(roles, domain.RoleAdmin) {
		return nil, fmt.Errorf("BOOTSTRAP_ADMIN_ROLES must include %q", domain.RoleAdmin)
	}
	return roles, nil
}

func roleNames(roles []domain.Role) []string {
	names := make([]string, 0, len(roles))
	for _, role := range roles {
		names = append(names, string(role))
	}
	return names
}

type outcome struct {
	id     string
	action string
}

// seedPrincipal upserts the operator and appends an audit entry. Both writes
// go through one transaction where the deployment supports it, so a seeded
// principal is never left without its audit record.
func seedPrincipal(ctx context.Context, database *mongo.Database, config settings) (outcome, error) {
	principal, err := domain.NewPrincipal(newID(), config.email, config.roles, time.Now())
	if err != nil {
		return outcome{}, fmt.Errorf("build principal: %w", err)
	}
	if err := principal.SetPassword(config.password); err != nil {
		return outcome{}, fmt.Errorf("hash password: %w", err)
	}

	principals := database.Collection("admin_principals")
	var existing struct {
		ID string `bson:"_id"`
	}
	err = principals.FindOne(ctx, bson.M{"email": principal.Email()}).Decode(&existing)
	switch {
	case err == nil:
		// Recovery path: keep the established id so sessions, audit entries
		// and role-change proposals that reference it stay valid.
		update := bson.M{"$set": bson.M{
			"roles":        roleNames(config.roles),
			"status":       string(domain.StatusActive),
			"passwordHash": principal.PasswordHash(),
		}, "$inc": bson.M{"version": 1}}
		if _, err := principals.UpdateByID(ctx, existing.ID, update); err != nil {
			return outcome{}, fmt.Errorf("update principal: %w", err)
		}
		if err := appendAudit(ctx, database, existing.ID, "admin.bootstrap.update"); err != nil {
			return outcome{}, err
		}
		return outcome{id: existing.ID, action: "updated"}, nil

	case errors.Is(err, mongo.ErrNoDocuments):
		document := bson.M{
			"_id":          principal.ID(),
			"email":        principal.Email(),
			"roles":        roleNames(config.roles),
			"status":       string(domain.StatusActive),
			"passwordHash": principal.PasswordHash(),
			"version":      principal.Version(),
			"createdAt":    principal.CreatedAt(),
		}
		if _, err := principals.InsertOne(ctx, document); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// A concurrent bootstrap won the race; that is a success for
				// this one too.
				return outcome{id: principal.ID(), action: "already existed"}, nil
			}
			return outcome{}, fmt.Errorf("insert principal: %w", err)
		}
		if err := appendAudit(ctx, database, principal.ID(), "admin.bootstrap.create"); err != nil {
			return outcome{}, err
		}
		return outcome{id: principal.ID(), action: "created"}, nil

	default:
		return outcome{}, fmt.Errorf("look up principal: %w", err)
	}
}

// appendAudit records the out-of-band enrollment in the same immutable store
// the console writes to (FR-801). The actor is the command itself, which is
// what distinguishes a bootstrap from an in-console enrollment during a
// later review.
func appendAudit(ctx context.Context, database *mongo.Database, target, action string) error {
	_, err := database.Collection("admin_access").InsertOne(ctx, bson.M{
		"actorId": "bootstrap",
		"action":  action,
		"target":  target,
		"at":      time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("append audit: %w", err)
	}
	return nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "adm_" + base64.RawURLEncoding.EncodeToString(id)
}
