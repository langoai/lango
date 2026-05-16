## ADDED Requirements
### Requirement: Auxiliary docs drift guards stay executable
Repository-level docs guards SHALL continue to cover stale output-contract drift in auxiliary docs and older specs, not only primary CLI reference pages.

#### Scenario: Auxiliary stale output references are rejected
- **WHEN** a guarded auxiliary doc or older main spec reintroduces a stale `--json` example or flag entry
- **THEN** an executable repository test SHALL fail
