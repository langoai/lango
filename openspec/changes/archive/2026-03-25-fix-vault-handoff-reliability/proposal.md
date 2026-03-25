## Why

Structured multi-agent turns can leave behind interrupted specialist tool-call state when a streaming run fails before cleanup. In structured orchestration mode, the recovery wrapper can then retry the same input without steering away from the failed specialist, which repeats the failure and surfaces OpenAI orphan-repair errors instead of a stable handoff.

## What Changes

- Ensure streaming agent failures discard isolated child sessions the same way collection-based failures already do.
- Add runtime cleanup that closes dangling parent-visible tool calls after failed turns so retries and later turns do not inherit orphaned specialist tool state.
- Make structured recovery agent-aware so specialist tool failures reroute away from the failed specialist instead of blindly retrying the same input.
- Record structured recovery attempts in turn traces and make trace persistence resilient for long turns by using a fresh detached timeout per write.
- Update operator-facing documentation to describe reroute-aware recovery and trace-backed diagnostics for structured multi-agent mode.

## Capabilities

### New Capabilities
- `runtime-failed-turn-cleanup`: Close dangling parent-visible tool calls after failed turns without persisting raw isolated child history.

### Modified Capabilities
- `multi-agent-orchestration`: Structured recovery must carry failed specialist identity and avoid repeating the same specialist after specialist tool failures.
- `sub-session-isolation`: Streaming failure paths must discard isolated child sessions consistently and keep raw child history out of parent persistence.
- `agent-error-handling`: Recovery policy must distinguish pre-specialist retries from post-specialist reroute recovery.
- `agent-turn-tracing`: Trace writes must use per-write detached timeouts and record structured recovery attempts.

## Impact

- Affected code: `internal/adk`, `internal/agentrt`, `internal/turnrunner`, `internal/provider/openai`, and session/runtime helpers under `internal/session`.
- Affected docs: multi-agent runtime documentation in `README.md` and `docs/features/multi-agent.md`.
- No public API or CLI surface changes are expected.
