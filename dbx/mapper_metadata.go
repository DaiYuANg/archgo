package dbx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/DaiYuANg/arcgo/collectionx"
)

type mapperMetadata struct {
	entityType         reflect.Type
	fields             collectionx.List[MappedField]
	byColumn           collectionx.Map[string, MappedField]
	byNormalizedColumn collectionx.Map[string, MappedField]
	scanPlans          collectionx.ConcurrentMap[string, *scanPlan]
}

func buildMapperMetadata(entityType reflect.Type) (*mapperMetadata, error) {
	if entityType.Kind() != reflect.Struct {
		return nil, ErrUnsupportedEntity
	}

	fields := collectionx.NewListWithCapacity[MappedField](entityType.NumField())
	byColumn := collectionx.NewMapWithCapacity[string, MappedField](entityType.NumField())
	byNormalizedColumn := collectionx.NewMapWithCapacity[string, MappedField](entityType.NumField())
	if err := collectMappedFields(entityType, nil, fields, byColumn, byNormalizedColumn); err != nil {
		return nil, err
	}

	return &mapperMetadata{
		entityType:         entityType,
		fields:             fields,
		byColumn:           byColumn,
		byNormalizedColumn: byNormalizedColumn,
		scanPlans:          collectionx.NewConcurrentMapWithCapacity[string, *scanPlan](8),
	}, nil
}

func resolveEntityColumn(field reflect.StructField) (string, map[string]bool) {
	raw := strings.TrimSpace(field.Tag.Get("dbx"))
	if raw == "-" {
		return "", nil
	}
	if raw == "" {
		return toSnakeCase(field.Name), map[string]bool{}
	}

	parts := strings.Split(raw, ",")
	name := strings.TrimSpace(parts[0])
	if name == "" {
		name = toSnakeCase(field.Name)
	}
	options := make(map[string]bool, len(parts)-1)
	for _, option := range parts[1:] {
		trimmed := strings.ToLower(strings.TrimSpace(option))
		if trimmed == "" {
			continue
		}
		options[trimmed] = true
	}
	return name, options
}

func collectMappedFields(entityType reflect.Type, prefix []int, fields collectionx.List[MappedField], byColumn collectionx.Map[string, MappedField], byNormalizedColumn collectionx.Map[string, MappedField]) error {
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		if !field.IsExported() {
			continue
		}
		rawTag := strings.TrimSpace(field.Tag.Get("dbx"))
		if rawTag == "-" {
			continue
		}

		path := appendIndexPath(prefix, i)
		if field.Anonymous && rawTag == "" {
			if embeddedType, ok := indirectStructType(field.Type); ok {
				if err := collectMappedFields(embeddedType, path, fields, byColumn, byNormalizedColumn); err != nil {
					return err
				}
				continue
			}
		}

		columnName, options := resolveEntityColumn(field)
		if options["inline"] {
			embeddedType, ok := indirectStructType(field.Type)
			if !ok {
				return fmt.Errorf("dbx: inline field %s must be a struct or pointer to struct", field.Name)
			}
			if err := collectMappedFields(embeddedType, path, fields, byColumn, byNormalizedColumn); err != nil {
				return err
			}
			continue
		}
		if columnName == "" {
			continue
		}

		mapped := MappedField{
			Name:       field.Name,
			Column:     columnName,
			Index:      path[0],
			Path:       path,
			Type:       field.Type,
			Insertable: !options["readonly"] && !options["-insert"] && !options["noinsert"],
			Updatable:  !options["readonly"] && !options["-update"] && !options["noupdate"],
		}
		fields.Add(mapped)
		byColumn.Set(columnName, mapped)
		normalized := normalizeResultColumnName(columnName)
		if normalized != "" {
			byNormalizedColumn.Set(normalized, mapped)
		}
	}
	return nil
}
