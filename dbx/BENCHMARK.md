# dbx Benchmarks

Run: `go test ./dbx ./dbx/migrate -run '^$' -bench . -benchmem -count=3`

## Memory vs IO Backends

SQLite benchmarks run two variants to isolate bottlenecks:

- **Memory**: `:memory:` SQLite — no disk I/O, CPU + alloc bound
- **IO**: temp file SQLite — real disk I/O, reflects production behavior

Compare ns/op and allocs: if Memory ≪ IO (Memory faster), the path is I/O-bound; if Memory ≈ IO (similar), the path is CPU-bound.

```
go test ./dbx -run '^$' -bench 'BenchmarkLoadBelongsTo|BenchmarkQueryAllStructMapper' -benchmem -count=2
```

## Bottleneck Summary (real sqlite, arm64)

| Benchmark | ns/op | allocs | Notes |
|-----------|-------|--------|-------|
| ValidateSchemasSQLiteAtlasMatched | ~262k | 629 | Schema diff + Atlas; cache+short-circuit applied |
| PlanSchemaChangesSQLiteAtlasEmpty | ~55k | 308 | Atlas schema planning; compiled cache applied |
| LoadManyToMany | ~35k | 160 | 3+ queries, join table scan |
| LoadBelongsTo | ~17k | 68 | 2 queries (parent + children) |
| LoadHasMany | ~12k | 97 | Batch relation load |
| QueryAllStructMapper | ~8k | 67 | Full query + scan |
| SQLList / SQLGet | ~5k | 34–44 | Statement + scan |
| BuildInsertUpsertReturning | ~2.3k | 47–51 | Query build |
| MapperInsertAssignments | ~800 | 11 | Assignment build |
| NewStructMapperCached | ~32 | 1 | Metadata cache hit |

## Optimization Priorities

1. **ValidateSchemas / PlanSchemaChanges** — Atlas + schema diff; consider caching compiled schema or reducing allocs.
2. **Relation loading** — Multiple round-trips; batch or reduce queries where possible.
3. **Query + scan path** — Mapper scan, column binding; already optimized with scan plan cache.
4. **Build* (render)** — SQL building; moderate allocs, acceptable for non-hot path.

---

## Optimization Ideas

### 1. ValidateSchemas / PlanSchemaChanges (~262ms, 629 allocs)

**Compiled schema cache**  
`compileAtlasSchema` 是纯 CPU，每次 ValidateSchemas 都会重新编译。可对 `(dialectName, schemaName, schemaFingerprint)` 做 cache，schema 在运行时通常不变。

**Short-circuit when matched**  
在 “matched” 场景（已 migrate 且无 diff）可提前退出：在内存中维护 DB schema 的 hash，若与期望 hash 一致则跳过 `driver.SchemaDiff` 和后续处理。

**Reduce Atlas allocs**  
`compileAtlasSchema` 会创建大量 atlas 小对象。可考虑复用或预分配常用结构（受限于 ariga/atlas 对外 API）。

### 2. Relation Loading (LoadManyToMany ~35ms, LoadBelongsTo ~17ms)

**Fewer round-trips**  
- LoadManyToMany：当前 2 次查询（join 表 + target 表）。可用单次 `JOIN` 同时取 pairs 与 target 行，减少一次 RTT。
- LoadBelongsTo/HasMany：已是一查，主要空间在 batch 的 `IN` 长度与结果处理。

**Prepared statement / bound query 复用**  
每次 `Build()` 会重新生成 SQL。对相同 (schema, relation, sourceKeys 模式) 的 relation load 可 cache `BoundQuery`，避免重复 build。  
→ ✅ 已实现：`relationQueryCache`（hot LRU 64），按 `(dialect, table, columns, inColumn, keyCount)` 缓存 SQL，复用 Build 结果。

**Alloc 优化**  
`collectSourceRelationKeys`、`indexRelationTargets`、`groupManyToManyTargets` 等会分配 map/slice。可用 `sync.Pool` 复用 scratch buffer，或预分配 capacity 降低 grow。  
→ ✅ `groupRelationTargets`、`groupManyToManyTargets` 已实现两遍预分配（先 count 再 alloc），减少 append 扩容。

### 3. QueryAll / SQLList (~5–8ms, 67 allocs)

**Bound query 复用**  
若同一 query 被频繁执行，用 `Build` 一次得到 `BoundQuery`，再多次调用 `QueryAllBound` / `QueryCursorBound`，避免每次 build。

**预分配结果 slice**  
有 `LIMIT` 时可 `make([]T, 0, limit)` 预分配，减少 append 时的多次扩容。  
→ ✅ 已实现：`BoundQuery.CapacityHint`（从 LIMIT 填充）+ `CapacityHintScanner`，`QueryAllBound` 使用预分配 slice。

### 4. 实施优先级建议

| 优先级 | 项 | 收益 | 改动量 |
|--------|-----|------|--------|
| 高 | Compiled schema cache | 消除重复 compile | ✅ 已实现 |
| 高 | Relation bound query 复用 | 减少 Build 和 alloc | ✅ 已实现 |
| 中 | LoadManyToMany 单 JOIN | 少一次 RTT | 中（需要改 query 结构） |
| 中 | Matched short-circuit | 跳过不必要的 diff | ✅ 已实现 |
| 低 | 预分配 group slice | 降低 relation load alloc | ✅ 已实现 |
| 低 | sync.Pool scratch | 进一步降低 relation alloc | ✅ 已实现 |

→ `seenSetPool`、`countsMapPool` 复用 `map[any]struct{}` 与 `map[any]int`，用于 `collectSourceRelationKeys`、`uniqueRelationKeysFromPairs`、`groupRelationTargets`、`groupManyToManyTargets`。

Schema/Atlas 相关路径一般不在请求级 hot path，relation load 和 query 的优化对接口响应更直接。

## Notes

- Benchmarks use SQLite files in temp dir (`b.TempDir()/bench.db`) for realistic disk I/O; tests use `:memory:` for speed.
- Hot paths (mapper scan, bind) are kept allocation-conscious.
- Schema/Atlas operations are not hot in typical request handling.
