## ADDED Requirements
### Requirement: Architecture doc broken-path guards stay executable
Repository-level regressions that reintroduce known-broken implementation paths into public architecture docs SHALL be enforced by an executable test.

#### Scenario: Stale librarian proactive buffer path is rejected
- **WHEN** the architecture data-flow page reintroduces `internal/librarian/buffer.go`
- **THEN** an executable repository test SHALL fail
