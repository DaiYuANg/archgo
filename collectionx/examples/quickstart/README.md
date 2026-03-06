# collectionx quickstart

This example shows a complete 0-to-1 usage flow across the main `collectionx` subpackages:

- `set`
- `mapping` (`OrderedMap`, `MultiMap`, `Table`)
- `list`
- `prefix` (`Trie`)
- `interval` (`RangeSet`, `RangeMap`)

## Run

From repository root:

```bash
go run ./collectionx/examples/quickstart
```

## What to look for

- Set de-duplication behavior
- Ordered map insertion order behavior
- MultiMap one-key-to-many-values pattern
- Table row/column projections
- Deque and List basic operations
- Trie prefix lookup
- Interval merge/override behavior in `RangeSet` / `RangeMap`

