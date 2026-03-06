package configx

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
)

// loadDefaultsStruct loads related configuration.
func loadDefaultsStruct(k *koanf.Koanf, defaults any) error {
	defaultMap, err := structToMap(defaults)
	if err != nil {
		return fmt.Errorf("configx: convert defaults struct: %w", err)
	}
	if err := k.Load(confmap.Provider(defaultMap, "."), nil); err != nil {
		return fmt.Errorf("configx: load defaults struct into koanf: %w", err)
	}
	return nil
}

// structToMap converts related values.
// Note.
func structToMap(s any) (map[string]any, error) {
	// Case 1: already map[string]any
	if m, ok := s.(map[string]any); ok {
		return m, nil
	}

	if s == nil {
		return nil, &structToMapError{"expected struct or map, got <nil>"}
	}

	result := map[string]any{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &result,
		TagName: "mapstructure",
	})
	if err != nil {
		return nil, fmt.Errorf("configx: build defaults decoder: %w", err)
	}

	if err := decoder.Decode(s); err != nil {
		return nil, &structToMapError{"expected struct or map"}
	}

	return normalizeMapKeys(result), nil
}

type structToMapError struct {
	msg string
}

func (e *structToMapError) Error() string { return e.msg }

func normalizeMapKeys(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}

	out := make(map[string]any, len(input))
	for key, value := range input {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		out[normalizedKey] = normalizeValue(value)
	}
	return out
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeMapKeys(typed)
	case []any:
		return lo.Map(typed, func(item any, _ int) any {
			return normalizeValue(item)
		})
	default:
		return value
	}
}
