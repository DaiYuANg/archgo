package configx

import (
	"github.com/go-playground/validator/v10"
	"github.com/samber/mo"
)

// Source documents related behavior.
type Source int

const (
	SourceDotenv Source = iota
	SourceFile
	SourceEnv
)

// ValidateLevel documents related behavior.
type ValidateLevel int

const (
	ValidateLevelNone ValidateLevel = iota
	ValidateLevelStruct
	ValidateLevelRequired
)

// Options loads related configuration.
type Options struct {
	dotenvFiles     []string
	envPrefix       string
	files           []string
	priority        []Source
	defaults        mo.Option[map[string]any]
	defaultsStruct  any
	validate        *validator.Validate
	validateLevel   ValidateLevel
	ignoreDotenvErr bool
}

// Option documents related behavior.
type Option func(*Options)

// NewOptions creates related functionality.
func NewOptions() *Options {
	return &Options{
		dotenvFiles:     []string{".env", ".env.local"},
		priority:        []Source{SourceDotenv, SourceFile, SourceEnv},
		validateLevel:   ValidateLevelNone,
		ignoreDotenvErr: true,
	}
}

// WithDotenv enables related functionality.
func WithDotenv(files ...string) Option {
	return func(o *Options) {
		if len(files) > 0 {
			o.dotenvFiles = files
		}
	}
}

// WithEnvPrefix configures related behavior.
func WithEnvPrefix(prefix string) Option {
	return func(o *Options) { o.envPrefix = prefix }
}

// WithFiles configures related behavior.
func WithFiles(files ...string) Option {
	return func(o *Options) { o.files = files }
}

// WithPriority configures related behavior.
func WithPriority(p ...Source) Option {
	return func(o *Options) { o.priority = p }
}

// WithDefaults configures related behavior.
func WithDefaults(m map[string]any) Option {
	return func(o *Options) {
		o.defaults = mo.Some(m)
	}
}

// WithDefaultsStruct configures related behavior.
func WithDefaultsStruct(s any) Option {
	return func(o *Options) {
		o.defaultsStruct = s
	}
}

// WithValidator configures related behavior.
func WithValidator(v *validator.Validate) Option {
	return func(o *Options) { o.validate = v }
}

// WithValidateLevel configures related behavior.
func WithValidateLevel(level ValidateLevel) Option {
	return func(o *Options) { o.validateLevel = level }
}

// WithIgnoreDotenvError configures related behavior.
func WithIgnoreDotenvError(ignore bool) Option {
	return func(o *Options) { o.ignoreDotenvErr = ignore }
}
