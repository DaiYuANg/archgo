package fx

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/DaiYuANg/arcgo/logx"
)

// LogParams defines parameters for logx module.
type LogParams struct {
	fx.In

	// Options for creating logger.
	Options []logx.Option `optional:"true"`
}

// LogResult defines result for logx module.
type LogResult struct {
	fx.Out

	// Logger is the created logger.
	Logger *logx.Logger
}

// NewLogger creates a new logger.
func NewLogger(params LogParams) (LogResult, error) {
	logger, err := logx.New(params.Options...)
	if err != nil {
		return LogResult{}, err
	}
	return LogResult{Logger: logger}, nil
}

// NewSlogLogger creates a slog.Logger from logx.Logger.
func NewSlogLogger(logger *logx.Logger) *slog.Logger {
	return logger.SlogLogger()
}

// NewLogxModule creates a logx module.
func NewLogxModule(opts ...logx.Option) fx.Option {
	return fx.Module("logx",
		fx.Provide(
			func() []logx.Option { return opts },
			NewLogger,
		),
	)
}

// NewLogxModuleWithSlog creates a logx module with slog.Logger support.
func NewLogxModuleWithSlog(opts ...logx.Option) fx.Option {
	return fx.Module("logx",
		fx.Provide(
			func() []logx.Option { return opts },
			NewLogger,
			NewSlogLogger,
		),
	)
}

// NewDevelopmentModule creates a development logx module.
func NewDevelopmentModule() fx.Option {
	return NewLogxModule(logx.DevelopmentConfig()...)
}

// NewProductionModule creates a production logx module.
func NewProductionModule(logPath string) fx.Option {
	return NewLogxModule(logx.ProductionConfig(logPath)...)
}
