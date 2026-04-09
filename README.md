# sqlx-v2

[![Go Reference](https://pkg.go.dev/badge/github.com/rpadovani/sqlx-v2.svg)](https://pkg.go.dev/github.com/rpadovani/sqlx-v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/rpadovani/sqlx-v2)](https://goreportcard.com/report/github.com/rpadovani/sqlx-v2)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Zero-alloc, type-safe `database/sql` struct scanning. Drop-in replacement for [`jmoiron/sqlx`](https://github.com/jmoiron/sqlx) with Go 1.24+ generics and `iter.Seq2` support. Scans rows via O(1) pointer arithmetic instead of per-field reflection.

```go
// One row.
user, err := sqlx.GetG[User](ctx, db, "SELECT * FROM users WHERE id = $1", id)

// All rows into a slice.
users, err := sqlx.SelectG[User](ctx, db, "SELECT * FROM users WHERE active = $1", true)

// Streaming iterator — O(1) peak memory, no slice materialization.
for user, err := range sqlx.SelectIter[User](ctx, db, "SELECT * FROM users") {
    if err != nil {
        return err
    }
    process(user)
}
```

**Existing `jmoiron/sqlx` code compiles unchanged.** Change the import path, done:

```go
import "github.com/rpadovani/sqlx-v2"

db, err := sqlx.Connect("pgx", "postgres://user:pass@localhost/db")

var users []User
err = sqlx.SelectContext(ctx, db, &users, "SELECT * FROM users WHERE active = $1", true)
```

---

## Install

```bash
go get github.com/rpadovani/sqlx-v2
```

Requires **Go 1.24+**.

---

## Benchmarks

50-field struct, 1,000 rows. Full methodology and raw data: [BENCHMARK.md](BENCHMARK.md).

| Function | ns/op | vs v1 | B/op | vs v1 | allocs/op |
| :--- | ---: | ---: | ---: | ---: | ---: |
| `SelectG[T]` | 2.467ms | **−28.7%** | 1,014 KiB | **−50.9%** | 2,023 |
| `SelectIter[T]` | 2.842ms | **−17.8%** | 989 KiB | **−52.1%** | 2,023 |
| `Select` (drop-in) | 2.981ms | **−13.8%** | — | — | 2,023 |

Single-row `GetG[T]` on a 50-field struct: **14.06µs**, 5,249 B/op, 124 allocs — 30% less heap than v1.

Measured with `go test -bench=. -benchmem -count=20`, compared via `benchstat`.

---

## How It Works

**Discovery (once per type).** `internal/reflectx` walks the struct via `reflect`, computes the absolute byte offset of every field from the struct base, and caches the result in a `sync.Map` keyed by `reflect.Type`.

**Scan (per row).** Inside `rows.Next()`, each field's address is `base + offset` — a single pointer addition. The resulting pointers go directly into `rows.Scan`. A `sync.Pool` of `[]any` slices reuses scan buffers across rows.

**Pointer embeddings** (`*EmbeddedStruct`) can't use flat-offset math because the GC needs to track intermediate allocations. These fall back to `AddrByTraversal`, which uses `reflect.NewAt(...).Elem().Set()` to trigger the write barrier.

---

## Safety & Correctness

This library uses `unsafe.Pointer`. The following invariants are enforced across all scan paths:

- **GC Liveness.** Every scan path calls `runtime.KeepAlive(vp)` after `rows.Scan` completes. Without it, liveness analysis could mark the `reflect.New`-allocated struct as dead while the driver is still writing to it. `KeepAlive` anchors the allocation through the entire I/O window.

- **Write Barriers.** Fields reached through pointer embeddings cannot use flat-offset math — the GC must be notified of intermediate allocations. `AddrByTraversal` uses `reflect.NewAt(ptrType, ptr).Elem().Set(reflect.New(target))`, which triggers the runtime write barrier. No direct `*(*unsafe.Pointer)(p) = ...` assignments.

- **`unsafe.Pointer` Rule Compliance.** All arithmetic follows Go spec rules (3) and (4): `base := unsafe.Pointer(vp.Pointer())` then `ptr := unsafe.Pointer(uintptr(base) + offset)` — no `uintptr` intermediaries escape to separate expressions. Alignment is deferred to `reflect.StructField.Offset` (compiler-guaranteed).

Fuzz targets (`FuzzTypeMap`, `FuzzRowScanBounds`) verify offset bounds. Property-based tests enforce `Get(ID) == NamedExec(Struct)` symmetry. All tests run under `-race`.

---

## Unsafe Mode

By default, sqlx-v2 returns an error for columns without a matching struct field — catching schema drift early.

**Unsafe mode** discards unmatched columns silently. Useful for `SELECT *` when you only need a subset:

```go
db.Unsafe = true
// or: unsafeDB := db.UnsafeDB()  (original unchanged)
// or on a transaction: tx.Unsafe = true
```

```
// Safe mode error:
sqlx: missing destination name "extra_col" in struct MyStruct

// Unsafe mode: extra_col silently discarded. Matched columns scan normally.
```

Inherited from `jmoiron/sqlx` for full behavioral compatibility.

---

## v1 → v2

| | sqlx-v2 | jmoiron/sqlx (v1) |
| :--- | :--- | :--- |
| Go version | 1.24+ | any |
| Scan engine | Offset arithmetic, O(1) per field | Reflection tree traversal, O(depth) per field |
| Generic API | `SelectG[T]`, `GetG[T]`, `SelectIter[T]` | — |
| Classic API | `Select`, `Get`, `StructScan`, `MapScan`, `SliceScan` | same |
| `interface{}` boxing | Eliminated in generic path | Every field, every row |
| Import path | `github.com/rpadovani/sqlx-v2` | `github.com/jmoiron/sqlx` |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Architecture details: [ARCHITECTURE.md](ARCHITECTURE.md).

All changes must pass `go test -race ./...` and show no regression under `benchstat`.

## License

Apache 2.0 — [LICENSE](LICENSE).

## Disclaimer

This project is not an official Google project. It is not supported by Google and Google specifically disclaims all warranties as to its quality, merchantability, or fitness for any particular purpose.