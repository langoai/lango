## Purpose

Shared test utilities, mock implementations, assertion standards, and conventions for the Lango test suite. Provides a foundation of reusable test infrastructure to eliminate duplication, improve consistency, and enable parallel test execution across all packages.

## Requirements

### Requirement: Shared test helper package
The system SHALL provide an `internal/testutil/` package with shared test utilities including `NopLogger()`, `TestEntClient(t)`, and `SkipShort(t)` helper functions.

#### Scenario: NopLogger returns usable logger
- **WHEN** a test calls `testutil.NopLogger()`
- **THEN** the returned `*zap.SugaredLogger` SHALL be non-nil and SHALL not panic on log calls

#### Scenario: TestEntClient returns functional client
- **WHEN** a test calls `testutil.TestEntClient(t)`
- **THEN** the returned `*ent.Client` SHALL be backed by an in-memory SQLite database with auto-migration
- **THEN** the client SHALL be automatically closed when the test completes via `t.Cleanup()`

#### Scenario: SkipShort skips in short mode
- **WHEN** a test calls `testutil.SkipShort(t)` and the test is run with `-short` flag
- **THEN** the test SHALL be skipped

### Requirement: Canonical mock implementations
The system SHALL provide thread-safe mock implementations for core interfaces: `session.Store`, `provider.Provider`, `embedding.EmbeddingProvider`, `graph.Store`, `security.CryptoProvider`, `cron.Store`, and utility types `TextGenerator`, `AgentRunner`, `ChannelSender`.

#### Scenario: Mocks are thread-safe
- **WHEN** a mock is accessed concurrently from parallel subtests
- **THEN** no data races SHALL occur (verified by `-race` flag)

#### Scenario: Mocks support error injection
- **WHEN** a test sets an error field on a mock (e.g., `mock.CreateErr = errors.New("fail")`)
- **THEN** the corresponding method SHALL return that error

#### Scenario: Mocks support call inspection
- **WHEN** a test calls inspection methods (e.g., `mock.CreateCalls()`)
- **THEN** the mock SHALL return the accurate count of method invocations

#### Scenario: Compile-time interface verification
- **WHEN** the testutil package is compiled
- **THEN** each mock SHALL have a compile-time interface check (e.g., `var _ session.Store = (*MockSessionStore)(nil)`)

### Requirement: Testify assertion standardization
All test files SHALL use `testify/assert` for non-fatal assertions and `testify/require` for fatal assertions. Raw `if`/`t.Errorf`/`t.Fatalf` patterns SHALL be converted.

#### Scenario: Fatal error checks use require
- **WHEN** a test checks an error that would prevent the test from continuing
- **THEN** it SHALL use `require.NoError(t, err)` instead of `if err != nil { t.Fatalf(...) }`

#### Scenario: Non-fatal checks use assert
- **WHEN** a test checks a value that does not prevent continuation
- **THEN** it SHALL use `assert.Equal(t, want, got)` instead of `if got != want { t.Errorf(...) }`

### Requirement: Parallel test execution
All test functions and subtests SHALL include `t.Parallel()` at their top, except tests in `internal/app/` which may depend on shared initialization state.

#### Scenario: Top-level test parallelism
- **WHEN** a test function is defined outside of `internal/app/`
- **THEN** it SHALL call `t.Parallel()` as its first statement

#### Scenario: Subtest parallelism
- **WHEN** a `t.Run()` subtest is defined outside of `internal/app/`
- **THEN** it SHALL call `t.Parallel()` as its first statement inside the closure

### Requirement: Zero-coverage package tests
The system SHALL provide test files for packages with 0% coverage: `cron`, `logging`, and `mdparse`.

#### Scenario: Cron package test coverage
- **WHEN** tests are run for `internal/cron/`
- **THEN** coverage SHALL be at least 70% covering scheduler lifecycle, executor, and delivery

#### Scenario: Logging package test coverage
- **WHEN** tests are run for `internal/logging/`
- **THEN** coverage SHALL be at least 80% covering logger creation and level configuration

#### Scenario: Mdparse package test coverage
- **WHEN** tests are run for `internal/mdparse/`
- **THEN** coverage SHALL be at least 90% covering frontmatter parsing edge cases

### Requirement: Performance benchmarks
The system SHALL provide benchmark functions with `b.ReportAllocs()` for hot-path code in types, memory, prompt, graph, asyncbuf, and embedding packages.

#### Scenario: Benchmark functions exist
- **WHEN** benchmarks are run with `go test -bench=.`
- **THEN** at least 15 benchmark functions SHALL execute across the 6 packages

#### Scenario: Benchmarks report allocations
- **WHEN** a benchmark function runs
- **THEN** it SHALL call `b.ReportAllocs()` to report memory allocation statistics

### Requirement: Transcript replay fixtures for multi-agent runtime failures
The test infrastructure SHALL provide sanitized transcript replay fixtures for real multi-agent runtime failures so that end-to-end harness regressions can be reproduced without live network or external model dependencies.

#### Scenario: Vault balance loop fixture reproduces the failure shape
- **WHEN** the transcript replay harness runs the sanitized vault-balance-loop fixture
- **THEN** the test SHALL reproduce repeated same-signature specialist tool calls and a missing visible completion
- **AND** the runtime assertions SHALL evaluate the resulting classified outcome

#### Scenario: Replay fixture avoids external dependencies
- **WHEN** transcript replay tests execute in CI
- **THEN** they SHALL run without live Telegram, RPC, or external LLM access
- **AND** they SHALL rely only on local fixtures and test doubles

### Requirement: End-to-end assertions cover isolation and outcome parity
Replay-driven integration tests SHALL assert both persistence invariants and user-facing outcome parity across channel and gateway entrypoints.

#### Scenario: Isolated raw turns do not leak into parent history
- **WHEN** a replay fixture exercises an isolated specialist loop
- **THEN** the resulting persisted parent history SHALL contain only summary/discard entries
- **AND** raw specialist assistant/tool turns SHALL remain absent

#### Scenario: Channel and gateway classify the same failure identically
- **WHEN** the same replay fixture is executed through channel-style and gateway-style turn runners
- **THEN** both paths SHALL report the same terminal classification (for example `loop_detected` or `empty_after_tool_use`)
- **AND** both paths SHALL reference the same trace-backed root-cause summary semantics
