# sqlx-v2

[![Go Reference](https://pkg.go.dev/badge/github.com/rpadovani/sqlx-v2.svg)](https://pkg.go.dev/github.com/rpadovani/sqlx-v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/rpadovani/sqlx-v2)](https://goreportcard.com/report/github.com/rpadovani/sqlx-v2)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

`sqlx-v2` is an extension of the standard `database/sql` library, implementing generic type instantiation and offset-based pointer arithmetic for mapping SQL rows directly into Go structures, maps, and slices. It operates as a binary-compatible successor to `jmoiron/sqlx`.

---

## Overview

The standard `database/sql` library requires callers to iterate over result sets and pass variable references manually via `rows.Scan`.

### Standard `database/sql`
```go
for rows.Next() {
    var u User
    err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.Active)
    if err != nil {
        return nil, err
    }
    users = append(users, u)
}
```

### `sqlx-v2`
`sqlx-v2` automates struct mapping using generic interface parameters.
```go
users, err := sqlx.SelectG[User](ctx, db, "SELECT * FROM users")
```

---

## Usage

### 1. Requirements

```bash
go get github.com/rpadovani/sqlx-v2
```
Requires Go 1.24+ to compile the iterator and type parameter patterns.

### 2. Initialization

`sqlx-v2` delegates to standard Go database drivers (e.g. `pgx`, `mysql`, `sqlite3`).

```go
import (
    "github.com/rpadovani/sqlx-v2"
    _ "github.com/jackc/pgx/v5/stdlib"
)

db, err := sqlx.Connect("pgx", "postgres://user:pass@localhost/db")
```

### 3. Execution Data-Paths

```go
// Single row instantiation
user, err := sqlx.GetG[User](ctx, db, "SELECT * FROM users WHERE id = $1", id)

// Contiguous memory array instantiation
users, err := sqlx.SelectG[User](ctx, db, "SELECT * FROM users WHERE active = $1", true)

// Iterative sequence evaluation (O(1) memory bounds)
for user, err := range sqlx.SelectIter[User](ctx, db, "SELECT * FROM users") {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(user.Name)
}
```

---

## Performance Profile

`sqlx-v2` executes contiguous memory writes during result mapping. Memory footprint metrics are bound correctly by skipping dynamic reflection evaluation during the core row reading loop.

**[Read Benchmark Specification →](BENCHMARK.md)**

| Process (1,000 Rows, 50-Field Struct) | Execution Duration Cost | Output Memory Heap |
| :--- | :--- | :--- |
| `SelectIter[T]` | 17.8% reduction (vs v1) | 52.1% reduction (vs v1) |
| `SelectG[T]` | 28.7% reduction (vs v1) | 50.9% reduction (vs v1) |

---

## Implementation Characteristics

### Contiguous Memory Writes
`sqlx-v2` indexes structural dimensions into linear scalar byte offsets. During map iterations, it calculates fields by invoking a single `runtime.KeepAlive()` protected arithmetic statement (`P_target = uintptr(base) + offset`) to pass exact buffer addresses to the driver. The recursive tree is exclusively mapped once per struct definition lifetime.

### Primitive Yield Structures
Methods implement native generics, terminating immediately in fully allocated struct models, bypassing `interface{}` allocation steps.

### Buffered Synchronization
The system caches a `sync.Pool` of standardized output bindings. Output pointers dynamically reuse these boundaries to subtract slice duplication processes inside the map engine loop.

### Validation Matrices
Data paths undergo algorithmic bounds verification via Go property mapping frameworks, establishing bidirectional integrity (Write == Read). Buffer integrity spans execution sequences without crossing safety zones.

---

## Subsystem Interface

The library exports legacy signatures corresponding to internal definitions modeled by `jmoiron/sqlx` for binary-compatible drop-in. 

| Attribute | **sqlx-v2** | **v1 (jmoiron/sqlx)** |
| :--- | :--- | :--- |
| **Go requirement** | 1.24+ | Backwards compliant |
| **Address Generation**| Offset-based Calculation | Type Field Traversal |
| **Parameterization**| Generics Definitions | Interface Wrapping |
| **Compatibility Type**| Compliant Interfaces | Source Standard |

---

## Design and Runtime Mechanisms

The execution map assumes definitions conform to runtime constraint standards:

1. **KeepAlive Enforcement**: Evaluates object references explicitly so GC routines maintain allocation maps through native driver block scopes.
2. **Barrier Adherence**: Embedded object chaining defaults to `reflect.NewAt()` to uphold Go write-barriers during address generation paths.
3. **Execution Mapping**: Calculation arithmetic honors the runtime limits defined against `unsafe.Pointer` manipulation boundaries.

**Proceed to exact definitions: [ARCHITECTURE.md](ARCHITECTURE.md)**

---

## Development

Modifications process strictly inline with internal tracking specs: **[CONTRIBUTING.md](CONTRIBUTING.md)**.
Computational execution paths must pass evaluation frameworks like `benchstat`. Commits exhibiting >2% delta reductions inside critical bounds correctly isolate execution loops and halt commit progress limits.

## License

Apache 2.0. Reference [LICENSE](LICENSE).

## Disclaimer

This project is not an official Google project. It is not supported by Google and Google specifically disclaims all warranties as to its quality, merchantability, or fitness for any particular purpose.