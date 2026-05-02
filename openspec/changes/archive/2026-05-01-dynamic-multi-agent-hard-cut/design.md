# Design

## Contract Inventory

- `multi-agent-orchestration`: built-in production path still references legacy transfer compatibility.
- `agent-control-plane-tools`: built-in teammate runtime already exists but is not yet the only production path.
- `agent-routing`: embedded prompt files still require `transfer_to_agent("lango-orchestrator")`.
- `agent-registry`: embedded `AGENT.md` defaults remain part of the production prompt contract.
- `adk-architecture`: `failed to find agent` retry still assumes a useful sub-agent list exists.
- `tool-capability-layer`: grant/recheck semantics must be aligned with the hard cut.
- `run-ledger`: durable visibility expectations must be explicit for built-in teammate runs.

## RunLedger Audit Verdict

| Item | Verdict | Notes |
|------|---------|-------|
| teammate spawn submission | `recorded` | `internal/agentrt/control_tools.go` `buildAgentSpawn` submits built-in teammate work with `background.Origin{Channel: "agent_control"}` and `internal/agentrt/agent_run_projection.go` `BackgroundProjection.PrepareTask` preserves the canonical run ID into RunLedger write-through. |
| run status transitions | `recorded` | `internal/agentrt/agent_run_projection.go` `BackgroundProjection.SyncTask` mirrors background state to both `AgentRunProjection.SyncTask` and RunLedger, while `internal/runledger/writethrough.go` `BackgroundWriteThrough.SyncTask` appends durable started/completed/failed events. |
| projection sync markers | `recorded` | `internal/runledger/writethrough.go` `BackgroundWriteThrough.PrepareTaskWithID` and `BackgroundWriteThrough.SyncTask` both append `projection_synced` markers through `appendProjectionSyncEvent`. |
| approval-blocked conditions | `follow-up` | `internal/agentrt/capability_runtime.go` `HandleBlockedToolCall` records `blocked_waiting_approval` only in `AgentRunStore.UpdateProjection`; there is no RunLedger or projection-sync mirror for that operator-visible condition today. |
| recovery states | `follow-up` | `internal/agentrt/control_tools.go` `agentRunResponse` can surface `recovery_state`, but this change found no production writer that persists `AgentRun.RecoveryState`, so there is no durable or projection-synced recovery trail for teammate runs yet. |

## agent_control Propagation Check

| Surface | Result | Notes |
|---------|--------|-------|
| background submission origin | `present` | `internal/agentrt/control_tools.go` `buildAgentSpawn` stamps `agent_control` onto the background origin, and `internal/app/modules.go` wires the same control plane through `agentrt.NewBackgroundProjection(...)`. |
| recovery surface | `partial` | `internal/agentrt/control_tools.go` `agentRunResponse` exposes `recovery_state`, but `internal/cli/cockpit/runtimebridge.go` only forwards `RecoveryDecisionEvent` from the coordinating executor. Built-in teammate run recovery is therefore not projected into the TUI runtime feed yet. |
| approval flow | `partial` | `internal/agentrt/capability_runtime.go` `HandleBlockedToolCall` projects `blocked_waiting_approval`, `blocked_reason`, and `grant_request_id`, and `agent_wait` returns them. The state is truthful in CLI/API responses, but it is not mirrored durably into RunLedger. |
| CLI/TUI runtime views | `present` | `internal/cli/agent/status.go` reports `dynamic-v1` from the minimal `agent.multiAgent && background.enabled` truth condition and prints the user-facing `background.enabled` hint when that path is disabled. `internal/app/modules.go` also keeps the `agent_control` category registration keyed to the automation module that actually exposes it. |
| public docs | `present` | `docs/features/multi-agent.md` already documents that the dynamic built-in teammate path depends on the background submit path and that the operator surface is `agent_spawn` / `agent_wait` / `agent_stop`. The default-config hint is carried by the CLI status text rather than a new doc edit in this pass. |
