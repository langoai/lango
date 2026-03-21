## Context

The current implementation records child-session provenance, but it still writes sub-agent events into the parent session history and uses a synthetic lifecycle model. That is sufficient for provenance scaffolding, but not for actual session isolation semantics.

Within the current ADK integration, the least risky way to honor `SessionIsolation` is to separate **storage behavior** for isolated agents:

- isolated agent events are stored in a child session object
- the parent session only receives a summary on successful completion
- rejected or failed isolated runs are discarded

This keeps the ADK runner integration intact while making the session model meaningfully different for isolated agents.

## Goals / Non-Goals

**Goals**
- Prevent isolated sub-agent events from directly polluting parent history
- Keep isolated child history accessible until merge/discard
- Merge isolated child runs back to parent with a summary authored as the root/orchestrator agent
- Honor the `SessionIsolation` flag from both fallback specs and embedded/default AGENT.md specs

**Non-Goals**
- Replacing the ADK runner with a custom orchestration engine
- Making every sub-agent isolated by default
- Persisting child session history in Ent as a full session store in this change

## Decisions

### D1. Isolation is storage/runtime routing, not a new runner

This change keeps the existing runner and agent tree. The runtime effect of `SessionIsolation` is implemented in the session service by routing isolated agent events into child sessions instead of the parent session.

### D2. Summary merge is the only merge strategy

Successful isolated child runs merge back to the parent with a summary derived from the child history. Full child history is never appended to the parent.

### D3. Root-authored merge summaries

Merged summary messages are authored by the root/orchestrator agent rather than the child agent. This avoids making the next turn resume from the last child agent by accident.

### D4. Built-in specialists are isolated

Built-in specialist agents that perform bounded delegated work (`operator`, `navigator`, `vault`, `librarian`, `automator`, and embedded `chronicler`) are marked `session_isolation: true`. Planner remains non-isolated.

## Risks / Trade-offs

- **ADK still sees parent session during the same invocation**: this change improves actual stored history separation without replacing ADK’s overall control flow.
- **Child history is in-memory**: child sessions remain runtime-scoped in this change. Provenance persistence still captures lifecycle and merged summaries.
- **Behavior change for built-in specialists**: more delegation paths now summarize into parent history instead of appending raw turns, so regression coverage must focus on multi-agent continuity.
