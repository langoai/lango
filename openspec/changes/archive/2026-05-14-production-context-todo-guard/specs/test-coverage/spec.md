## ADDED Requirements

### Requirement: Production context-placeholder guards stay executable
Repository-level production-code hygiene regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Production context.TODO regressions are rejected
- **WHEN** a non-test Go file under `cmd/` or `internal/` reintroduces `context.TODO()`
- **THEN** an executable repository test SHALL fail
