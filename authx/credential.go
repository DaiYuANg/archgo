package authx

const (
	// CredentialKindPassword is username/password credential kind.
	CredentialKindPassword = "password"
	// CredentialKindAPIKey is API key credential kind.
	CredentialKindAPIKey = "api_key"
	// CredentialKindAnonymous is anonymous credential kind.
	CredentialKindAnonymous = "anonymous"
)

// Credential is the authentication input contract.
type Credential interface {
	Kind() string
}

// PasswordCredential represents username/password login input.
type PasswordCredential struct {
	Username string
	Password string
}

// Kind returns credential kind.
func (PasswordCredential) Kind() string {
	return CredentialKindPassword
}

// APIKeyCredential represents API key login input.
type APIKeyCredential struct {
	Key string
}

// Kind returns credential kind.
func (APIKeyCredential) Kind() string {
	return CredentialKindAPIKey
}

// AnonymousCredential represents anonymous visitor input.
type AnonymousCredential struct{}

// Kind returns credential kind.
func (AnonymousCredential) Kind() string {
	return CredentialKindAnonymous
}
