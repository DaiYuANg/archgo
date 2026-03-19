package render

import (
	"reflect"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/hot"
)

var structMetadataCache = hot.NewHotCache[reflect.Type, *structMetadata](hot.LRU, 256).Build()

type structMetadata struct {
	fields []structFieldMetadata
	lookup map[string]structFieldMetadata
}

type structFieldMetadata struct {
	index      int
	name       string
	foldedName string
	aliases    []string
}

func cachedStructMetadata(t reflect.Type) *structMetadata {
	if cached, ok := structMetadataCache.Peek(t); ok {
		return cached
	}

	metadata := buildStructMetadata(t)
	if cached, ok := structMetadataCache.Peek(t); ok {
		return cached
	}
	structMetadataCache.Set(t, metadata)
	return metadata
}

func buildStructMetadata(t reflect.Type) *structMetadata {
	fields := lo.FilterMap(lo.Range(t.NumField()), func(index int, _ int) (structFieldMetadata, bool) {
		field := t.Field(index)
		if !field.IsExported() {
			return structFieldMetadata{}, false
		}

		return structFieldMetadata{
			index:      index,
			name:       field.Name,
			foldedName: strings.ToLower(field.Name),
			aliases:    fieldAliases(field),
		}, true
	})

	lookup := make(map[string]structFieldMetadata, len(fields)*3)
	lo.ForEach(fields, func(field structFieldMetadata, _ int) {
		lookup[field.foldedName] = field
		lo.ForEach(field.aliases, func(alias string, _ int) {
			lookup[strings.ToLower(alias)] = field
		})
	})

	return &structMetadata{
		fields: fields,
		lookup: lookup,
	}
}

func indirectValue(input any) (reflect.Value, bool) {
	value := reflect.ValueOf(input)
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}

	return value, value.IsValid()
}
