package authx

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aarondl/authboss/v3"
)

// UserDetailsMapper maps user details into AuthX identity.
type UserDetailsMapper func(ctx context.Context, user UserDetails) (Identity, error)

// AuthbossPasswordAuthenticator authenticates password credential via authboss.
type AuthbossPasswordAuthenticator struct {
	service IdentityProvider
	mapper  UserDetailsMapper
	logger  *slog.Logger
}

// NewAuthbossPasswordAuthenticator creates an authboss password authenticator.
func NewAuthbossPasswordAuthenticator(
	service IdentityProvider,
	mapper ...UserDetailsMapper,
) (*AuthbossPasswordAuthenticator, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: user details service is nil", ErrInvalidAuthenticator)
	}

	identityMapper := defaultUserDetailsMapper
	if len(mapper) > 0 && mapper[0] != nil {
		identityMapper = mapper[0]
	}

	return &AuthbossPasswordAuthenticator{
		service: service,
		mapper:  identityMapper,
		logger:  normalizeLogger(nil).With("component", "authx.authenticator", "name", "authboss-password"),
	}, nil
}

// SetLogger sets slog logger for this authenticator.
func (a *AuthbossPasswordAuthenticator) SetLogger(logger *slog.Logger) {
	if a == nil {
		return
	}
	a.logger = normalizeLogger(logger).With("component", "authx.authenticator", "name", "authboss-password")
}

// Name returns authenticator name.
func (a *AuthbossPasswordAuthenticator) Name() string {
	return "authboss-password"
}

// Kind returns supported credential kind.
func (a *AuthbossPasswordAuthenticator) Kind() string {
	return CredentialKindPassword
}

// Authenticate authenticates password credential by authboss verification.
func (a *AuthbossPasswordAuthenticator) Authenticate(ctx context.Context, credential Credential) (Identity, error) {
	if a == nil || a.service == nil || a.mapper == nil {
		return nil, fmt.Errorf("%w: authboss authenticator is not configured", ErrInvalidAuthenticator)
	}

	passwordCredential, ok := credential.(PasswordCredential)
	if !ok {
		return nil, fmt.Errorf("%w: expected password credential", ErrInvalidCredential)
	}

	principal := strings.TrimSpace(passwordCredential.Username)
	password := strings.TrimSpace(passwordCredential.Password)
	if principal == "" || password == "" {
		return nil, fmt.Errorf("%w: username/password is required", ErrInvalidCredential)
	}
	a.logger.Debug("authenticate started", "principal", principal, "kind", CredentialKindPassword)

	user, err := a.service.LoadByPrincipal(ctx, principal)
	if err != nil {
		a.logger.Warn("load principal failed", "principal", principal, "error", err.Error())
		return nil, err
	}

	normalizedUser := user.normalize()
	if err := normalizedUser.validate(); err != nil {
		return nil, err
	}

	if err := authboss.VerifyPassword(newAuthbossPasswordUser(normalizedUser.ID, normalizedUser.PasswordHash), password); err != nil {
		a.logger.Warn("password verification failed", "principal", principal)
		return nil, ErrUnauthorized
	}
	identity, err := a.mapper(ctx, normalizedUser)
	if err != nil {
		a.logger.Warn("identity mapping failed", "principal", principal, "error", err.Error())
		return nil, err
	}
	a.logger.Info("authenticate succeeded", "principal", principal, "principal_id", identity.ID())
	return identity, nil
}

func defaultUserDetailsMapper(ctx context.Context, user UserDetails) (Identity, error) {
	_ = ctx

	normalized := user.normalize()
	if err := normalized.validate(); err != nil {
		return nil, err
	}

	return NewIdentity(
		normalized.ID,
		"user",
		normalized.Name,
		WithPrincipal(normalized.Payload),
		WithRoles(normalized.Roles...),
		WithPermissions(normalized.Permissions...),
		WithAttributes(normalized.Attributes),
		WithAuthenticated(true),
	), nil
}

type authbossPasswordUser struct {
	pid      string
	password string
}

func newAuthbossPasswordUser(pid, passwordHash string) *authbossPasswordUser {
	return &authbossPasswordUser{
		pid:      strings.TrimSpace(pid),
		password: strings.TrimSpace(passwordHash),
	}
}

func (u *authbossPasswordUser) GetPID() string {
	return u.pid
}

func (u *authbossPasswordUser) PutPID(pid string) {
	u.pid = strings.TrimSpace(pid)
}

func (u *authbossPasswordUser) GetPassword() string {
	return u.password
}

func (u *authbossPasswordUser) PutPassword(password string) {
	u.password = strings.TrimSpace(password)
}
