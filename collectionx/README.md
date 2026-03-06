# collectionx

`collectionx` provides strongly typed collection data structures for Go, including concurrent-safe variants and non-standard structures (e.g. `MultiMap`, `Table`, `Trie`, interval structures).

[Chinese](./README_ZH.md)

## Why collectionx

Go standard containers are intentionally minimal. `collectionx` focuses on:

- Generic, strongly typed APIs
- Predictable semantics and explicit method names
- Optional concurrent-safe structures where needed
- Useful non-standard structures inspired by Java ecosystems

## Package Layout

- `collectionx/set`
  - `Set`, `ConcurrentSet`, `MultiSet`, `OrderedSet`
- `collectionx/mapping`
  - `Map`, `ConcurrentMap`, `BiMap`, `OrderedMap`, `MultiMap`, `Table`
- `collectionx/list`
  - `List`, `ConcurrentList`, `Deque`, `RingBuffer`, `PriorityQueue`
- `collectionx/interval`
  - `Range`, `RangeSet`, `RangeMap`
- `collectionx/prefix`
  - `Trie` / `PrefixMap`
- `collectionx/tree`
  - `Tree`, `ConcurrentTree` (parent-children hierarchy)

## 0-to-1 Runnable Example

- Quickstart directory: [collectionx/examples/quickstart](./examples/quickstart)
- Run from repo root:

```bash
go run ./collectionx/examples/quickstart
```

## Usage Scenarios

### 1) Fast de-duplication with `Set`

```go
s := set.NewSet[string]()
s.Add("A", "A", "B")
fmt.Println(s.Len()) // 2
fmt.Println(s.Contains("B"))
```

### 2) Preserve insertion order with `OrderedSet` / `OrderedMap`

```go
os := set.NewOrderedSet[int]()
os.Add(3, 1, 3, 2)
fmt.Println(os.Values()) // [3 1 2]

om := mapping.NewOrderedMap[string, int]()
om.Set("x", 1)
om.Set("y", 2)
om.Set("x", 9) // update does not move order
fmt.Println(om.Keys())   // [x y]
fmt.Println(om.Values()) // [9 2]
```

### 3) One key -> many values with `MultiMap`

```go
mm := mapping.NewMultiMap[string, int]()
mm.PutAll("tag", 1, 2, 3)
fmt.Println(mm.Get("tag"))        // [1 2 3]
fmt.Println(mm.ValueCount())       // 3
removed := mm.DeleteValueIf("tag", func(v int) bool { return v%2 == 0 })
fmt.Println(removed, mm.Get("tag")) // 1 [1 3]
```

### 4) 2D indexing with `Table` (Guava-style)

```go
t := mapping.NewTable[string, string, int]()
t.Put("r1", "c1", 10)
t.Put("r1", "c2", 20)
t.Put("r2", "c1", 30)

v, ok := t.Get("r1", "c2")
fmt.Println(v, ok) // 20 true
fmt.Println(t.Row("r1"))
fmt.Println(t.Column("c1"))
```

### 5) Prefix lookup with `Trie`

```go
tr := prefix.NewTrie[int]()
tr.Put("user:1", 1)
tr.Put("user:2", 2)
tr.Put("order:9", 9)

fmt.Println(tr.KeysWithPrefix("user:")) // [user:1 user:2]
```

### 6) Queueing and buffering with `list` package

```go
dq := list.NewDeque[int]()
dq.PushBack(1, 2)
dq.PushFront(0)
fmt.Println(dq.Values()) // [0 1 2]

rb := list.NewRingBuffer[int](2)
_ = rb.Push(1)
_ = rb.Push(2)
evicted := rb.Push(3) // evicts 1
fmt.Println(evicted)
```

### 7) Interval operations

```go
rs := interval.NewRangeSet[int]()
rs.Add(1, 5)
rs.Add(5, 8) // adjacent ranges are normalized
fmt.Println(rs.Ranges())

rm := interval.NewRangeMap[int, string]()
rm.Put(0, 10, "A")
rm.Put(3, 5, "B") // overlap override
v, _ := rm.Get(4)
fmt.Println(v) // B
```

