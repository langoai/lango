## ADDED Requirements

### Requirement: Shared CLI loader helpers remain available
The shared `internal/testutil` package SHALL continue to provide lightweight config-loader and bootstrap-loader helpers for CLI regression tests after the global stdout interception harness is removed.

#### Scenario: CLI config loader tests still compile
- **WHEN** CLI regression tests construct commands that accept `func() (*config.Config, error)` dependencies
- **THEN** `internal/testutil` SHALL provide helper loaders that return a supplied config or error

#### Scenario: CLI bootstrap loader tests still compile
- **WHEN** CLI regression tests construct commands that accept `func() (*bootstrap.Result, error)` dependencies
- **THEN** `internal/testutil` SHALL provide helper loaders that return a supplied bootstrap result or error
- **AND** those helpers SHALL NOT require the removed global stdout interception harness
