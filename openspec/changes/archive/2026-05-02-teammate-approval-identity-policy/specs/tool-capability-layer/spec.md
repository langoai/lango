## ADDED Requirements

### Requirement: Teammate approval request identity policy
The teammate capability layer SHALL define approval identity using a stable logical `GrantRequestID` per `(runID, toolName)`. Repeated approval blocks for the same logical request SHALL reuse that `GrantRequestID` and SHALL surface renewed-attempt semantics through explicit attempt metadata.

For this slice, the required attempt metadata is:

- `grant_attempt` — integer counter for the current active blocked cycle
- `grant_state` — one of `pending`, `granted`, `denied`, or empty when no approval cycle is currently active

The current slice does **not** introduce a separate long-lived cycle identifier. It only preserves the latest active blocked cycle semantics.

#### Scenario: First approval block initializes stable identity
- **WHEN** a built-in teammate first blocks on a dangerous in-scope tool
- **THEN** the runtime SHALL assign a stable logical `GrantRequestID`
- **AND** the runtime SHALL initialize `grant_attempt = 1`
- **AND** the runtime SHALL expose a pending-style attempt state

#### Scenario: Repeated block for the same run and tool reuses the logical identity
- **WHEN** the same built-in teammate run blocks again on the same tool while it is still in the same active approval-blocked cycle
- **THEN** the runtime SHALL reuse the same logical `GrantRequestID`
- **AND** it SHALL increment separate attempt metadata instead of rotating the logical request ID

#### Scenario: Grant or denial ends the active attempt cycle
- **WHEN** a blocked approval request is granted or denied
- **THEN** the runtime SHALL clear the active blocked projection state
- **AND** it SHALL clear the active-cycle attempt metadata

#### Scenario: Later fresh block starts a new active cycle with the same logical request ID
- **WHEN** the same `(run, tool)` blocks again after a prior grant or denial cleared the active blocked cycle
- **THEN** the runtime SHALL reuse the same logical `GrantRequestID`
- **AND** it SHALL restart `grant_attempt` at `1`
- **AND** that later block SHALL be treated as the latest active cycle of the same logical request rather than as a separately identified historical cycle
