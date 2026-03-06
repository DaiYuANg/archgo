# collectionx

`collectionx` 是 `arcgo` 的集合增强包，目标是提供：

- 泛型强类型集合（避免 `any` 和类型断言）
- 并发安全集合（基于 `sync.RWMutex`）
- `mo.Option` 风格读取 API
- Go 风格 API（零值可用、语义直接）

## 包结构（推荐）

实现已经按领域拆分到子包：

- `collectionx/set`：`Set`、`ConcurrentSet`、`MultiSet`、`OrderedSet`
- `collectionx/mapping`：`Map`、`ConcurrentMap`、`BiMap`、`OrderedMap`、`MultiMap`、`Table`
- `collectionx/list`：`List`、`ConcurrentList`、`Deque`、`RingBuffer`、`PriorityQueue`
- `collectionx/interval`：`Range`、`RangeSet`、`RangeMap`
- `collectionx/prefix`：`Trie` / `PrefixMap`
- `collectionx/tree`：`Tree`、`ConcurrentTree`（父子层级结构）

根包 `collectionx` 保留了兼容别名和构造器，历史调用方式不会中断。

实现中使用了：

- [`github.com/samber/lo`](https://github.com/samber/lo)（keys/values/filter/clone 等）
- [`github.com/samber/mo`](https://github.com/samber/mo)（`Option` 表达可空结果）

## Benchmark

已在子包内补充基准测试（`*_bench_test.go`），可按需执行：

```bash
# 跑 collectionx 全部 benchmark
go test ./collectionx/... -run ^$ -bench . -benchmem

# 只跑某个子包
go test ./collectionx/mapping -run ^$ -bench . -benchmem
go test ./collectionx/prefix -run ^$ -bench Trie -benchmem
```

## 当前版本（v1.1）

### `Set[T comparable]`

- `Add(items ...T)`
- `Remove(item T) bool`
- `Contains(item T) bool`
- `Len() int`
- `Clear()`
- `Values() []T`
- `Range(fn func(item T) bool)`
- `Clone() *Set[T]`
- `Union(other *Set[T]) *Set[T]`
- `Intersect(other *Set[T]) *Set[T]`
- `Difference(other *Set[T]) *Set[T]`

### `ConcurrentSet[T comparable]`

- `Add(items ...T)`
- `Remove(item T) bool`
- `Contains(item T) bool`
- `Len() int`
- `Clear()`
- `Values() []T`
- `Range(fn func(item T) bool)`（基于快照遍历）
- `Snapshot() *Set[T]`

### `Map[K comparable, V any]`

- `Set(key K, value V)`
- `SetAll(source map[K]V)`
- `Get(key K) (V, bool)`
- `GetOption(key K) mo.Option[V]`
- `GetOrDefault(key K, fallback V) V`
- `Delete(key K) bool`
- `Len() int`
- `Clear()`
- `Keys() []K`
- `Values() []V`
- `All() map[K]V`（返回副本）
- `Range(fn func(key K, value V) bool)`
- `Clone() *Map[K, V]`

### `ConcurrentMap[K comparable, V any]`

- `Set(key K, value V)`
- `SetAll(source map[K]V)`
- `Get(key K) (V, bool)`
- `GetOption(key K) mo.Option[V]`
- `GetOrDefault(key K, fallback V) V`
- `GetOrStore(key K, value V) (actual V, loaded bool)`
- `Delete(key K) bool`
- `LoadAndDelete(key K) (V, bool)`
- `LoadAndDeleteOption(key K) mo.Option[V]`
- `Len() int`
- `Clear()`
- `Keys() []K`
- `Values() []V`
- `All() map[K]V`（返回副本）
- `Range(fn func(key K, value V) bool)`（基于快照遍历）

### `List[T any]`（ArrayList 风格）

- `Add(items ...T)`
- `AddAt(index int, item T) bool`
- `AddAllAt(index int, items ...T) bool`
- `Get(index int) (T, bool)`
- `GetOption(index int) mo.Option[T]`
- `Set(index int, item T) bool`
- `RemoveAt(index int) (T, bool)`
- `RemoveAtOption(index int) mo.Option[T]`
- `RemoveIf(fn func(item T) bool) int`
- `Len() int`
- `Clear()`
- `Values() []T`（返回副本）
- `Range(fn func(index int, item T) bool)`
- `Clone() *List[T]`

### `ConcurrentList[T any]`

- `Add(items ...T)`
- `AddAt(index int, item T) bool`
- `AddAllAt(index int, items ...T) bool`
- `Get(index int) (T, bool)`
- `GetOption(index int) mo.Option[T]`
- `Set(index int, item T) bool`
- `RemoveAt(index int) (T, bool)`
- `RemoveAtOption(index int) mo.Option[T]`
- `RemoveIf(fn func(item T) bool) int`
- `Len() int`
- `Clear()`
- `Values() []T`（返回副本）
- `Range(fn func(index int, item T) bool)`（基于快照遍历）
- `Snapshot() *List[T]`

### `MultiMap[K comparable, V any]`

- `Put(key K, value V)`
- `PutAll(key K, values ...V)`
- `Set(key K, values ...V)`（替换）
- `Get(key K) []V`（返回副本）
- `GetOption(key K) mo.Option[[]V]`
- `Delete(key K) bool`
- `DeleteValueIf(key K, fn func(V) bool) int`
- `ContainsKey(key K) bool`
- `Len() int`（key 数）
- `ValueCount() int`（value 总数）
- `Keys() []K`
- `All() map[K][]V`（深拷贝）
- `Range(fn func(key K, values []V) bool)`

### `ConcurrentMultiMap[K comparable, V any]`

- 与 `MultiMap` 同语义，线程安全
- 额外提供 `Snapshot() *MultiMap[K, V]`

### `Table[R comparable, C comparable, V any]`（Guava Table 风格）

- `Put(row, col, value)`
- `Get(row, col) (V, bool)`
- `GetOption(row, col) mo.Option[V]`
- `SetRow(row, map[col]value)`（替换整行）
- `Row(row) map[col]value`（拷贝）
- `Column(col) map[row]value`（拷贝）
- `Delete(row, col) bool`
- `DeleteRow(row) bool`
- `DeleteColumn(col) int`
- `Has(row, col) bool`
- `RowCount() int`
- `Len() int`（cell 总数）
- `RowKeys() []R`
- `ColumnKeys() []C`
- `All() map[row]map[col]value`（深拷贝）
- `Range(fn func(row, col, value) bool)`

### `ConcurrentTable[R comparable, C comparable, V any]`

- 与 `Table` 同语义，线程安全
- 额外提供 `Snapshot() *Table[R, C, V]`

### `BiMap[K comparable, V comparable]`

- 双向一一映射（`K <-> V`）
- `Put(key, value)`（自动维护唯一性）
- `GetByKey / GetByValue`
- `GetValueOption / GetKeyOption`
- `DeleteByKey / DeleteByValue`
- `Keys / Values / All / Inverse`

### `MultiSet[T comparable]`（Bag）

- `Add / AddN`
- `Remove / RemoveN`
- `Count / Contains`
- `Len`（总元素数）/ `UniqueLen`（去重元素数）
- `Distinct / Elements / AllCounts`

### `OrderedMap[K comparable, V any]`（LinkedHashMap 风格）

- 保持 key 的插入顺序
- 更新已有 key 不改变顺序
- `Set / Get / Delete / At`
- `Keys / Values / Range`

### `OrderedSet[T comparable]`（LinkedHashSet 风格）

- 保持元素插入顺序
- `Add / Remove / Contains / At`
- `Values / Range`

### `Deque[T any]`

- 双端队列：`PushFront / PushBack`
- 双端弹出：`PopFront / PopBack`
- `Front / Back / Get`
- 动态扩容 ring-buffer 实现

### `RingBuffer[T any]`

- 固定容量循环缓冲区
- `Push` 满时覆盖最老元素并返回被淘汰值（`mo.Option`）
- `Pop / Peek / Values`

### `PriorityQueue[T any]`

- 泛型优先队列（基于 `container/heap`）
- 通过 `less(a, b)` 定义优先级（最小堆/最大堆都可）
- `Push / Pop / Peek`
- `ValuesSorted`（排序视图快照）

### `RangeSet[T cmp.Ordered]`

- 半开区间语义：`[start, end)`
- `Add / Remove / Contains / Overlaps`
- 自动归一化（排序、合并重叠和相邻区间）

### `RangeMap[T cmp.Ordered, V any]`

- 半开区间键：`[start, end) -> value`
- `Put` 时覆盖重叠区间
- `Get(point) / GetOption(point)`
- `DeleteRange / Entries`

### `Trie[V any]` / `PrefixMap[V any]`

- 字符串前缀树
- `Put / Get / Delete / Has / HasPrefix`
- `KeysWithPrefix / ValuesWithPrefix / RangePrefix`
- `NewPrefixMap` 是 `Trie` 的别名构造器

### `Tree[K comparable, V any]`

- 父子层级结构（Parent / Children）
- `AddRoot / AddChild / Move / Remove`
- `Parent / Children / Roots / Ancestors / Descendants`
- `Build([]Entry)` 可从平铺数据构建树

### `ConcurrentTree[K comparable, V any]`

- 与 `Tree` 同语义，线程安全
- 读 API 返回快照节点，不暴露内部可变结构
- 额外提供 `Snapshot() *Tree[K, V]`

## JSON 与日志输出快捷方法

多数结构已提供：

- `ToJSON() ([]byte, error)`：快速序列化
- `MarshalJSON() ([]byte, error)`：支持 `json.Marshal(x)` 直接输出
- `String() string`：日志友好的字符串输出

示例：

```go
s := set.NewSet[string]("a", "b")
raw, _ := s.ToJSON()
fmt.Println(string(raw)) // ["a","b"]
fmt.Println(s.String())  // ["a","b"]
```

## 示例

```go
package main

import (
	"fmt"

	"github.com/DaiYuANg/arcgo/collectionx"
)

func main() {
	// List
	list := collectionx.NewList("a", "c")
	list.AddAt(1, "b")
	fmt.Println(list.Values()) // [a b c]

	// MultiMap
	mm := collectionx.NewMultiMap[string, int]()
	mm.PutAll("tag", 1, 2, 3)
	fmt.Println(mm.Get("tag")) // [1 2 3]

	// Table
	tb := collectionx.NewTable[string, string, int]()
	tb.Put("u1", "score", 100)
	score, _ := tb.Get("u1", "score")
	fmt.Println(score) // 100

	// BiMap
	bm := collectionx.NewBiMap[string, int]()
	bm.Put("alice", 1)
	id, _ := bm.GetByKey("alice")
	fmt.Println(id) // 1

	// Trie
	tr := collectionx.NewTrie[int]()
	tr.Put("car", 1)
	tr.Put("cat", 2)
	fmt.Println(tr.KeysWithPrefix("ca")) // [car cat]

	// Tree
	org := collectionx.NewTree[int, string]()
	_ = org.AddRoot(1, "CEO")
	_ = org.AddChild(1, 2, "CTO")
	_ = org.AddChild(2, 3, "Platform Lead")
	fmt.Println(len(org.Descendants(1))) // 2
}
```
