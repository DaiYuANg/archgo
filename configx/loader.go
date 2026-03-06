package configx

import (
	"context"
	"fmt"
	"time"

	"github.com/DaiYuANg/arcgo/observability"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

const (
	metricConfigLoadTotal            = "configx_load_total"
	metricConfigLoadDurationMS       = "configx_load_duration_ms"
	metricConfigSourceLoadTotal      = "configx_source_load_total"
	metricConfigSourceLoadDurationMS = "configx_source_load_duration_ms"
)

// Loader loads related configuration.
type Loader struct {
	opts *Options
}

// Load loads related configuration.
func (l *Loader) Load(out any) error {
	cfg, err := l.loadInternal()
	if err != nil {
		return fmt.Errorf("configx: load config: %w", err)
	}
	if err := cfg.k.Unmarshal("", out); err != nil {
		return fmt.Errorf("configx: unmarshal config into output: %w", err)
	}
	if err := cfg.validateStruct(out); err != nil {
		return fmt.Errorf("configx: validate output: %w", err)
	}
	return nil
}

// LoadConfig returns related data.
func (l *Loader) LoadConfig() (*Config, error) {
	return l.loadInternal()
}

func (l *Loader) loadInternal() (*Config, error) {
	return loadConfigFromOptions(l.opts)
}

// New creates related functionality.
func New(opts ...Option) *Loader {
	options := NewOptions()
	lo.ForEach(opts, func(opt Option, _ int) {
		if opt != nil {
			opt(options)
		}
	})
	return &Loader{opts: options}
}

// LoaderT loads related configuration.
type LoaderT[T any] struct {
	opts *Options
}

// Load loads related configuration.
func (l *LoaderT[T]) Load() mo.Result[T] {
	cfg, err := l.loadInternal()
	if err != nil {
		return mo.Err[T](err)
	}

	var out T
	if err := cfg.k.Unmarshal("", &out); err != nil {
		return mo.Err[T](fmt.Errorf("configx: unmarshal config into typed output: %w", err))
	}
	if err := cfg.validateStruct(out); err != nil {
		return mo.Err[T](fmt.Errorf("configx: validate typed output: %w", err))
	}
	return mo.Ok(out)
}

// LoadConfig returns related data.
func (l *LoaderT[T]) LoadConfig() (*Config, error) {
	return l.loadInternal()
}

func (l *LoaderT[T]) loadInternal() (*Config, error) {
	return loadConfigFromOptions(l.opts)
}

func loadConfigFromOptions(opts *Options) (*Config, error) {
	if opts == nil {
		opts = NewOptions()
	}

	obs := observability.Normalize(opts.observability, nil)
	ctx, span := obs.StartSpan(context.Background(), "configx.load")
	defer span.End()

	start := time.Now()
	result := "success"
	defer func() {
		obs.AddCounter(ctx, metricConfigLoadTotal, 1,
			observability.String("result", result),
		)
		obs.RecordHistogram(ctx, metricConfigLoadDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("result", result),
		)
	}()

	k := koanf.New(".")

	// Note.
	if opts.defaults.IsPresent() {
		defaults, _ := opts.defaults.Get()
		if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
			result = "error"
			span.RecordError(err)
			return nil, fmt.Errorf("configx: load defaults map: %w", err)
		}
	}

	// Note.
	if opts.defaultsStruct != nil {
		if err := loadDefaultsStruct(k, opts.defaultsStruct); err != nil {
			result = "error"
			span.RecordError(err)
			return nil, fmt.Errorf("configx: load defaults struct: %w", err)
		}
	}

	// Note.
	for _, src := range opts.priority {
		switch src {
		case SourceDotenv:
			if err := loadSourceWithObservability(ctx, obs, src, func() error {
				return loadDotenv(opts.dotenvFiles, opts.ignoreDotenvErr)
			}); err != nil {
				result = "error"
				span.RecordError(err)
				return nil, fmt.Errorf("configx: load dotenv source: %w", err)
			}
		case SourceFile:
			if err := loadSourceWithObservability(ctx, obs, src, func() error {
				return loadFiles(k, opts.files)
			}); err != nil {
				result = "error"
				span.RecordError(err)
				return nil, fmt.Errorf("configx: load file source: %w", err)
			}
		case SourceEnv:
			if err := loadSourceWithObservability(ctx, obs, src, func() error {
				return loadEnv(k, opts.envPrefix)
			}); err != nil {
				result = "error"
				span.RecordError(err)
				return nil, fmt.Errorf("configx: load env source: %w", err)
			}
		}
	}

	return newConfig(k, opts), nil
}

// NewT creates related functionality.
func NewT[T any](opts ...Option) *LoaderT[T] {
	options := NewOptions()
	lo.ForEach(opts, func(opt Option, _ int) {
		if opt != nil {
			opt(options)
		}
	})
	return &LoaderT[T]{opts: options}
}

// Load loads related configuration.
func Load(out any, opts ...Option) error {
	loader := New(opts...)
	return loader.Load(out)
}

// LoadT loads related configuration.
func LoadT[T any](opts ...Option) mo.Result[T] {
	loader := NewT[T](opts...)
	return loader.Load()
}

// LoadTErr loads typed config and returns regular (value, error) tuple.
func LoadTErr[T any](opts ...Option) (T, error) {
	result := LoadT[T](opts...)
	if result.IsError() {
		var zero T
		return zero, result.Error()
	}
	return result.Get()
}

// LoadConfig returns related data.
func LoadConfig(opts ...Option) (*Config, error) {
	loader := New(opts...)
	return loader.LoadConfig()
}

// LoadConfigT returns related data.
func LoadConfigT[T any](opts ...Option) (*Config, error) {
	loader := NewT[T](opts...)
	return loader.LoadConfig()
}

func loadSourceWithObservability(
	ctx context.Context,
	obs observability.Observability,
	source Source,
	fn func() error,
) error {
	if fn == nil {
		return nil
	}

	sourceName := source.String()
	sourceCtx, sourceSpan := obs.StartSpan(ctx, "configx.load."+sourceName,
		observability.String("source", sourceName),
	)
	defer sourceSpan.End()

	start := time.Now()
	result := "success"
	defer func() {
		obs.AddCounter(sourceCtx, metricConfigSourceLoadTotal, 1,
			observability.String("source", sourceName),
			observability.String("result", result),
		)
		obs.RecordHistogram(sourceCtx, metricConfigSourceLoadDurationMS, float64(time.Since(start).Milliseconds()),
			observability.String("source", sourceName),
			observability.String("result", result),
		)
	}()

	if err := fn(); err != nil {
		result = "error"
		sourceSpan.RecordError(err)
		return err
	}

	return nil
}
