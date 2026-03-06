package configx

import (
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Loader loads related configuration.
type Loader struct {
	opts *Options
}

// Load loads related configuration.
func (l *Loader) Load(out any) error {
	cfg, err := l.loadInternal()
	if err != nil {
		return err
	}
	if err := cfg.k.Unmarshal("", out); err != nil {
		return err
	}
	return cfg.validateStruct(out)
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
		return mo.Err[T](err)
	}
	if err := cfg.validateStruct(out); err != nil {
		return mo.Err[T](err)
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
	k := koanf.New(".")

	// Note.
	if opts.defaults.IsPresent() {
		defaults, _ := opts.defaults.Get()
		if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
			return nil, err
		}
	}

	// Note.
	if opts.defaultsStruct != nil {
		if err := loadDefaultsStruct(k, opts.defaultsStruct); err != nil {
			return nil, err
		}
	}

	// Note.
	for _, src := range opts.priority {
		switch src {
		case SourceDotenv:
			if err := loadDotenv(k, opts.dotenvFiles, opts.ignoreDotenvErr); err != nil {
				return nil, err
			}
		case SourceFile:
			if err := loadFiles(k, opts.files); err != nil {
				return nil, err
			}
		case SourceEnv:
			if err := loadEnv(k, opts.envPrefix); err != nil {
				return nil, err
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
