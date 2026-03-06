package authx

import (
	"fmt"
	"maps"
	"strings"
)

// UserDetails is an AuthX-owned authentication user snapshot.
type UserDetails struct {
	ID           string
	Principal    string
	PasswordHash string
	Name         string
	Payload      any
	Roles        []string
	Permissions  []string
	Attributes   map[string]string
}

func (u UserDetails) normalize() UserDetails {
	normalized := UserDetails{
		ID:           strings.TrimSpace(u.ID),
		Principal:    strings.TrimSpace(u.Principal),
		PasswordHash: strings.TrimSpace(u.PasswordHash),
		Name:         strings.TrimSpace(u.Name),
		Payload:      u.Payload,
		Roles:        uniqueStrings(u.Roles),
		Permissions:  uniqueStrings(u.Permissions),
		Attributes:   make(map[string]string, len(u.Attributes)),
	}

	for k, v := range maps.Clone(u.Attributes) {
		trimmedKey := strings.TrimSpace(k)
		if trimmedKey == "" {
			continue
		}
		normalized.Attributes[trimmedKey] = strings.TrimSpace(v)
	}

	if normalized.ID == "" {
		normalized.ID = normalized.Principal
	}
	if normalized.Name == "" {
		normalized.Name = normalized.Principal
	}

	return normalized
}

func (u UserDetails) validate() error {
	switch {
	case strings.TrimSpace(u.Principal) == "":
		return fmt.Errorf("%w: user principal is required", ErrInvalidAuthenticator)
	case strings.TrimSpace(u.PasswordHash) == "":
		return fmt.Errorf("%w: user password hash is required", ErrInvalidAuthenticator)
	case strings.TrimSpace(u.ID) == "":
		return fmt.Errorf("%w: user id is required", ErrInvalidAuthenticator)
	default:
		return nil
	}
}
