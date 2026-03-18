package advanced

import (
	"github.com/DaiYuANg/arcgo/dix"
	do "github.com/samber/do/v2"
)

func typedName[T any]() string {
	return dix.TypedService[T]().Name
}

func invokeTyped[T any](injector do.Injector) (T, error) {
	return do.InvokeNamed[T](injector, typedName[T]())
}
