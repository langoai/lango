## ADDED Requirements

### Requirement: Config Fingerprint Computation
The system SHALL compute a SHA-256 fingerprint at app initialization from the JSON serialization of ExplicitKeys, AutoEnabled, and HooksConfig. The fingerprint SHALL be cached for the app's lifetime as a hex string.

#### Scenario: Fingerprint computed at init
- **WHEN** the provenance module initializes
- **THEN** a SHA-256 hex string is computed from JSON(ExplicitKeys) + JSON(AutoEnabled) + JSON(HooksConfig)
- **AND** the fingerprint is cached for the app instance lifetime

#### Scenario: Deterministic fingerprint
- **WHEN** the same config is loaded twice
- **THEN** the fingerprint is identical

### Requirement: Hook Registry Snapshot
The system SHALL capture a snapshot of the hook registry at app initialization as a JSON array of `{name, priority}` objects for both pre-hooks and post-hooks. The snapshot SHALL be cached for the app's lifetime.

#### Scenario: Hook snapshot captures all registered hooks
- **WHEN** the hook registry has pre-hooks and post-hooks registered
- **THEN** the snapshot includes all hooks with their name and priority

#### Scenario: Hook snapshot format
- **WHEN** a hook snapshot is serialized
- **THEN** it produces a JSON array of objects with "name" and "priority" fields

### Requirement: Session Config Checkpoint
The system SHALL create a `session_config_snapshot` checkpoint on the first access (Create or Get) of each root session per app instance. The checkpoint SHALL contain metadata keys `config_fingerprint` (hex SHA-256) and `hook_registry` (JSON array).

#### Scenario: Config checkpoint created on first session access
- **WHEN** a root session is accessed for the first time in an app instance
- **THEN** a checkpoint with label `session_config_snapshot` and trigger `manual` is created
- **AND** metadata contains `config_fingerprint` and `hook_registry` keys

#### Scenario: Idempotent checkpoint creation
- **WHEN** the same session is accessed multiple times (Create then Get, or multiple Gets)
- **THEN** only one `session_config_snapshot` checkpoint is created per app instance

#### Scenario: No run ID required
- **WHEN** a session config checkpoint is created
- **THEN** the checkpoint's runID MAY be empty (session-init checkpoints precede runs)
