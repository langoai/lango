## Context

The March 21 real-session-isolation change correctly turned `SessionIsolation` into runtime routing, but it pushed isolated writes entirely out of the parent session stream. That is incompatible with ADK's same-run contract:

- `runner.Run()` appends non-partial events through `session.Service.AppendEvent`
- the next LLM step rebuilds contents from `ctx.Session().Events()`
- ADK assumes both operations act on the same causal stream

When Lango routes isolated events into child-only storage, ADK's next step cannot observe the tool response it just executed.

## Decisions

### D1. `SessionIsolation` means cross-turn persistence isolation

`SessionIsolation` will no longer mean "the parent cannot see child raw events during the current run".

It will mean:

- the current run may read child raw events through an ephemeral parent overlay
- the persisted parent history for later turns never keeps raw child events

### D2. Parent in-memory overlay, child persistent isolation

For isolated agents:

- write every event to child history
- also append the same message to the parent adapter's in-memory history
- never persist the raw message to the parent store

This gives ADK same-run causal visibility while preserving cross-turn isolation.

### D3. Roll back overlay before final parent outcome

On child completion or discard:

- remove the overlay tail from the parent in-memory history
- then append the final parent outcome

Outcomes:

- success: root-authored summary
- discard/failure: root-authored compact failure note

### D4. Failure notes are deterministic and compact

Discarded isolated runs leave a short deterministic note in the parent history:

- includes agent name
- includes a stable discard reason
- states that raw child history was discarded

This preserves operator visibility without leaking raw tool chatter across turns.

## Non-Goals

- Persisting child raw history to Ent
- Replacing ADK's runner or request processors
- Changing provenance lifecycle event schema

## Risks and Mitigations

- **Risk:** in-memory overlay drifts from the active adapter
  - **Mitigation:** bind the active child to the current parent adapter at fork/append time and roll back to a recorded base history length.
- **Risk:** successful child with no assistant text could accidentally raw-merge
  - **Mitigation:** use a deterministic fallback summary string instead of allowing empty-summary raw merge semantics.
- **Risk:** discard notes become noisy
  - **Mitigation:** keep the format short and only emit notes for explicit discard-reason paths.
