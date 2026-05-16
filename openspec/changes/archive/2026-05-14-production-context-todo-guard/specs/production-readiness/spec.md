## ADDED Requirements

### Requirement: Production Go code avoids placeholder contexts
Production Go code SHALL not contain `context.TODO()` placeholders in non-test source files.

#### Scenario: Production Go files contain no context.TODO calls
- **WHEN** repository quality guard tests scan non-test Go files under `cmd/` and `internal/`
- **THEN** the scan SHALL find zero `context.TODO()` occurrences
