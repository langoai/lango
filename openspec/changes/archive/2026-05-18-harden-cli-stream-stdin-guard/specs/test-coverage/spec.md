## MODIFIED Requirements

### Requirement: CLI production stream guards stay executable
Repository-level CLI production stream regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: CLI production stream regressions are rejected
- **WHEN** CLI production code reintroduces raw print calls or forbidden direct standard-stream references including `os.Stdin`, `os.Stdout`, or `os.Stderr`
- **THEN** an executable repository test SHALL fail
