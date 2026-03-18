package dix

// Profile represents an application profile (environment).
type Profile string

const (
	ProfileDefault Profile = "default"
	ProfileDev     Profile = "dev"
	ProfileTest    Profile = "test"
	ProfileProd    Profile = "prod"
)

// AppMeta contains application metadata.
type AppMeta struct {
	Name        string
	Version     string
	Description string
}

// AppState represents the current state of the application.
type AppState int32

const (
	AppStateCreated AppState = iota
	AppStateBuilt
	AppStateStarted
	AppStateStopped
)

type debugSettings struct {
	scopeTree                bool
	namedServiceDependencies []string
}
