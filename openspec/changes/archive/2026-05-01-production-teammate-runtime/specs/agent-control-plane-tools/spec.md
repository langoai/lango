## MODIFIED Requirements

### Requirement: agent_spawn tool creates AgentRun with enriched prompt and advisory routing

The existing `agent_spawn` response shape and basic ID semantics remain preserved unless explicitly changed by this requirement: it still creates an `AgentRun`, still returns the spawned run identifier, and still uses the pre-registered `AgentRun.ID` as the canonical control-plane ID. In addition, `agent_spawn` SHALL accept optional `spawn_reason` and `allowed_tools` fields alongside the existing instruction and advisory routing parameters. For built-in teammate types, `allowed_tools` SHALL be validated against the teammate role max scope before execution begins. The runtime SHALL store `spawn_reason` in the projected `AgentRun` state and, when a submitter is configured, SHALL submit the teammate prompt through the existing background manager using the pre-registered `AgentRun.ID`.

#### Scenario: Spawn reason is stored in projection
- **WHEN** `agent_spawn` is called with `spawn_reason: "need file-only helper for manifest audit"`
- **THEN** the created `AgentRun` projection SHALL persist that spawn reason for later inspection

#### Scenario: Allowed tools outside role scope are rejected
- **WHEN** a built-in teammate is spawned with an `allowed_tools` entry outside its role max scope
- **THEN** `agent_spawn` SHALL fail before background submission
- **AND** no `AgentRun` SHALL transition into execution

#### Scenario: Spawn submits through existing in-process execution path
- **WHEN** a submitter and projection are configured for spawned teammates
- **THEN** the runtime SHALL register the pending `AgentRun.ID`
- **AND** SHALL submit the teammate prompt through the existing background manager
- **AND** SHALL preserve parent session linkage and the dynamic allowlist in the child execution context

### Requirement: agent_wait polls AgentRunStore until terminal status

The existing `agent_wait` polling contract remains preserved unless explicitly changed by this requirement: it still polls the `AgentRunStore`, still waits for terminal state or timeout, and still treats timeout as a non-terminal observation result. In addition, `agent_wait` SHALL include projected condition fields in timeout responses for non-terminal runs. A timeout on `blocked_waiting_approval` SHALL remain non-terminal and SHALL return the projected condition instead of coercing the run into failure.

#### Scenario: Timeout returns projected condition fields
- **GIVEN** an `AgentRun` that remains non-terminal at timeout
- **WHEN** `agent_wait` returns with `timeout: true`
- **THEN** the response SHALL include the current projected condition fields needed to understand why the run is waiting

#### Scenario: Approval wait timeout remains non-terminal
- **GIVEN** an `AgentRun` in projected condition `blocked_waiting_approval`
- **WHEN** `agent_wait` times out
- **THEN** the response SHALL keep the run non-terminal
- **AND** SHALL report that approval is still pending

### Requirement: AgentRunProjection implements background.Projection for ID unification

`AgentRunProjection` SHALL continue to unify control-plane and background IDs, and it SHALL additionally project spawn reason, teammate type, dynamic allowlist state, and current wait condition into the control-plane snapshot returned by `agent_wait`.

#### Scenario: Projected state includes spawn metadata
- **WHEN** a teammate run has stored spawn reason and teammate type
- **THEN** `AgentRunProjection` SHALL mirror those fields into the control-plane snapshot

#### Scenario: Projected state reflects approval-blocked condition
- **WHEN** background execution is paused on a capability approval decision
- **THEN** the projection SHALL expose that non-terminal condition to the waiting caller