### 8) Parent-children hierarchy with `Tree`

```go
org := tree.NewTree[int, string]()
_ = org.AddRoot(1, "CEO")
_ = org.AddChild(1, 2, "CTO")
_ = org.AddChild(2, 3, "Platform Lead")

parent, _ := org.Parent(3)
fmt.Println(parent.ID())          // 2
fmt.Println(len(org.Descendants(1))) // 2
```

## Concurrent-Safe Types: When To Use

Use concurrent variants only when shared access across goroutines is required:

- `ConcurrentSet`
- `ConcurrentMap`
- `ConcurrentMultiMap`
- `ConcurrentTable`
- `ConcurrentList`
- `ConcurrentTree`

For single-goroutine or externally synchronized workflows, non-concurrent types are usually faster.

## API Style Notes

- Most `All/Values/Row/Column` style methods return copies/snapshots to avoid accidental mutation leaks.
- `GetOption` methods use `mo.Option` for nullable-style reads.
- Zero-value behavior is supported by many structures, but constructors are still recommended for clarity.

## JSON And Logging Helpers

Most structures now provide:

- `ToJSON() ([]byte, error)` for quick serialization
- `MarshalJSON() ([]byte, error)` so `json.Marshal(x)` works directly
- `String() string` for log-friendly output

Example:

```go
s := set.NewSet[string]("a", "b")
raw, _ := s.ToJSON()
fmt.Println(string(raw))  // ["a","b"]
fmt.Println(s.String())   // ["a","b"]

payload, _ := json.Marshal(s) // same behavior via MarshalJSON
_ = payload
```

## Benchmark

```bash
go test ./collectionx/... -run ^$ -bench . -benchmem
```

You can target one package:

```bash
go test ./collectionx/mapping -run ^$ -bench . -benchmem
go test ./collectionx/prefix -run ^$ -bench Trie -benchmem
```

## Practical Tips

- Prefer `Table` when you otherwise use nested maps manually.
- Prefer `OrderedMap/OrderedSet` when result order matters (serialization, deterministic tests).
- Prefer `Trie` for high-volume prefix searches over repeated linear scans.
- Use `MultiSet` when counting frequencies is your primary operation.
- Prefer `Tree` when your model is naturally parent-children (org charts, categories, menu trees).

## FAQ

### Should I always use concurrent variants?

No. Use concurrent variants only when multiple goroutines share the same structure instance.  
If access is single-threaded or already externally synchronized, non-concurrent variants are simpler and faster.

### Are returned slices/maps safe to mutate?

For most snapshot-style APIs (`Values`, `All`, `Row`, `Column`, etc.), returned values are copies.  
Mutating the returned object usually does not mutate internal state.

### Why does `OrderedMap` keep old insertion order on update?

It intentionally behaves like insertion-order maps in other ecosystems: update changes value, not order.

### How does `RangeSet` handle adjacent intervals?

Adjacent intervals are normalized and merged for half-open ranges (for example `[1,5)` + `[5,8)`).

## Troubleshooting

### `Publish`-style code expects deterministic order but map-backed structures look random

`Map`, `Set`, and similar hash-backed structures do not guarantee iteration order.  
Use `OrderedMap` / `OrderedSet` if deterministic order is required.

### `Trie.KeysWithPrefix` allocates more than expected

Prefix collection returns new slices and traverses matched subtree.  
For hot paths:

- Narrow prefix as much as possible.
- Reuse `RangePrefix` callback style when possible.
- Avoid converting very large snapshots on each request.

### Memory grows in `MultiMap` or `Table` after long runtime

Common causes are unbounded key growth and missing cleanup paths.  
Use `Delete`, `DeleteColumn`, `DeleteRow`, `DeleteValueIf`, or periodic resets based on business lifecycle.

## Anti-Patterns

- Using `Concurrent*` structures everywhere by default.
- Depending on hash-map iteration order in tests or business logic.
- Treating snapshot-return APIs as live views and expecting in-place sync.
- Rebuilding huge temporary collections per request when incremental updates suffice.
- Using `RangeMap` for point lookups only; use plain maps if interval semantics are unnecessary.
