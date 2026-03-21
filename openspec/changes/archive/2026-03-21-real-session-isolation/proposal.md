## Why

`SessionIsolation` exists in agent metadata, but runtime execution still keeps sub-agent work in the parent session and only emits synthetic provenance lifecycle events. That means the flag is not a real behavior contract yet: parent history is still polluted by sub-agent turns and isolated child history is not retained as a first-class runtime object.

## What Changes

- Make `SessionIsolation=true` alter runtime behavior instead of being metadata only
- Route isolated sub-agent events into child session history rather than the parent session history
- Merge isolated child sessions back to parent using summary merge, and discard them on failed/rejected runs
- Keep `SessionIsolation=false` agents on the existing parent-session path

## Capabilities

### Modified Capabilities

- `multi-agent-orchestration`: honor `SessionIsolation` at runtime
- `session-provenance`: persist child lifecycle from actual isolated runtime flow rather than generic author heuristics

## Impact

- `internal/adk/`: isolated child-session routing and merge/discard policy
- `internal/session/`: concrete child-store helper for summary merge with parent-author attribution
- `internal/agentregistry/defaults/` and fallback orchestration specs updated to mark built-in specialist agents as isolated
- docs/specs updated so `session_isolation` is treated as a runtime contract
