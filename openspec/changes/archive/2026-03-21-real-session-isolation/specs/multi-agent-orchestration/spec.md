## MODIFIED Requirements

### Requirement: Session Isolation
`SessionIsolation` SHALL be a runtime behavior contract, not metadata only.

#### Scenario: Isolated sub-agent avoids parent history pollution
- **WHEN** a sub-agent with `SessionIsolation=true` runs
- **THEN** its raw events are written to child session history rather than directly into the parent session history

#### Scenario: Successful isolated run summary-merges to parent
- **WHEN** an isolated child run completes successfully
- **THEN** the parent session receives only a summary message
- **AND** the full child history is not appended to the parent

#### Scenario: Failed isolated run discarded
- **WHEN** an isolated child run fails or returns only a rejection/escalation path
- **THEN** the child session is discarded
- **AND** no child summary is merged into the parent

#### Scenario: Non-isolated sub-agent unchanged
- **WHEN** a sub-agent has `SessionIsolation=false`
- **THEN** it continues to use the existing parent-session execution path
