## ADDED Requirements

### Requirement: Internal CLI packages avoid direct process exits

Non-test Go files under `internal/cli/` SHALL NOT call or assign `os.Exit` directly. CLI packages that need command-specific non-zero status codes SHALL return structured errors to the binary entrypoint, and only `cmd/*/main.go` SHALL terminate the process.

#### Scenario: Internal CLI package exit hygiene
- **WHEN** repository quality tests scan non-test Go files under `internal/cli/`
- **THEN** the scan SHALL find zero direct `os.Exit` references

#### Scenario: Extension CLI returns exit-code errors
- **WHEN** an extension command needs to signal exit code 1, 2, or 3
- **THEN** it SHALL return a structured CLI error carrying that code
- **AND** it SHALL NOT terminate the process itself
