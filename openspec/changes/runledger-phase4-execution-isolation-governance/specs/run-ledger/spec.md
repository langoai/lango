## MODIFIED Requirements

### Requirement: Workspace Isolation
Production runtime SHALL activate workspace isolation for coding steps once the execution-isolation stage is enabled.

#### Scenario: Runtime isolation active
- **WHEN** `runLedger.workspaceIsolation` is enabled
- **THEN** the app runtime wires `PEVEngine.WithWorkspace(...)`
- **AND** coding-step validators execute inside isolated worktrees rather than the base tree

#### Scenario: Retry-safe repeated validation
- **WHEN** the same step is validated multiple times under isolation
- **THEN** each attempt uses a retry-safe workspace identity
- **AND** previous attempts do not block later ones via reused branch metadata

### Requirement: Tool Governance
The system SHALL expose tools to execution agents according to the active step's `ToolProfile`.

#### Scenario: Coding profile
- **WHEN** the active step uses the `coding` profile
- **THEN** only coding-safe execution tools are available

#### Scenario: Supervisor profile
- **WHEN** the active step uses the `supervisor` profile
- **THEN** only supervisor-safe run inspection/approval tools are available
