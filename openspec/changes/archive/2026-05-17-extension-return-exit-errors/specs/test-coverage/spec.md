## ADDED Requirements

### Requirement: Structured CLI exit-code tests

Executable tests SHALL verify that command packages can return structured CLI exit-code errors and that the `lango` entrypoint preserves those codes.

#### Scenario: Main preserves structured CLI exit code
- **WHEN** the root Cobra command returns a structured CLI error carrying exit code 3
- **THEN** `runMain()` SHALL return 3
- **AND** stderr SHALL include the underlying error message exactly once

#### Scenario: Extension commands return structured exit errors
- **WHEN** `lango extension install` or `lango extension remove` exits through user-declined or user-error paths
- **THEN** direct command tests SHALL observe a returned structured CLI error with the documented code
- **AND** the tests SHALL NOT intercept `os.Exit` through panic seams

### Requirement: Internal CLI os.Exit hygiene guard

Executable repository tests SHALL reject direct `os.Exit` usage from non-test Go files under `internal/cli/`.

#### Scenario: Internal CLI os.Exit regressions are rejected
- **WHEN** an `internal/cli/**/*.go` production file reintroduces a direct `os.Exit` reference
- **THEN** a repository-level test SHALL fail and identify the offending file and line
