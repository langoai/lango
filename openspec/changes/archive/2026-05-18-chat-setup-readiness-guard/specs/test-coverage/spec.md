## ADDED Requirements

### Requirement: Focused chat setup readiness coverage stays executable
Repository-level regressions in focused chat first-run readiness behavior SHALL be enforced by executable tests.

#### Scenario: Focused chat setup readiness coverage blocks regressions
- **WHEN** focused chat is constructed with an incomplete default config
- **THEN** executable chat tests SHALL fail if the shell renders a ready/send state
- **AND** executable chat tests SHALL fail if normal input reaches the turn runner before setup is ready
- **AND** executable chat tests SHALL fail if slash commands are unavailable before setup is ready
