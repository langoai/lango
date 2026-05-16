## ADDED Requirements
### Requirement: Auxiliary docs stay aligned with actual output contracts
Auxiliary docs and legacy main specs SHALL not advertise `--json` behavior that the implementation no longer uses or never supported.

#### Scenario: Stale auxiliary output docs are rejected
- **WHEN** an auxiliary doc or legacy spec reintroduces a stale `--json` example or flag description for a command whose current contract differs
- **THEN** an executable repository test SHALL fail
