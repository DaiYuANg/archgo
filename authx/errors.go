package authx

import "errors"

var (
	// ErrInvalidCredential is returned when credential input is invalid.
	ErrInvalidCredential = errors.New("authx: invalid credential")
	// ErrInvalidAuthenticator is returned when authenticator definition is invalid.
	ErrInvalidAuthenticator = errors.New("authx: invalid authenticator")
	// ErrInvalidAuthorizer is returned when authorizer definition is invalid.
	ErrInvalidAuthorizer = errors.New("authx: invalid authorizer")
	// ErrInvalidPolicy is returned when policy input is invalid.
	ErrInvalidPolicy = errors.New("authx: invalid policy")
	// ErrInvalidRequest is returned when authorization request is invalid.
	ErrInvalidRequest = errors.New("authx: invalid authorization request")
	// ErrAuthenticatorNotFound is returned when no authenticator matches credential kind.
	ErrAuthenticatorNotFound = errors.New("authx: authenticator not found")
	// ErrDuplicateAuthenticator is returned when registering duplicate credential kind.
	ErrDuplicateAuthenticator = errors.New("authx: duplicate authenticator kind")
	// ErrUnauthorized indicates authentication failure.
	ErrUnauthorized = errors.New("authx: unauthorized")
	// ErrForbidden indicates authorization deny result.
	ErrForbidden = errors.New("authx: forbidden")
	// ErrNoIdentity indicates context has no attached identity.
	ErrNoIdentity = errors.New("authx: no identity in context")
)
