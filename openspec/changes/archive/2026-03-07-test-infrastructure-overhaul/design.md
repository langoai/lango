## Context

The Lango codebase has 869 Go files with ~180K LOC but only 211 test files with inconsistent patterns. Mock types like `mockStore` are duplicated across 6 files, assertions mix raw `if` checks with testify, and no tests use `t.Parallel()` or benchmarks. Three packages (cron, logging, mdparse) have 0% test coverage.

## Goals / Non-Goals

**Goals:**
- Create a shared `internal/testutil/` package with canonical mocks and helpers
- Standardize all test assertions to testify `assert`/`require`
- Add `t.Parallel()` across all safe test functions and subtests
- Achieve test coverage for all zero-coverage packages
- Add benchmarks for performance-critical hot paths

**Non-Goals:**
- Refactoring production code
- Achieving 100% test coverage across all packages
- Replacing local mocks that have specialized behavior (e.g., adk's mockStore)
- Adding integration/e2e test infrastructure

## Decisions

### 1. Shared testutil package over code generation
**Decision**: Hand-written `internal/testutil/` package with canonical mocks.
**Rationale**: Code generation tools (mockgen, counterfeiter) add build complexity and require regeneration on interface changes. Hand-written mocks are simpler, more readable, and sufficient for the project's interface count (~10 key interfaces).

### 2. testify over raw assertions
**Decision**: Standardize on `testify/assert` + `testify/require` for all assertions.
**Rationale**: testify is already a dependency (used in ~48% of tests). It provides better error messages, reduces assertion boilerplate, and the `require`/`assert` split maps naturally to fatal vs non-fatal checks.

### 3. t.Parallel() everywhere except app/ tests
**Decision**: Add `t.Parallel()` to all test functions and subtests, except `internal/app/` which may have shared initialization state.
**Rationale**: Parallel tests reduce total test time and expose race conditions. The app package's test setup involves complex initialization that may not be safe to parallelize.

### 4. Local mocks preserved where specialized
**Decision**: Keep package-local mocks when they have specialized behavior (e.g., `expiredKeys` maps in adk, in-memory DB behavior). Only centralize generic interface implementations.
**Rationale**: Moving specialized mocks to testutil would create unnecessary coupling and reduce test readability.

### 5. Table-driven tests with give/want convention
**Decision**: All table-driven tests use `tests := []struct`, loop var `tt`, fields prefixed `give`/`want`.
**Rationale**: Consistent naming reduces cognitive load when reading tests across packages.

## Risks / Trade-offs

- **[Risk] t.Parallel() may expose latent race conditions** → Run all tests with `-race` flag during verification. Fix any races found.
- **[Risk] Mass assertion changes may alter test semantics** → Each unit covers independent directories with no file overlap. Verify with `go test -race -count=1` after each unit.
- **[Trade-off] Some raw assertions remain (~12%)** → These are in complex test patterns where mechanical conversion is error-prone. Can be addressed incrementally.
