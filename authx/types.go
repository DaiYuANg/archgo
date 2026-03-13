package authx

// AuthenticationResult stores identity resolved by authentication.
type AuthenticationResult struct {
	Principal any
	Details   map[string]any
}

// AuthorizationModel is the transport-agnostic input for authorization.
type AuthorizationModel struct {
	Principal any
	Action    string
	Resource  string
	Context   map[string]any
}

// Decision is the authorization output.
type Decision struct {
	Allowed  bool
	Reason   string
	PolicyID string
}

// Principal is the default identity shape used by built-in helpers.
type Principal struct {
	ID          string
	Roles       []string
	Permissions []string
	Attributes  map[string]any
}
