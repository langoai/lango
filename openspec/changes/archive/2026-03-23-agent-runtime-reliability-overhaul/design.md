## Context

The current multi-agent stack has several local safeguards, but they do not compose into a truthful end-to-end harness:

- `internal/app/channels.go` and `internal/gateway/server.go` each finalize turns independently and only understand "text or generic empty fallback".
- `internal/adk/session_service.go` only creates a child-session store when `WithChildLifecycleHook()` is enabled. In practice that ties specialist isolation to provenance/session-tree wiring instead of to `WithAgentIsolatedAgents()`.
- `internal/adk/agent.go` has turn-limit and same-tool churn defenses, but a real Telegram session on March 23, 2026 still persisted repeated `vault` calls to `payment_balance {}` with zero visible completion.
- The parent session database for `telegram:1496773987:1496773987` contains 24 raw `vault` turns, including repeated zero-length assistant turns and tool outputs, which contradicts the intended child-session isolation model.
- The same failure sequence later allowed the tool-less orchestrator to emit a direct `payment_balance` tool attempt, proving that recovery is still prompt-driven rather than runtime-enforced.

Competitor harnesses reinforce the same lesson. The public `openai/codex` repository centralizes tool execution, approval, and sandbox policy in shared core modules (`codex-rs/core/src/tools/orchestrator.rs`, `codex-rs/core/src/exec_policy.rs`) and routes multiple frontends through the same rollout/runtime layer instead of duplicating end-of-turn behavior per surface. The public `anthropics/claude-code` repository does not expose its full harness engine, but it does expose explicit lifecycle hooks and plugins as first-class extension points, which is still a stronger operational contract than relying on post-hoc generic fallbacks.

## Goals / Non-Goals

**Goals:**
- Make every user turn end with a deterministic, machine-readable outcome.
- Guarantee specialist session isolation regardless of provenance wiring.
- Stop repeated same-call specialist loops before they consume the full request timeout.
- Replace "silent success + empty fallback" with truthful, classified failures tied to trace evidence.
- Add replayable regression fixtures for the exact failure classes we have already seen in production.

**Non-Goals:**
- Replacing ADK with a fully custom runner in this change.
- Introducing new specialist agent roles.
- Solving all prompt-quality issues through prompt rewriting alone.
- Building a generalized observability platform for every subsystem beyond agent-turn reliability.

## Decisions

### D1. Introduce a shared `TurnRunner` as the only entrypoint into agent execution

Channels, gateway, and automation paths will stop owning their own timeout/fallback/finalization logic. A shared `TurnRunner` will:

- create a `turn_id` and `trace_id`
- resolve idle/hard deadlines
- invoke `RunAndCollect` / streaming collection
- classify the final outcome
- return a single `TurnResult` to the caller

This moves the product contract from "whatever each caller does after agent completion" to a single runtime contract.

**Rationale:** The current duplication between `channels.go` and `gateway/server.go` guarantees behavioral drift. A shared core is also the strongest lesson from Codex's public architecture.

**Alternatives considered:**
- Keep the current callers and share only helper functions. Rejected because it preserves multiple finalization sites and weak ownership of trace/outcome semantics.
- Push outcome handling down into each channel adapter. Rejected because reliability rules belong to runtime, not transport.

### D2. Make child-session isolation independent of provenance observers

`WithAgentIsolatedAgents()` will instantiate the child-session routing machinery on its own. Provenance/session-tree hooks become optional observers of the lifecycle, not prerequisites for the lifecycle.

**Rationale:** Isolation is a correctness guarantee, not an observability feature. The current coupling means specialist isolation silently disappears in real runtimes where provenance is not fully wired.

**Alternatives considered:**
- Require provenance everywhere. Rejected because it turns a correctness guarantee into a deployment/configuration trap.
- Leave isolation best-effort and document the limitation. Rejected because production evidence already shows this leaks specialist raw turns into user-facing history.

### D3. Enforce a runtime completion contract for specialist agents

After a specialist uses tools, the runtime will not treat "no visible text" as success. A specialist turn must end in exactly one of:

