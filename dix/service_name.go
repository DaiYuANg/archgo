package dix

import (
	"reflect"
	"sync"

	typetostring "github.com/samber/go-type-to-string"
)

var serviceNameCache sync.Map

func serviceNameOf[T any]() string {
	typ := reflect.TypeFor[T]()
	if name, ok := serviceNameCache.Load(typ); ok {
		return name.(string)
	}

	name := typetostring.GetReflectType(typ)
	actual, _ := serviceNameCache.LoadOrStore(typ, name)
	return actual.(string)
}
