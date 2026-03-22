# dbx: collectionx / lo / mo 简化候选

扫描 dbx 实现，列出可用 collectionx、lo、mo 简化的代码位置。

## 高优先级（收益明显）

### 1. migrate/runner_internal.go

**indexAppliedRecords (L21–26)**  
`make(map) + for` → `lo.Associate`

```go
// 当前
indexed := make(map[string]AppliedRecord, len(records))
for _, record := range records {
    indexed[appliedRecordKey(record.Kind, record.Version, record.Description)] = record
}
return indexed

// 建议
return lo.Associate(records, func(record AppliedRecord) (string, AppliedRecord) {
    return appliedRecordKey(record.Kind, record.Version, record.Description), record
})
```

**appliedRecordForVersion (L124–130)**  
线性查找 → `lo.Find`

```go
// 当前
for _, item := range items {
    if item.Kind == record.Kind && item.Version == record.Version && item.Description == record.Description {
        return item, nil
    }
}
return AppliedRecord{}, fmt.Errorf(...)

// 建议
found, ok := lo.Find(items, func(item AppliedRecord) bool {
    return item.Kind == record.Kind && item.Version == record.Version && item.Description == record.Description
})
if !ok {
    return AppliedRecord{}, fmt.Errorf(...)
}
return found, nil
```

---

### 2. migrate/history_store.go

**GetLatestVersion (L112–118)**  
max 循环 → `lo.MaxBy`

```go
// 当前
var maxVersion int64
for _, item := range items {
    if item.Version > maxVersion {
        maxVersion = item.Version
    }
}
return maxVersion, nil

// 建议（items 为 []*goosedatabase.ListMigrationsResult，Version 为 int64）
if len(items) == 0 {
    return 0, nil
}
maxItem := lo.MaxBy(items, func(a, b *goosedatabase.ListMigrationsResult) bool {
    return a.Version > b.Version
})
return maxItem.Version, nil
```

---

### 3. table.go

**resolveTagNameAndOptions (L367–373)**  
`make(map) + for` → `lo.Associate` + `lo.FilterMap`

```go
// 当前
options := make(map[string]string, len(parts)-1)
for _, part := range parts[1:] {
    key, value := splitTagOption(part)
    if key != "" {
        options[key] = value
    }
}

// 建议
pairs := lo.FilterMap(parts[1:], func(part string, _ int) (lo.Entry[string, string], bool) {
    k, v := splitTagOption(part)
    if k == "" {
        return lo.Entry[string, string]{}, false
    }
    return lo.Entry[string, string]{Key: k, Value: v}, true
})
options := lo.Associate(pairs, func(e lo.Entry[string, string]) (string, string) { return e.Key, e.Value })
```

**parseTagOptions (L385–391)**  
同上模式。

**resolveColumnName (L343–356)**  
`for + continue + return` → `lo.Find`（需辅助函数或内联 predicate）

---

### 4. mapper_metadata.go

**resolveEntityColumn 内 options 构建 (L55–62)**  
与 table.go 的 `parseTagOptions` 相同，可用 `lo.FilterMap` + `lo.Associate`。

---

### 5. dialect/postgres/postgres.go

**parseIndexColumns (L347–354)**  
`for + filter + Add` → `lo.Compact` + `lo.Map`

```go
// 当前
columns := collectionx.NewListWithCapacity[string](len(parts))
for _, part := range parts {
    trimmed := strings.TrimSpace(strings.Trim(part, `"`))
    if trimmed != "" {
        columns.Add(trimmed)
    }
}
return columns.Values()

// 建议
return lo.Compact(lo.Map(parts, func(part string, _ int) string {
    return strings.TrimSpace(strings.Trim(part, `"`))
}))
```

---

### 6. schema_constraint.go

**splitColumnsOption (L165–171)**  
同上，`lo.Compact` + `lo.Map`：

```go
return lo.Compact(lo.Map(parts, func(part string, _ int) string {
    return strings.TrimSpace(part)
}))
```

---

### 7. projection.go

**projectionOfDefinition (L46–59)**  
`for + map lookup + validation` → `lo.FilterMap` + `lo.Find`

```go
// 当前逻辑可简化为
items := lo.FilterMap(fields, func(field MappedField, _ int) (SelectItem, bool) {
    column, ok := columns[field.Column]
    return schemaSelectItem{meta: column}, ok
})
if unmapped, ok := lo.Find(fields, func(field MappedField) bool {
    _, ok := columns[field.Column]
    return !ok
}); ok {
    return nil, &UnmappedColumnError{Column: unmapped.Column}
}
return items, nil
```

---

## Go 1.26 / 标准库 语言特性简化（已完成）

### slices.Clone 替代 `append([]T(nil), x...)`

项目为 Go 1.26，可使用 `slices.Clone`（Go 1.21+）替代手写 clone 模式，语义更清晰。

**已替换位置：**

| 文件 | 用途 |
|------|------|
| schema_constraint.go | cloneIndexMeta, clonePrimaryKeyMeta, cloneForeignKeyMeta |
| db.go | Hooks(), Bound() |
| tx.go | Bound() |
| sql_executor.go, sql_statement.go | Bind 时 Args clone |
| codec.go, codec_builtin.go | JSON/bytes clone |
| expression.go | CaseBuilder Branches |
| schema_migrate.go | buildTableSpec, clonePrimaryKeyState |
| schema_migrate_atlas.go | MissingColumns/Indexes/ForeignKeys/Checks, Args |
| sqltmplx/template.go | BoundQuery Args |

### Go 1.26 新增特性（dbx 中暂无可直接应用）

- **`new(expr)`**：可用表达式初始化指针，适用于 `encoding/json` 等可选字段场景；dbx 中多为 struct 字面量 `&X{}`，暂无可简化处。
- **自引用泛型**：`type Adder[A Adder[A]]` 等约束；dbx 当前泛型设计无需此项。

---

## 中优先级（可选）

### 8. dialect/mysql/mysql.go, dialect/sqlite/sqlite.go

类似 postgres 的 `parseIndexColumns`，存在 `for + append` 构建 columns / foreign key state，可考虑 `lo.Map`、`lo.FilterMap` 等简化，但需结合错误处理与状态累积逻辑。

---

## 低优先级 / 保持现状

- **relation_load_internal.go**：`indexRelationTargets`、`groupRelationTargets`、`groupManyToManyTargets` 含错误处理、条件分支和两阶段逻辑，函数式改写未必更清晰。
- **migrate/runner.go loadSQLMigrations**：含 I/O 与错误传播，适合保留显式循环。
- **schema_migrate.go AutoMigrate 执行循环**：需要提前返回错误，现有写法已足够直观。

---

## 汇总

| 文件 | 行号 | 模式 | 建议 |
|------|------|------|------|
| migrate/runner_internal.go | 21–26 | make(map) + for | lo.Associate |
| migrate/runner_internal.go | 124–130 | 线性查找 | lo.Find |
| migrate/history_store.go | 112–118 | max 循环 | lo.MaxBy |
| table.go | 367–373, 385–391 | make(map) + for | lo.FilterMap + lo.Associate |
| table.go | 343–356 | for + early return | lo.Find + helper |
| mapper_metadata.go | 55–62 | make(map) + for | lo.FilterMap + lo.Associate |
| dialect/postgres/postgres.go | 347–354 | for + filter + add | lo.Compact + lo.Map |
| schema_constraint.go | 165–171 | for + filter + add | lo.Compact + lo.Map |
| projection.go | 46–59 | for + map + validation | lo.FilterMap + lo.Find |
