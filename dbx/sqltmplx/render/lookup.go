package render

import (
	"reflect"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

func lookup(params any, name string) mo.Option[any] {
	parts := strings.Split(name, ".")
	cur := params
	for _, part := range parts {
		next := lookupOne(cur, part)
		if next.IsAbsent() {
			return mo.None[any]()
		}
		cur = next.MustGet()
	}
	return mo.Some(cur)
}

func lookupOne(params any, name string) mo.Option[any] {
	v := reflect.ValueOf(params)
	for v.IsValid() && v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return mo.None[any]()
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return mo.None[any]()
	}
	if v.Kind() == reflect.Map {
		for _, key := range []string{name, strings.ToLower(name), strings.ToUpper(name)} {
			mv := v.MapIndex(reflect.ValueOf(key))
			if mv.IsValid() {
				return mo.Some(mv.Interface())
			}
		}
		return mo.None[any]()
	}
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if f.Name == name || strings.EqualFold(f.Name, name) {
				return mo.Some(v.Field(i).Interface())
			}
			for _, alias := range fieldAliases(f) {
				if alias == name || strings.EqualFold(alias, name) {
					return mo.Some(v.Field(i).Interface())
				}
			}
		}
	}
	return mo.None[any]()
}

func fieldAliases(f reflect.StructField) []string {
	var out []string
	for _, tagKey := range []string{"sqltmpl", "db", "json"} {
		raw := strings.TrimSpace(f.Tag.Get(tagKey))
		if raw == "" || raw == "-" {
			continue
		}
		alias := strings.TrimSpace(strings.Split(raw, ",")[0])
		if alias != "" && alias != "-" {
			out = append(out, alias)
		}
	}
	return lo.Uniq(out)
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	}
	return false
}

func isBlank(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return true
	}
	return lo.IsEmpty(rv.Interface())
}

func isPresent(v any) bool {
	return !isBlank(v)
}
