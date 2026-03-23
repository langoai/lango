## Why

Recent Telegram production failures exposed a systemic reliability gap in the end-to-end agent harness, not a single tool bug. A March 23, 2026 session showed `vault` repeatedly calling `payment_balance {}` and `payment_wallet_info {}` without ever producing a user-visible completion, while the channel layer fell back to a generic empty-response message and the tool-less orchestrator later hallucinated direct execution of `payment_balance`.

The same session also persisted 24 raw `vault` turns in the parent conversation history, which means the current "session isolation" contract is not reliably active in real runtime configurations. Until the runtime can deterministically explain what happened between user input, orchestration, tool use, and final output, we will continue shipping brittle behavior that is hard to diagnose and easy for the model to misrepresent.

## What Changes

- Add a shared turn runtime that owns timeout policy, per-turn tracing, outcome classification, and response finalization across channels, gateway, and automation entrypoints.
- Introduce an append-only turn trace capability that records delegation, tool calls, tool results, retries, loop detection, and final outcome for each user turn.
- Make specialist child-session isolation self-contained so it works even when provenance/session-tree observers are disabled.
- Enforce a runtime completion contract for specialist agents: after tool use they must either produce a visible completion, transfer control, or terminate with a structured incomplete outcome.
- Detect repeated same-agent same-tool same-params loops even when call IDs change, and convert zero-text tool-only terminal states into structured runtime failures instead of silent success.
- Prevent the tool-less orchestrator from hallucinating direct specialist tool execution during recovery; recovery must be evidence-based or re-routed.
- Add transcript replay and integration harness coverage for real-world multi-agent failures before further prompt or routing tweaks.

## Capabilities

### New Capabilities
- `agent-turn-tracing`: Append-only per-turn execution traces with classified outcomes and diagnostic summaries.

### Modified Capabilities
- `multi-agent-orchestration`: Add specialist completion contracts, repeated-call containment, and evidence-only orchestrator recovery.
- `sub-session-isolation`: Guarantee child-session isolation without requiring provenance hooks and prevent raw isolated turns from leaking into parent history.
- `agent-error-handling`: Add dedicated classifications for empty-after-tool-use and repeated-call loops, plus truthful recovery messaging.
- `test-infrastructure`: Add transcript replay fixtures and end-to-end regression coverage for multi-agent runtime failures.

## Impact

- `internal/app/channels.go`, `internal/gateway/server.go`, and automation entrypoints for shared turn execution and response finalization.
- `internal/adk/agent.go`, `internal/adk/session_service.go`, `internal/adk/errors.go`, and related runtime adapters for completion contracts, loop detection, and isolation guarantees.
- `internal/orchestration/*` and prompt/agent registry files for evidence-only recovery and tighter specialist handoff rules.
- New trace persistence, logging, and diagnostic surfaces in the application database and/or internal observability layers.
- `internal/testutil/*`, integration fixtures, and reliability replay tests.
- Downstream CLI/TUI diagnostics, `README.md`, docs, prompts, and skills that describe runtime guarantees and debugging workflows.
