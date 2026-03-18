package dix

import (
	"log/slog"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	collectionset "github.com/DaiYuANg/arcgo/collectionx/set"
)

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

// AppState represents the current runtime state.
type AppState int32

const (
	AppStateCreated AppState = iota
	AppStateBuilt
	AppStateStarted
	AppStateStopped
)

// App is an immutable application specification.
type App struct {
	spec *appSpec
}

// Runtime is a built application runtime produced from an App spec.
type Runtime struct {
	spec      *appSpec
	plan      *buildPlan
	container *Container
	lifecycle *lifecycleImpl
	logger    *slog.Logger
	state     AppState
}

// Module is an immutable module specification.
type Module struct {
	spec *moduleSpec
}

type appSpec struct {
	meta    AppMeta
	profile Profile
	modules collectionlist.List[Module]
	logger  *slog.Logger
	debug   debugSettings
}

type moduleSpec struct {
	name            string
	description     string
	providers       collectionlist.List[ProviderFunc]
	setups          collectionlist.List[SetupFunc]
	invokes         collectionlist.List[InvokeFunc]
	hooks           collectionlist.List[HookFunc]
	imports         collectionlist.List[Module]
	profiles        collectionset.Set[Profile]
	excludeProfiles collectionset.Set[Profile]
	disabled        bool
	tags            collectionset.OrderedSet[string]
}

type debugSettings struct {
	scopeTree                bool
	namedServiceDependencies collectionset.OrderedSet[string]
}
