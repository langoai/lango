## ADDED Requirements
### Requirement: Interactive top-level mode validation fails before app build
Interactive top-level entrypoints SHALL reject unknown `--mode` values before they construct or start their TUI application instance.

#### Scenario: Workbench mode validation fails before app build
- **WHEN** `lango --mode does-not-exist` reaches the interactive workbench path
- **THEN** it SHALL return an actionable unknown-mode error
- **AND** it SHALL NOT construct the workbench application

#### Scenario: Cockpit mode validation fails before app build
- **WHEN** `lango cockpit --mode does-not-exist` runs
- **THEN** it SHALL return an actionable unknown-mode error
- **AND** it SHALL NOT construct the cockpit application

#### Scenario: Chat mode validation fails before app build
- **WHEN** `lango chat --mode does-not-exist` runs
- **THEN** it SHALL return an actionable unknown-mode error
- **AND** it SHALL NOT construct the chat application
