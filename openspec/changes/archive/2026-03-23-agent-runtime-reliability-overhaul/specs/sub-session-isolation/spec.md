## ADDED Requirements

### Requirement: Isolation activation is independent of provenance observers
Enabling session isolation for an agent SHALL activate child-session routing even when provenance/session-tree observers are not configured.

#### Scenario: Isolated specialist without provenance still forks child state
- **WHEN** `WithAgentIsolatedAgents()` marks `vault` as isolated and no provenance child-lifecycle observer is installed
- **THEN** `vault` turns SHALL still be routed through a child session
- **AND** child-session merge/discard behavior SHALL remain active

### Requirement: Raw isolated turns never persist to the parent session store
Raw assistant/tool messages authored by isolated agents SHALL NOT be persisted into the parent session store under any runtime configuration. Only merged summaries or compact discard notes may persist in the parent history.

#### Scenario: Successful isolated run persists summary only
- **WHEN** an isolated specialist completes successfully
- **THEN** the parent session store SHALL persist only the merged summary authored as the root/orchestrator agent
- **AND** raw isolated assistant/tool turns SHALL remain absent from persisted parent history

#### Scenario: Failed isolated run persists discard note only
- **WHEN** an isolated specialist run is discarded after failure
- **THEN** the parent session store SHALL persist only a compact discard note
- **AND** raw isolated assistant/tool turns SHALL remain absent from persisted parent history

### Requirement: Discard notes include classified failure reason
When an isolated child session is discarded, the persisted parent note SHALL include the runtime failure classification while continuing to exclude raw child history.

#### Scenario: Loop failure discard note
- **WHEN** an isolated specialist run is discarded because of repeated identical tool calls
- **THEN** the parent discard note SHALL include the classification `loop_detected`
- **AND** SHALL NOT embed raw child messages or tool payloads

#### Scenario: Empty after tool use discard note
- **WHEN** an isolated specialist run is discarded because tool work completed without visible synthesis
- **THEN** the parent discard note SHALL include the classification `empty_after_tool_use`
- **AND** SHALL NOT embed raw child messages or tool payloads
