# AI Agent Guidelines for sqlx-v2

If you are an autonomous AI agent or an LLM assisting a developer in this repository, you must strictly adhere to the rules and architectural principles defined below. `sqlx-v2` is an extension of `jmoiron/sqlx` using Go 1.24+ Generics and Iterators.

## 1. Core Priorities
Every change you propose or implement must evaluate trade-offs according to the following strict hierarchy:

1. **Compatibility (Drop-in Replacement):** The absolute highest priority is remaining a drop-in, bug-for-bug compatible replacement for `jmoiron/sqlx` (v1). Existing `jmoiron/sqlx` code must compile and run flawlessly without unexpected behavior changes when importing `sqlx-v2`. Any v1 quirks must be replicated exactly.
2. **Correctness & Safety:** We deal with `unsafe.Pointer` for our hot loops. Memory safety, write barriers, alignment guarantees, and bounded slice sizes are paramount. Your modifications to the offset-based inner engine must never introduce out-of-bounds writes, panic conditions, or garbage collection (GC) prematurely.
3. **Performance:** If compatibility and correctness are satisfied, the goal is high throughput and low memory footprint (zero-allocation per row where possible). If a performance optimization increases allocations significantly, it **must** be gated behind an explicit configuration flag and justified. **No PR can introduce a performance regression of >2%** in benchmark functions.

## 2. Modern Go Standards
- **Go Version:** The minimum supported Go version is **1.24+**. You must be comfortable working with and utilizing features like `iter.Seq2`.
- **Generics & Iterators:** Prefer Go 1.24 features (such as `SelectG[T]`, `GetG[T]`, and `SelectIter[T]`) for new APIs, while keeping the classic `interface{}` APIs entirely intact for v1 compatibility.
- **Formatting:** Code must be formatted using `gofmt` and `goimports`. Run `go vet` and strictly avoid unnecessary complexities. 
- **Documentation:** All exported functions must have comprehensive, godoc-compliant comments including a summary line, descriptions for non-obvious parameters, and error behavior.
- **Linting:** Run `golangci-lint run -E modernize ./...` and fix all issues. All code must pass linting checks.

## 3. Architecture & Memory Management
- **Phase 1 (Safe Discovery) vs. Phase 2 (Unsafe Execution):** Maintain the strict boundary between Phase 1 (safe, reflection-based offset calculation cached in `sync.Map`) and Phase 2 (unsafe, O(1) pointer arithmetic in the `rows.Next()` hot loop).
- **GC Safety (`runtime.KeepAlive`):** Ensure `runtime.KeepAlive()` is called appropriately after any scan operation that derives an `unsafe.Pointer` to anchor the memory allocated via `reflect.New()`. 
- **Pointer Traversal (Write Barrier):** Fields reached through `*EmbeddedStruct` (pointer embedding) cannot use flat-offset arithmetic. You must use the write-barrier-safe `AddrByTraversal` path (which uses `reflect.NewAt(ptrType, ptr).Elem().Set(reflect.New(targetType))`) to ensure the Go GC is aware of new nested allocations. Do not bypass the write barrier.
- **Buffers:** Utilize the scan buffer pool (`sync.Pool` of `[]interface{}`) to eliminate per-row heap allocations.

## 4. Testing Requirements
Before declaring a task finished, you must verify the code passes our multi-tier test suite. There are no exceptions.

1. **Tier 1 (Unit):** Run `go test -v ./...`.
2. **Tier 2 (Shadow Integration):** Run `go test -v ./internal/shadow/...` to verify edge cases explicitly against v1 behavior using the in-memory SQLite driver. 
   - *If adding a new edge case or structural shape, it must be added to `internal/shadow/nasty_test.go`.*
3. **Tier 3 (Docker Integration):** Will run automatically under the shadow suite if Docker is present to test against PostgreSQL and MySQL natively.
4. **Race Detector:** Always run `go test -race ./...`. No race conditions are tolerated.
5. **Benchmarks:** If modifying the engine, always benchmark (`go test -bench=. -benchmem -count=5 ./...`) and compare with `benchstat`. Ensure there is no regression. 
6. **Fuzzing:** If making changes to the reflection map, offsets, or memory scanning phase, refer to the Fuzzing suite instructions (`FuzzTypeMap`, `FuzzRowScanBounds`) and test safely (or document their execution).
7. **Coverage & Bugfixes:** Meaningful test coverage must be provided for any new feature. The overall project test coverage should aim to be **80%+**. If you are fixing a bug, you **must first write a failing test** that triggers the bug before writing the fix.

## 5. Git & Commits Invariants
You MUST use the following format for every commit: `<type>(<scope>): <description>`.

### Types
- **feat**: New functionality (e.g., `SelectIter`).
- **fix**: Bug fixes (especially memory safety or GC issues).
- **perf**: Performance improvements that don't change logic (e.g., Task 1 remediation).
- **refactor**: Structural changes that don't fix bugs or add features.
- **chore**: Maintenance, CI, or documentation updates.

### Instruction
Do not commit multiple tasks at once. One logical change per commit. If you fix the `NamedExec` regression, that is a `perf(bind): ...`.

---

> **Note to Agents:** Guess-and-check edits are not acceptable. Read the implementation thoroughly before touching `unsafe` blocks.
