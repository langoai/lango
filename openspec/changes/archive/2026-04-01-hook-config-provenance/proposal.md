## Why

Session reproducibility requires knowing the exact configuration and hook setup active when a session started. Currently, checkpoints record journal state but not the config fingerprint or hook registry snapshot, making it impossible to verify whether a session replay uses the same system configuration.

## What Changes

- Add `CreateManualWithMetadata` method to `CheckpointService` allowing metadata-enriched checkpoints where `runID` is optional (session-init checkpoints may not have a run yet)
- Modify internal `create()` to accept and store a `metadata` map
- Compute a config fingerprint (SHA-256 of ExplicitKeys + AutoEnabled + HooksConfig) at app init and cache it
- Capture hook registry snapshot (pre/post hooks with name:priority) at app init
- Extend `rootSessionObserver` to create a `session_config_snapshot` checkpoint on first session access per app instance, using `sync.Map` for idempotency

## Capabilities

### New Capabilities
- `config-checkpoint`: Session-scoped config fingerprint and hook registry snapshot captured as a provenance checkpoint at session start

### Modified Capabilities
- `session-provenance`: Checkpoint creation now supports optional metadata and optional runID via `CreateManualWithMetadata`

## Impact

- `internal/provenance/checkpoint.go` — new public method + internal `create()` signature change
- `internal/app/modules_provenance.go` — config fingerprint computation and caching
- `internal/app/wiring_provenance.go` — extended rootSessionObserver with idempotent checkpoint creation
- No CLI changes, no API changes, no breaking changes
