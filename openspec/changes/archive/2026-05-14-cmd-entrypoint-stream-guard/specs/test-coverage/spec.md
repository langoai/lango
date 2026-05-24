## ADDED Requirements

### Requirement: Cmd entrypoint stream guards stay executable
Repository-level top-level entrypoint stream regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Cmd entrypoint stream regressions are rejected
- **WHEN** `cmd/` production code reintroduces raw print calls or forbidden direct standard-stream references
- **THEN** an executable repository test SHALL fail
