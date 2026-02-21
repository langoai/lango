## MODIFIED Requirements

### Requirement: Error message format
All error messages in the codebase SHALL follow the concise format `"context: %w"` without "failed to" prefix, per project rule `go-errors.md`.

#### Scenario: Error wrapping without "failed to"
- **WHEN** a function wraps an error
- **THEN** it SHALL use `fmt.Errorf("verb noun: %w", err)` format (e.g., `"create session: %w"` not `"failed to create session: %w"`)

#### Scenario: Batch removal across codebase
- **WHEN** the refactoring is applied
- **THEN** `grep -r "failed to" internal/ --include="*.go" | grep -v _test.go | grep -v ent/` SHALL return zero results
