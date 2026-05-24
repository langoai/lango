## Why

Configured remote A2A agents can currently be omitted from the orchestrator with only per-agent warnings inside the loader. Startup should still degrade gracefully, but the caller must receive a load error when any configured remote agent is invalid or fails to build so operators can see that configured remote capacity is missing.

## What Changes

- Make remote A2A loading return successfully loaded agents plus an aggregate error for skipped configured remotes.
- Treat missing `agentCardUrl` as a loader error instead of a silent empty-agent outcome.
- Preserve graceful degradation: app startup continues and builds the local agent tree with any successfully loaded remotes.
- Add regression tests for missing-card and partial-success behavior.
- Update A2A docs to clarify partial remote loading and startup warnings.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `a2a-protocol`: remote agent loading must surface skipped configured remotes as warnings/errors while preserving partial successful loads and startup degradation.

## Impact

- `internal/a2a/remote.go`
- `internal/app/wiring.go` startup warning behavior through existing error handling
- A2A remote loading tests
- `docs/features/a2a-protocol.md`
