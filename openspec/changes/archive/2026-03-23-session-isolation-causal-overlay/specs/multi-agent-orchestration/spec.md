## MODIFIED Requirements

### Requirement: Session Isolation
`SessionIsolation` SHALL be a runtime behavior contract, not metadata only.

#### Scenario: Isolated sub-agent uses same-run overlay
- **WHEN** a sub-agent with `SessionIsolation=true` runs
- **THEN** its raw events SHALL be written to child session history
- **AND** the active parent session's in-memory view SHALL also see those events for the current run
- **AND** the parent persistent history SHALL not store those raw events

#### Scenario: Successful isolated run summary-merges to parent
- **WHEN** an isolated child run completes successfully
- **THEN** any raw in-memory overlay for that child SHALL be removed from the parent view
- **AND** the parent session SHALL retain only a root-authored summary outcome

#### Scenario: Failed isolated run leaves compact failure note
- **WHEN** an isolated child run fails or returns only a rejection/escalation path
- **THEN** any raw in-memory overlay for that child SHALL be removed from the parent view
- **AND** the parent session SHALL retain only a compact root-authored failure note

#### Scenario: Non-isolated sub-agent unchanged
- **WHEN** a sub-agent has `SessionIsolation=false`
- **THEN** it continues to use the existing parent-session execution path
