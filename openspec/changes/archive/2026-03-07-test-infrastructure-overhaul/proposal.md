## Why

The Lango project has grown to 869 Go files / ~180K LOC but test infrastructure has not kept pace: zero `t.Parallel()`, zero benchmarks, zero `TestMain`, mock duplication across 6+ files, inconsistent assertions (52% raw `if` vs 48% testify), and 3 packages with 0% coverage. This creates a foundation of shared test utilities, standardizes patterns, and fills critical coverage gaps.

## What Changes

- Create `internal/testutil/` package with shared helpers (`NopLogger`, `TestEntClient`, `SkipShort`) and canonical mock implementations (`MockSessionStore`, `MockProvider`, `MockEmbeddingProvider`, `MockGraphStore`, `MockCryptoProvider`, `MockTextGenerator`, `MockAgentRunner`, `MockChannelSender`, `MockCronStore`)
- Convert all ~211 test files from raw `if`/`t.Errorf` assertions to testify `assert`/`require`
- Add `t.Parallel()` to ~180+ test files and subtests for faster test execution
- Add comprehensive tests for 3 zero-coverage packages: `cron`, `logging`, `mdparse`
- Add new test coverage for `config`, `mcp`, and `app` packages
- Add 23+ benchmark functions across 6 hot-path packages (`types`, `memory`, `prompt`, `graph`, `asyncbuf`, `embedding`)

## Capabilities

### New Capabilities
- `test-infrastructure`: Shared test utilities, mock implementations, helpers, and conventions for the entire test suite

### Modified Capabilities

## Impact

- All `internal/**/*_test.go` files (~230 files after changes)
- New `internal/testutil/` package (9 files)
- New benchmark files in 6 packages
- Test execution time may decrease due to `t.Parallel()` adoption
- No production code changes — all changes are to test files and new test utility packages
