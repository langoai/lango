## ADDED Requirements

### Requirement: Config fingerprint in provenance checkpoint
The provenance subsystem SHALL compute a SHA-256 fingerprint of session-relevant configuration state (explicit keys, auto-enabled flags, hooks config) and store it as `config_fingerprint` metadata in session provenance checkpoints.

#### Scenario: Config fingerprint recorded at session start
- **WHEN** the provenance module initializes for a session
- **THEN** a provenance checkpoint SHALL include a `config_fingerprint` metadata field containing a hex-encoded SHA-256 digest of the serialized config state

### Requirement: Hook registry snapshot in provenance checkpoint
The provenance subsystem SHALL capture a JSON snapshot of the current hook registry (pre-hooks and post-hooks with name and priority) and store it as `hook_registry` metadata in session provenance checkpoints.

#### Scenario: Hook snapshot recorded in checkpoint metadata
- **WHEN** a session provenance checkpoint is created
- **THEN** the checkpoint metadata SHALL include a `hook_registry` field containing a JSON array of hook entries, each with `name` and `priority` fields
