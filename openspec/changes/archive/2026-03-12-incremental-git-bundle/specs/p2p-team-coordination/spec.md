## ADDED Requirements

### Requirement: Health monitor git state provider configuration
The HealthMonitorConfig SHALL accept optional GitStateProvider and WorkspaceIDsFn fields for enabling git state tracking.

#### Scenario: Configure git state tracking
- **WHEN** HealthMonitor is created with GitStateProvider and WorkspaceIDsFn
- **THEN** the monitor stores both functions and uses them during health pings

#### Scenario: Git state tracking disabled by default
- **WHEN** HealthMonitor is created without GitStateProvider
- **THEN** git state collection is skipped during health pings

### Requirement: Workspace git divergence event
The system SHALL publish WorkspaceGitDivergenceEvent via eventbus when divergence is detected.

#### Scenario: Divergence event published
- **WHEN** DetectDivergence finds members with different HEADs
- **THEN** a WorkspaceGitDivergenceEvent is published with WorkspaceID, MajorityHead, and list of divergent members
