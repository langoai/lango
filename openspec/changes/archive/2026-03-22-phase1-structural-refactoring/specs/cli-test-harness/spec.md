## ADDED Requirements

### Requirement: Shared CLI test harness
`internal/testutil/cli_harness.go` SHALL provide reusable test infrastructure: fake config loader, in-memory store factory, stdout/stderr capture, and cobra command execution helper. The harness MUST work without real DB or network connections.

#### Scenario: Fake config loader returns preset config
- **WHEN** a test uses `testutil.FakeCfgLoader(cfg)`
- **THEN** it returns the given config without touching the filesystem

#### Scenario: CLI command stdout capture
- **WHEN** a test executes a cobra command via the harness
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
