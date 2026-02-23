# sqlx-v2 Benchmark Data

Collected via `go test -bench=. -benchmem -count=20 -timeout=60m` and processed with [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat).

## Benchmark Performance Matrix (1,000 Rows)

| Benchmark | v1 (ns/op) | v2 Best (ns/op) | Δ Speed | v1 Alloc Size | v2 Alloc Size | Δ Mem Heap |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: |
| `SelectIter[T]` Large | 3.459m | 2.842m (iter) | −17.8% | 2.016MiB | 988.7KiB | −52.1% |
| `SelectG[T]` Large | 3.459m | 2.467m (generic) | −28.7% | 2.016MiB | 1014KiB | −50.9% |
| `SliceScan` Large | 2.850m | 1.651m (drop-in) | −42.1% | 1.645MiB | 1.736MiB | +5.5% |
| `SelectIter[T]` Medium| 1.113m | 747.3µ (iter) | −32.9% | 905.8KiB | 407.6KiB | −55.0% |
| `MapScan` Large | 7.316m | 6.450m (drop-in) | −11.8% | 5.887MiB | 5.979MiB | +1.6% |

## Full Selection Operation Matrix

### 1-Row Scale (ns/op ± %)

| Function | Size | v1 | v2 Drop-in | v2 Generic/Iter |
| :--- | :--- | ---: | ---: | ---: |
| `Select` | Small | 2.812µ ± 0% | 2.851µ ± 0% | 2.75µ ± 0% |
| `Select` | Medium | 4.269µ ± 0% | 3.646µ ± 0% | 3.899µ ± 0% |
| `Select` | Large | 15.44µ ± 0% | 13.82µ ± 0% | 14.81µ ± 0% |
| `Get` | Small | 2.891µ ± 0% | 2.657µ ± 0% | 2.662µ ± 0% |
| `Get` | Medium | 4.033µ ± 0% | 3.454µ ± 0% | 3.464µ ± 0% |
| `Get` | Large | 15.8µ ± 0% | 13.87µ ± 0% | 14.06µ ± 0% |

### 1,000-Row Scale (ns/op ± %)

| Function | Size | v1 | v2 Drop-in | v2 Generic/Iter |
| :--- | :--- | ---: | ---: | ---: |
| `Select` | Small | 337.5µ ± 0% | 310.6µ ± 0% | 164.4µ ± 0% |
| `Select` | Medium | 1.113m ± 0% | 972.3µ ± 0% | 605.6µ ± 0% |
| `Select` | Large | 3.459m ± 0% | 2.981m ± 0% | 2.467m ± 0% |

## Engine Diagnostics (1,000 Rows)

| Method | Size | v1 (ns/op) | v2 (ns/op) | Δ | allocs/op |
| :--- | :--- | ---: | ---: | ---: | ---: |
| `StructScan` | Small | 288.6µ | 250.9µ | −13.1% | 2,023 |
| `StructScan` | Medium | 880.5µ | 708.6µ | −19.5% | 2,023 |
| `StructScan` | Large | 3.047m | 2.418m | −20.6% | 4,124 |
| `MapScan` | Small | 509.4µ | 463.7µ | −9.0% | 4,022 |
| `MapScan` | Medium | 2.961m | 2.682m | −9.4% | 9,024 |
| `MapScan` | Large | 7.316m | 6.450m | −11.8% | 11,130 |
| `SliceScan` | Small | 298.4µ | 237.2µ | −20.5% | 3,022 |
| `SliceScan` | Medium | 968.9µ | 600.0µ | −38.1% | 3,022 |
| `SliceScan` | Large | 2.850m | 1.651m | −42.1% | 3,125 |

Data validates reduced internal allocation requirements across multiple `MapScan` and `SliceScan` vectors. `StructScan` registers mapped computational regressions against small-struct configurations.

## Component Specific Diagnostics

### NamedExec Cache Cost Modeling

| Variant | ns/op | allocs/op |
| :--- | ---: | ---: |
| v1 `NamedExec` | 979.9n ± 0% | 10 |
| v2 `NamedExec` (Warm) | 682n ± 0% | 8 |
| v2 `NamedExec` (Cold) | 1.92µ ± 0% | 9 |

The standard execution cache definition logs compliance fault bounds for mapping properties against v1 limits.

### GetG Stack Operation Analysis

| Variant | ns/op | B/op | allocs/op |
| :--- | ---: | ---: | ---: |
| v1 `Get` (Large) | 15.8µ ± 0% | 7,480 | 128 |
| v2 `Get` (Large) | 13.87µ ± 0% | 5,249 | 124 |
| v2 `GetG[T]` (Large) | 14.06µ ± 0% | 5,249 | 124 |

Stack-derived instantiation via `GetG[T]` requires ~30% smaller heap demands relative to v1 interface mappings.

### Memory Extent Functions

- `v1 Select` aggregates ~2.016 MiB heap allocations.
- `v2 SelectIter[T]` yields ~988.7 KiB (52.1% comparative margin decrease) at 2.842ms operation length.
- `v2 SelectG[T]` yields ~1014 KiB (50.9% comparative margin decrease) at 2.467ms operation length.

## Reproducibility Parameters

- **Hardware Architecture:** 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
- **Compiler Version:** 1.26.0 (linux/amd64)
- **Parameters:** `go test -bench=. -benchmem -count=20 -timeout=60m > final_bench.txt`
- **Output Validation:** `benchstat -delta-test=ttest v1.txt v2.txt v2g.txt`
