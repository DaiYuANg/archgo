package dbx

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/samber/lo"
)

type PrimaryKeyMeta struct {
	Name    string
	Table   string
	Columns []string
}

type ForeignKeyMeta struct {
	Name          string
	Table         string
	Columns       []string
	TargetTable   string
	TargetColumns []string
	OnDelete      ReferentialAction
	OnUpdate      ReferentialAction
}

type CheckMeta struct {
	Name       string
	Table      string
	Expression string
}

type constraintBinder interface {
	bindConstraint(constraintBinding) any
}

type keyMetadata interface {
	keyMeta() keyBindingMeta
}

type constraintBinding struct {
	indexes    []IndexMeta
	primaryKey *PrimaryKeyMeta
	check      *CheckMeta
}

type keyBindingMeta struct {
	unique  bool
	primary bool
}

type Index[E any] struct {
	meta IndexMeta
}

type Unique[E any] struct {
	meta IndexMeta
}

type CompositeKey[E any] struct {
	meta PrimaryKeyMeta
}

type Check[E any] struct {
	meta CheckMeta
}

func (Index[E]) keyMeta() keyBindingMeta        { return keyBindingMeta{} }
func (Unique[E]) keyMeta() keyBindingMeta       { return keyBindingMeta{unique: true} }
func (CompositeKey[E]) keyMeta() keyBindingMeta { return keyBindingMeta{primary: true} }

func (i Index[E]) bindConstraint(binding constraintBinding) any {
	if len(binding.indexes) > 0 {
		i.meta = cloneIndexMeta(binding.indexes[0])
	}
	return i
}

func (u Unique[E]) bindConstraint(binding constraintBinding) any {
	if len(binding.indexes) > 0 {
		u.meta = cloneIndexMeta(binding.indexes[0])
	}
	return u
}

func (c CompositeKey[E]) bindConstraint(binding constraintBinding) any {
	if binding.primaryKey != nil {
		c.meta = clonePrimaryKeyMeta(*binding.primaryKey)
	}
	return c
}

func (c Check[E]) bindConstraint(binding constraintBinding) any {
	if binding.check != nil {
		c.meta = cloneCheckMeta(*binding.check)
	}
	return c
}

func (i Index[E]) Meta() IndexMeta             { return cloneIndexMeta(i.meta) }
func (u Unique[E]) Meta() IndexMeta            { return cloneIndexMeta(u.meta) }
func (c CompositeKey[E]) Meta() PrimaryKeyMeta { return clonePrimaryKeyMeta(c.meta) }
func (c Check[E]) Meta() CheckMeta             { return cloneCheckMeta(c.meta) }

func resolveConstraintBinding(def tableDefinition, field reflect.StructField, value any) (constraintBinding, error) {
	if key, ok := value.(keyMetadata); ok {
		tagName := "idx"
		if key.keyMeta().primary {
			tagName = "key"
		}
		options := parseTagOptions(field.Tag.Get(tagName))
		columns := splitColumnsOption(optionValue(options, "columns"))
		if len(columns) == 0 {
			return constraintBinding{}, fmt.Errorf("dbx: constraint %s on schema %s requires columns option", field.Name, def.schemaType.Name())
		}
		name := optionValue(options, "name")
		if name == "" {
			name = defaultConstraintName(def.name, field.Name, key.keyMeta())
		}
		if key.keyMeta().primary {
			return constraintBinding{
				primaryKey: &PrimaryKeyMeta{
					Name:    name,
					Table:   def.name,
					Columns: columns,
				},
			}, nil
		}
		return constraintBinding{
			indexes: []IndexMeta{{
				Name:    name,
				Table:   def.name,
				Columns: columns,
				Unique:  key.keyMeta().unique,
			}},
		}, nil
	}

	options := parseTagOptions(field.Tag.Get("check"))
	expression := strings.TrimSpace(optionValue(options, "expr"))
	if expression == "" {
		return constraintBinding{}, fmt.Errorf("dbx: check constraint %s on schema %s requires expr option", field.Name, def.schemaType.Name())
	}
	name := optionValue(options, "name")
	if name == "" {
		name = "ck_" + def.name + "_" + toSnakeCase(field.Name)
	}
	return constraintBinding{
		check: &CheckMeta{
			Name:       name,
			Table:      def.name,
			Expression: expression,
		},
	}, nil
}

func splitColumnsOption(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == '|' || r == ','
	})
	return lo.Compact(lo.Map(parts, func(part string, _ int) string {
		return strings.TrimSpace(part)
	}))
}

func defaultConstraintName(table, field string, meta keyBindingMeta) string {
	prefix := "idx"
	if meta.unique {
		prefix = "ux"
	}
	if meta.primary {
		prefix = "pk"
	}
	return prefix + "_" + table + "_" + toSnakeCase(field)
}

func cloneIndexMeta(meta IndexMeta) IndexMeta {
	meta.Columns = slices.Clone(meta.Columns)
	return meta
}

func clonePrimaryKeyMeta(meta PrimaryKeyMeta) PrimaryKeyMeta {
	meta.Columns = slices.Clone(meta.Columns)
	return meta
}

func cloneForeignKeyMeta(meta ForeignKeyMeta) ForeignKeyMeta {
	meta.Columns = slices.Clone(meta.Columns)
	meta.TargetColumns = slices.Clone(meta.TargetColumns)
	return meta
}

func cloneCheckMeta(meta CheckMeta) CheckMeta {
	return meta
}
