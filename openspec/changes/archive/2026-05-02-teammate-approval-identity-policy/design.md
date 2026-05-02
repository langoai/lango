# Design

## Problem

`GrantRequestID` currently comes from `grant-{runID}-{toolName}`. That is enough for the current runtime to clear blocked state, but it is not yet a complete approval identity policy. The missing contract is not just string shape. It is the behavioral meaning of the identity across:

- `CapabilityPolicy.Evaluate(...)`
- `CapabilityRuntime.HandleBlockedToolCall(...)`
- `CapabilityRuntime.ApplyGrant(...)`
- `agent_wait` / operator-facing blocked state
- RunLedger durable mirror

## Decision Surface

This change must decide four things together:

1. **Identity shape**
   - stable per `(run, tool)`
   - or attempt-scoped per `(run, tool, attempt)`

2. **Denial semantics**
   - does denial permanently suppress the same request identity?
   - or can the same request identity be re-issued later?

3. **Retry semantics**
   - when the same tool blocks again after denial or timeout, is that the same approval request or a new attempt?

4. **Surface semantics**
   - what `agent_wait`, runtime projection, and durable mirror must show for repeated requests

## Initial Direction

The chosen direction is:

- `GrantRequestID` remains a **stable logical request identity**
- repeated blocks for the same `(run, tool)` reuse that logical identity
- retry or re-request semantics, when exposed, are represented through separate **attempt metadata** rather than by changing the logical request ID

This keeps runtime matching and blocked-state clearing simple while still leaving room for richer operator-visible retry history.

## Identity Model

### Stable Logical Identity

For a built-in teammate approval-blocked request, the runtime uses one logical identity per `(runID, toolName)`.

That means:

- the same run blocking again on the same tool reuses the same `GrantRequestID`
- the runtime does not treat each repeated block as a brand-new logical request by default
- live and durable surfaces can interpret later blocked states as retries or renewed attempts of the same logical request

### Attempt Metadata

If the system needs to expose retry history, denial history, or renewed approval attempts, it should do so with separate metadata rather than by rotating the logical request ID.

Representative fields may include:

- `grant_attempt`
- `last_attempt_at`
- `last_denied_at`
- `grant_state`

Those fields are not all committed by this design. The point of the policy is that they are **attempt-level metadata**, not replacements for the stable logical request identity.

## Policy Consequences

### Denial

A denial does not imply a new logical request identity. If the same `(run, tool)` later blocks again, the runtime may still use the same `GrantRequestID` while surfacing a new attempt count or state transition.

### Re-request

Repeated requests for the same `(run, tool)` are treated as renewed attempts of the same logical blocked request unless a future policy explicitly defines a stronger boundary.

For the current slice, `grant_attempt` counts only the latest active blocked cycle. Grant or denial clears that attempt counter, and a later fresh blocked cycle for the same `(run, tool)` reuses the stable logical `GrantRequestID` while restarting `grant_attempt` at `1`.

### Surface Semantics

`agent_wait`, runtime projection, and durable RunLedger mirror should all treat `GrantRequestID` as the stable logical request key. Any UI or operational distinction between "same request, another attempt" versus "brand-new logical request" must be driven by metadata other than the request ID itself.