- visible assistant completion
- `transfer_to_agent` handoff
- structured incomplete outcome emitted by the runtime (`empty_after_tool_use`, `loop_detected`, `timeout_after_tool_use`)

The shared collector will watch for tool-result-only terminal states and convert them into structured failures before channel/gateway fallback logic runs.

**Rationale:** Prompt instructions alone are insufficient. The runtime must own the contract that tool work without completion is an error, not a successful empty turn.

**Alternatives considered:**
- Only strengthen prompts. Rejected because we already have production evidence that prompts do not prevent tool-only loops.
- Treat the last tool output as the user answer automatically. Rejected because raw tool JSON is not always safe or complete user-facing output.

### D4. Replace same-tool churn with call-signature containment

Loop detection will key on `(agent_name, tool_name, canonical_params_json)` and optionally use repeated output hashes as an accelerator. Changing only the tool call ID will not reset the loop counter.

Thresholds:
- repeated identical call signatures within the same specialist turn trigger loop containment
- repeated identical outputs without visible progress can also trigger incomplete classification

**Rationale:** The observed `payment_balance {}` loop changed call IDs but not behavior. Tool name alone is too weak; visible text alone is too late.

**Alternatives considered:**
- Count only tool names. Rejected because alternating helper calls or changed call IDs still bypass useful containment.
- Count only total turn budget. Rejected because that detects failure far too late and provides poor root-cause information.

### D5. Make orchestrator recovery evidence-only

The orchestrator will never attempt specialist tool calls directly when it has no tools. On specialist failure or incomplete outcome, it may:

- route to a different agent
- answer from evidence already collected in the turn trace
- report an honest inability to complete

It may not emit new specialist FunctionCalls itself.

**Rationale:** The March 23 session proves that prompt-level "you have no tools" is not enough. Recovery rules need runtime enforcement.

**Alternatives considered:**
- Keep recovery prompt-only. Rejected because the current system already violated that rule in production.
- Force orchestrator to always retry the same specialist once. Rejected because this amplifies loops instead of containing them.

### D6. Add transcript replay fixtures as a release gate

We will capture real failure transcripts as sanitized fixtures and assert:

- no raw isolated specialist turns persist to the parent session
- repeated same-call loops become `loop_detected`
- empty-after-tool-use becomes a structured failure, not silent success
- channel and gateway paths report the same classified outcome
- orchestrator recovery never emits unavailable direct specialist tool calls

**Rationale:** The current tests validate isolated logic fragments, but not the live end-to-end failure shapes we actually ship.

**Alternatives considered:**
- Keep unit-only coverage. Rejected because the current regressions already slipped through that model.
- Depend on manual Telegram/Gateway verification. Rejected because this is too slow and too easy to miss.

## Risks / Trade-offs

- [Additional trace persistence increases DB writes] → Mitigation: keep traces append-only, compact, and scoped to per-turn events; summarize aggressively after completion.
- [Stricter completion contracts may surface more explicit errors at first] → Mitigation: this is intentional; truthful failures are preferable to hallucinated success.
- [Isolation decoupling may change historical session behavior] → Mitigation: retain parent-summary merge/discard notes so cross-turn continuity remains readable.
- [Replay fixtures can become brittle if they overfit prompt text] → Mitigation: assert on outcomes, event classes, and persistence invariants rather than exact prose.

## Migration Plan

1. Introduce `TurnRunner` and `TurnTrace` behind existing channel/gateway interfaces without changing external APIs.
2. Decouple child-session store activation from provenance hooks and add parent-store persistence guards for isolated agents.
3. Add structured incomplete outcomes and call-signature loop containment to the ADK collection layer.
4. Enforce evidence-only orchestrator recovery and remove direct specialist-tool recovery paths.
5. Add replay fixtures, integration tests, and diagnostics/doc updates.
6. Verify with `go build ./...`, `go test ./...`, OpenSpec verify/sync/archive, and targeted regression replays.

## Open Questions

- Should turn traces reuse existing RunLedger journal primitives, or is a dedicated lightweight turn-trace store the safer first implementation?
- Should the first operator-facing diagnostic surface be `lango doctor`, a TUI screen, or both?
- How long should detailed per-turn traces be retained before compaction/archival?
