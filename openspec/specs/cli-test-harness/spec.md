## Purpose

Capability spec for cli-test-harness. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Shared CLI test harness
`internal/testutil/loaders.go` and `internal/testutil/helpers.go` SHALL provide reusable CLI test infrastructure: fake config loader, fake boot loader backed by an in-memory Ent client, and shared test helpers that work without real DB or network connections. Command execution helpers MAY remain package-local inside individual CLI `_test.go` files as long as they stay command-stream based.

#### Scenario: Fake config loader returns preset config
- **WHEN** a test uses `testutil.FakeCfgLoader(cfg)`
- **THEN** it returns the given config without touching the filesystem

#### Scenario: Fake boot loader returns in-memory bootstrap result
- **WHEN** a test uses `testutil.FakeBootLoader(t, cfg)`
- **THEN** it returns a bootstrap result backed by an in-memory Ent client
- **AND** it does not require a real on-disk database or network dependency

#### Scenario: CLI command stdout capture
- **WHEN** a CLI package executes a cobra command through a package-local helper that sets `cmd.SetOut`, `cmd.SetErr`, and `cmd.SetArgs`
- **THEN** stdout output is captured and available for assertions

### Requirement: Zero-coverage CLI packages have baseline tests
CLI packages `memory`, `graph`, `learning`, `librarian`, `approval`, and `cron` SHALL each have at least 2 tests (happy path + error path) using the shared harness.

#### Scenario: CLI memory tests pass
- **WHEN** running `go test ./internal/cli/memory/...`
- **THEN** at least 2 tests execute and pass

#### Scenario: CLI graph tests pass
- **WHEN** running `go test ./internal/cli/graph/...`
- **THEN** at least 2 tests execute and pass

#### Scenario: CLI approval tests pass
- **WHEN** running `go test ./internal/cli/approval/...`
- **THEN** at least 2 tests execute and pass

### Requirement: CLI test harness regressions stay command-stream based
CLI regression tests SHALL avoid process-global stdio replacement and legacy shared exec helpers after the command-stream harness migration.

#### Scenario: CLI tests reject process-global stdio replacement
- **WHEN** a CLI `_test.go` file reintroduces `os.Stdin`, `os.Stdout`, or `os.Stderr` reassignment
- **THEN** an executable repository test SHALL fail

#### Scenario: CLI tests reject legacy ExecCmd helpers
- **WHEN** a CLI `_test.go` file reintroduces `testutil.ExecCmd(...)` or `testutil.ExecCmdOK(...)`
- **THEN** an executable repository test SHALL fail
