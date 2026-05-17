# Sync Bootstrap Phase Inventory Copy

## Why

Bootstrap phase inventory copy has drifted from the implementation. `DefaultPhases()` returns 12 phases, but `internal/bootstrap/phases.go` still says 11 phases and `internal/bootstrap/bootstrap.go` documents an older 7-step startup sequence that omits envelope loading, credential acquisition, master-key unwrap/create, migration, and key derivation phases.

This is a production-readiness issue because bootstrap is a security-sensitive startup path. Stale phase copy makes future maintenance and operator-facing diagnostics harder to reason about.

## What Changes

- Update bootstrap source comments to describe the current 12-phase sequence.
- Add executable source-comment coverage that fails if the bootstrap phase inventory comments drift back to stale counts or omit phase names.
- Update the bootstrap-pipeline spec to require implementation copy and tests to stay aligned with the concrete `DefaultPhases()` inventory.

## Impact

- No runtime behavior change.
- Improves maintainability and documentation accuracy for the bootstrap pipeline.
- Adds a cheap regression guard for future phase inventory edits.
