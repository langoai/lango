## Context

`internal/a2a.LoadRemoteAgents` is invoked during multi-agent startup before `BuildAgentTree()`. The current loader logs skipped remotes internally but can return an empty agent list with a nil error when every configured remote is invalid. The app can then build a local-only tree while the operator believes remote capacity was configured.

## Goals / Non-Goals

**Goals:**

- Preserve startup resilience by allowing partial remote A2A loading.
- Return an aggregate loader error when any configured remote agent is skipped because of invalid config or ADK proxy construction failure.
- Keep successfully loaded remote agents available to the orchestrator even when other configured remotes fail.
- Keep `lango agent list` as a configuration inventory; it should not claim runtime load success unless `--check` is used.

**Non-Goals:**

- Do not make remote A2A failures fatal to app startup.
- Do not change the remote A2A JSON config shape.
- Do not add network probing to startup beyond existing ADK remote construction.

## Decisions

- `LoadRemoteAgents` will return `([]agent.Agent, error)` where the slice contains every successfully constructed remote proxy and the error joins all skipped configured remotes.
- Missing `agentCardUrl` will be treated as a loader error because the configured remote cannot ever be loaded.
- ADK construction failures will remain non-fatal for the whole app but will be included in the aggregate error returned to the caller.
- Tests will use an injectable remote-agent constructor seam so loader behavior can be covered without network-dependent ADK calls.

## Risks / Trade-offs

- Aggregate errors may add a second startup warning because the loader also logs per-remote context. This is acceptable because the caller-level warning makes the degraded startup visible at the orchestration boundary.
- Keeping partial success means operators must still use logs or `doctor`/`agent list --check` to identify exactly which remote failed, but it avoids making one bad remote disable all local agent operation.
