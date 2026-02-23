# Contributing to sqlx-v2

This document establishes the limits and standard test verification procedures for the project footprint.

## Test Validation Properties

### Tier 1 — Unit Testing Configuration

```bash
go test -v ./...
```

Executes core mapping subsystems (`internal/reflectx`, `internal/bind`). Does not require active dependencies.

### Tier 2 — Shadow Validation Properties

```bash
go test -v -msan -race ./internal/shadow/...
```

Validates output parameter generation properties relative to `jmoiron/sqlx` bounds. Configuration edges tested: multi-pointer depth bindings, embedded alias derivations, nullability limits, and block primitives. Runs mapped inside a system SQLite instance.

### Tier 3 — Protocol Matrix Integrations

```bash
go test -v -msan -race ./internal/shadow/...
```

Evaluates identical sequence execution commands against initialized PostgreSQL and MySQL binaries built utilizing Testcontainers. Local Docker binaries are verified programmatically. Runner routines automatically halt verification phases if external dependencies are missing from the configuration path.

Operation verification limits check:

```bash
docker info
```

## Fuzz Diagnostics

Bounds undergo isolated verification loops via the standard runtime fuzzer parameters (`testing.F`). The subsystem maps isolated evaluation binaries subject to memory validation rulesets (`-race` and `-msan`).

Target execution script:

```bash
# Instantiate target definitions
docker build -t sqlx-fuzz -f Dockerfile.fuzz .

# Process sequence blocks (1GB / 2vCPUs defined limits)
docker run --rm --memory="1g" --cpus="2" sqlx-fuzz
```

## System Regression Block Parameters

Execution changes map `benchstat` routines unconditionally:

```bash
go test -bench=. -benchmem -count=5 ./...
```

**System Bound:** Patches defining >2% absolute computational delay across targets recorded within [`BENCHMARK.md`](BENCHMARK.md) must fail. Any specific execution profile parameters replacing absolute latency mapping with space operations mandate physical flags and parameter justification paths via the pull request document.

### Benchmarking Procedures 

Validation maps directly:

```bash
# Evaluate baseline variables
go test -bench=. -benchmem -count=5 ./... > old.txt

# Evaluate candidate bounds
go test -bench=. -benchmem -count=5 ./... > new.txt

benchstat old.txt new.txt
```

## Address Violation Execution Bounds

Data race mappings force failure evaluation definitions across target spaces:

```bash
go test -msan -race ./...
```

## System Requirements

Merge definitions log parameters according to:
- **Baseline Limits:** Coverage parameters mapping equal or greater than 80% coverage limits.
- **Defect Mapping:** Remediation defines negative test paths that map system failure faults prior to code implementation mapping changes.

## Integration Evaluation Matrix

- [ ] All three target evaluations assert exit bound 0 (`go test ./...`).
- [ ] Concurrency address bindings execute zero faults (`go test -msan -race ./...`).
- [ ] Speed boundary outputs map values registering variance within 2% margin thresholds (`benchstat`).
- [ ] Target logic map limit yields bounds greater than 80%.
- [ ] Defect parameters locate regression physical logic.
- [ ] Exposed symbols process exact bound parameters defined in godoc formatting outputs.
- [ ] Shadow integration tests add limits mappings against modifications (`internal/shadow/suite_test.go`).

## Coding Specification

- Structure format verification requires zero output faults for `gofmt` and `go vet` targets.
- Library routing enforces `goimports` ordering map definitions.
- Public method footprints export descriptive behavioral definitions corresponding to expected structural bounds.
