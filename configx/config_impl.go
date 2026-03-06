package configx

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/v2"
)

// Config documents related behavior.
type Config struct {
	k        *koanf.Koanf
	validate *validator.Validate
	level    ValidateLevel
}

func newConfig(k *koanf.Koanf, opts *Options) *Config {
	v := opts.validate
	if v == nil {
		v = validator.New()
	}
	return &Config{
		k:        k,
		validate: v,
		level:    opts.validateLevel,
	}
}

// validateStruct documents related behavior.
func (c *Config) validateStruct(out any) error {
	switch c.level {
	case ValidateLevelNone:
		return nil
	case ValidateLevelStruct, ValidateLevelRequired:
		return c.validate.Struct(out)
	default:
		return nil
	}
}

// Get retrieves related data.
func (c *Config) Get(path string) any {
	return c.k.Get(path)
}

// GetString retrieves related data.
func (c *Config) GetString(path string) string {
	return c.k.String(path)
}

// GetInt retrieves related data.
func (c *Config) GetInt(path string) int {
	return c.k.Int(path)
}

// GetInt64 retrieves related data.
func (c *Config) GetInt64(path string) int64 {
	return c.k.Int64(path)
}

// GetFloat64 retrieves related data.
func (c *Config) GetFloat64(path string) float64 {
	return c.k.Float64(path)
}

// GetBool retrieves related data.
func (c *Config) GetBool(path string) bool {
	return c.k.Bool(path)
}

// GetDuration retrieves related data.
func (c *Config) GetDuration(path string) time.Duration {
	return c.k.Duration(path)
}

// GetStringSlice retrieves related data.
func (c *Config) GetStringSlice(path string) []string {
	return c.k.Strings(path)
}

// GetIntSlice retrieves related data.
func (c *Config) GetIntSlice(path string) []int {
	return c.k.Ints(path)
}

// Unmarshal documents related behavior.
// path documents related behavior.
func (c *Config) Unmarshal(path string, out any) error {
	return c.k.Unmarshal(path, out)
}

// UnmarshalWithValidate documents related behavior.
// path documents related behavior.
func (c *Config) UnmarshalWithValidate(path string, out any) error {
	if err := c.k.Unmarshal(path, out); err != nil {
		return err
	}
	return c.validate.Struct(out)
}

// Exists checks related state.
func (c *Config) Exists(path string) bool {
	return c.k.Exists(path)
}

// All retrieves related data.
func (c *Config) All() map[string]any {
	return c.k.All()
}

// Validate documents related behavior.
func (c *Config) Validate(out any) error {
	return c.validate.Struct(out)
}
