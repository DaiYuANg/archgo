package authx

import "errors"

var (
	ErrInvalidAuthenticationCredential    = errors.New("authx: invalid authentication credential")
	ErrInvalidAuthorizationModel          = errors.New("authx: invalid authorization model")
	ErrAuthenticationProviderNotFound     = errors.New("authx: authentication provider not found")
	ErrAuthenticationManagerNotConfigured = errors.New("authx: authentication manager not configured")
	ErrAuthorizerNotConfigured            = errors.New("authx: authorizer not configured")
	ErrUnauthenticated                    = errors.New("authx: unauthenticated")
)
