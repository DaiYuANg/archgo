package configx

import "errors"

var errNilConfig = errors.New("config is nil")

// GetAs converts related values.
func GetAs[T any](cfg *Config, path string) (T, error) {
	var zero T
	if cfg == nil {
		return zero, errNilConfig
	}

	var out T
	if err := cfg.Unmarshal(path, &out); err != nil {
		return zero, err
	}
	return out, nil
}

// GetAsOr returns related data.
func GetAsOr[T any](cfg *Config, path string, fallback T) T {
	if cfg == nil {
		return fallback
	}
	if path != "" && !cfg.Exists(path) {
		return fallback
	}

	out, err := GetAs[T](cfg, path)
	if err != nil {
		return fallback
	}
	return out
}

// MustGetAs converts related values.
func MustGetAs[T any](cfg *Config, path string) T {
	out, err := GetAs[T](cfg, path)
	if err != nil {
		panic(err)
	}
	return out
}
