package fx

import (
	"go.uber.org/fx"

	"github.com/DaiYuANg/arcgo/configx"
)

// ConfigParams defines parameters for configx module.
type ConfigParams struct {
	fx.In

	// Options for creating config.
	Options []configx.Option `optional:"true"`
}

// ConfigResult defines result for configx module.
type ConfigResult struct {
	fx.Out

	// Config is the created config.
	Config *configx.Config
}

// NewConfig creates a new config.
func NewConfig(params ConfigParams) (ConfigResult, error) {
	cfg, err := configx.NewConfig(params.Options...)
	if err != nil {
		return ConfigResult{}, err
	}
	return ConfigResult{Config: cfg}, nil
}

// NewConfigxModule creates a configx module.
func NewConfigxModule(opts ...configx.Option) fx.Option {
	return fx.Module("configx",
		fx.Provide(
			func() []configx.Option { return opts },
			NewConfig,
		),
	)
}

// NewConfigxModuleWithFiles creates a configx module with file sources.
func NewConfigxModuleWithFiles(files ...string) fx.Option {
	return NewConfigxModule(configx.WithFiles(files...))
}

// NewConfigxModuleWithEnv creates a configx module with environment variable sources.
func NewConfigxModuleWithEnv(prefix string) fx.Option {
	return NewConfigxModule(configx.WithEnvPrefix(prefix))
}

// NewConfigxModuleWithDotenv creates a configx module with dotenv file sources.
func NewConfigxModuleWithDotenv(files ...string) fx.Option {
	return NewConfigxModule(configx.WithDotenv(files...))
}
