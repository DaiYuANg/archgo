package configx

import (
	"reflect"
	"strings"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

// loadDefaultsStruct loads related configuration.
func loadDefaultsStruct(k *koanf.Koanf, defaults any) error {
	defaultMap, err := structToMap(defaults)
	if err != nil {
		return err
	}
	return k.Load(confmap.Provider(defaultMap, "."), nil)
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

	result := make(map[string]any)

	// Case 2: struct
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// Note.
	if t.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, &structToMapError{"expected struct or map, got " + t.Kind().String()}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Note.
		if !value.CanInterface() {
			continue
		}

		// Note.
		tag := field.Tag.Get("mapstructure")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		} else if tag == "-" {
			continue // note
		}

		// Note.
		if value.Kind() == reflect.Struct {
			nested, err := structToMap(value.Interface())
			if err != nil {
				return nil, err
			}
			result[tag] = nested
		} else {
			result[tag] = value.Interface()
		}
	}

	return result, nil
}

type structToMapError struct {
	msg string
}

func (e *structToMapError) Error() string { return e.msg }
