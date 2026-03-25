## Delta: agent-control-plane

### Added Requirements

- **REQ-CP-GUARD-TRANSFER**: The orchestrator direct-tool guard MUST exempt pure `transfer_to_agent` FunctionCall events. "Pure" means ALL FunctionCalls in the event are `transfer_to_agent`; mixed events (transfer + real tool) MUST still be blocked.
  - Rationale: ADK yields the model-response event (with FunctionCall) before promoting it to `Actions.TransferToAgent` in a subsequent event. The guard fires on the first event, killing legitimate delegations.

- **REQ-CP-RECOVERY-GUARD-VIOLATION**: Recovery policy MUST classify `CauseOrchestratorDirectTool` errors as `RecoveryEscalate` (not `RecoveryRetry`). Same-input retry cannot resolve a guard violation.

- **REQ-CP-RECOVERY-DIAGNOSTICS**: Recovery event logging MUST include `error_code` and `cause_class` from `AgentError` when available, to support root-cause analysis.

### Modified Behavior

- `isPureTransferToAgentCall` added to guard condition at `agent.go:331`
- `RecoveryPolicy.Decide` handles `CauseOrchestratorDirectTool` before generic tool-error fallthrough
- `CoordinatingExecutor.runWithRecovery` logs structured diagnostic before recovery action
