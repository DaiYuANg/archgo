package dbx

import (
	"reflect"

	"github.com/DaiYuANg/arcgo/collectionx"
)

type mapperRegistry struct {
	structMappers collectionx.ConcurrentMap[reflect.Type, *mapperMetadata]
}

type mapperRuntime struct {
	registry *mapperRegistry
}

var defaultMapperRuntime = newMapperRuntime()

func newMapperRuntime() mapperRuntime {
	return mapperRuntime{registry: newMapperRegistry()}
}

func newMapperRegistry() *mapperRegistry {
	return &mapperRegistry{
		structMappers: collectionx.NewConcurrentMap[reflect.Type, *mapperMetadata](),
	}
}

func getOrBuildStructMapperMetadata[E any]() (*mapperMetadata, error) {
	return getOrBuildMapperMetadata[E](defaultMapperRuntime.registry)
}

func getOrBuildMapperMetadata[E any](r *mapperRegistry) (*mapperMetadata, error) {
	entityType := reflect.TypeFor[E]()
	if cached, ok := r.structMappers.Get(entityType); ok {
		return cached, nil
	}

	mapper, err := buildMapperMetadata(entityType)
	if err != nil {
		return nil, err
	}
	actual, _ := r.structMappers.GetOrStore(entityType, mapper)
	return actual, nil
}
