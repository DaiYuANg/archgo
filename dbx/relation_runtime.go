package dbx

import (
	"sync"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/samber/hot"
)

// relationRuntime holds relation-load caches and pools per DB instance.
// Avoids package-level globals; enables per-instance config and test isolation.
type relationRuntime struct {
	queryCache   *hot.HotCache[string, string]
	seenSetPool  sync.Pool
	countsMapPool sync.Pool
}

func newRelationRuntime() *relationRuntime {
	rt := &relationRuntime{
		queryCache: hot.NewHotCache[string, string](hot.LRU, 64).Build(),
	}
	rt.seenSetPool = sync.Pool{New: func() any { return collectionx.NewMap[any, struct{}]() }}
	rt.countsMapPool = sync.Pool{New: func() any { return collectionx.NewMap[any, int]() }}
	return rt
}

// defaultRelationRuntime is used when Session does not provide RelationRuntime (e.g. custom Session impl).
var defaultRelationRuntime = newRelationRuntime()

type relationRuntimeProvider interface {
	RelationRuntime() *relationRuntime
}

func getRelationRuntime(session Session) *relationRuntime {
	if p, ok := session.(relationRuntimeProvider); ok {
		return p.RelationRuntime()
	}
	return defaultRelationRuntime
}
