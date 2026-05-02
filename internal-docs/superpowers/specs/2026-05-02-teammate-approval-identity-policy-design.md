# Teammate Approval Identity Policy Design

## Goal

Define a clear identity and lifecycle policy for built-in teammate approval requests so runtime behavior, operator-facing surfaces, and durable RunLedger state all interpret repeated blocked requests the same way.

This design focuses on the semantics of:

- `GrantRequestID`
- repeated approval-blocked requests for the same `(run, tool)`
- denial and re-request behavior
- the boundary between logical request identity and per-attempt metadata

## Problem

The current built-in teammate runtime generates approval request IDs deterministically:

- `grant-{runID}-{toolName}`

That is enough for the current runtime to:

- project `blocked_waiting_approval`
- clear blocked state after `ApplyGrant()`
- mirror blocked state into RunLedger

But it leaves the actual contract underspecified.

The missing question is not "what string format do we use?" but rather:

- whether a second block for the same `(run, tool)` is the same logical request or a new one
- how denial should affect future requests
- how `agent_wait`, runtime projection, and durable mirror should describe repeated attempts

Without a policy, different layers can drift:

- runtime may treat a later block as the same request
- a UI may treat it as a new request
- RunLedger may preserve only the latest request identity without attempt context

## Decision

The chosen policy is:

- `GrantRequestID` is a **stable logical request identity**
- the logical identity is keyed by `(runID, toolName)`
- repeated approval blocks for the same `(run, tool)` reuse the same `GrantRequestID`
- retry, denial, timeout, or renewed-request distinctions are represented through **attempt metadata**, not by rotating the logical request ID

## Why This Direction

### 1. Runtime Simplicity

The current runtime already uses `GrantRequestID` for:

- blocked projection writes
- grant clearing
- stale blocked reconciliation

Keeping one stable logical request identity per `(run, tool)` preserves that simplicity. Attempt-scoped IDs would force more stateful matching logic into:

- `CapabilityPolicy.Evaluate(...)`
- `CapabilityRuntime.HandleBlockedToolCall(...)`
- `CapabilityRuntime.ApplyGrant(...)`

That is avoidable complexity.

### 2. Operator Interpretability

What operators usually care about first is:

- this run is blocked on this tool

That is a logical request concept, not an attempt concept.

A stable request ID lets all surfaces say:

- this is the same blocked request for `(run, tool)`

Then, if needed, a separate field can answer:

- this is the second attempt
- this was denied before
- this request was re-issued after timeout

### 3. Durable Mirror Compatibility

The current durable mirror for approval-blocked teammate state already preserves:

- `runtime_condition`
- `blocked_reason`
- `grant_request_id`

That mirror becomes easier to reason about when `grant_request_id` is stable. Attempt-scoped IDs would require either:

- more event stitching to infer logical sameness
- or another stable identity field anyway

So the cleaner model is:

- stable logical request identity
- optional attempt metadata layered on top

## Identity Model

### Stable Logical Request Identity

For a built-in teammate approval request:

- logical identity = `(runID, toolName)`
- `GrantRequestID` is the durable and runtime-visible encoding of that logical identity

Consequences:

- if the same run blocks again on the same tool, the same logical request ID is reused
- the runtime does not treat each repeated block as a brand-new logical request
- deduplication can happen at the logical request level

### Attempt Metadata

Attempt metadata is intentionally separate from `GrantRequestID`.

Representative fields may include:

- `grant_attempt`
- `grant_state`
- `last_attempt_at`
- `last_denied_at`
- `last_timeout_at`

This design does not require all of these fields to ship immediately. It only fixes the policy boundary:

- `GrantRequestID` = logical identity
- attempt metadata = lifecycle detail

## Semantics

### Repeated Block

If a run blocks again on the same tool:

- reuse the same `GrantRequestID`
- treat it as another attempt of the same logical request

### Denial

A denial does not force a new logical request identity.

If the same run and tool later require approval again:

- the logical identity may stay the same
- attempt metadata should show that this is a renewed attempt after denial

### Timeout

An approval timeout also does not require a new logical request identity by itself.

If the request is surfaced again later:

- the same logical request ID may be reused
- attempt metadata should capture the renewed attempt or timeout history

### Grant

Grant behavior remains tied to the logical request:

- successful grant clears the blocked projection for that logical request
- it should not require rotating request identity just to acknowledge completion

## Surface Expectations

### Runtime Projection

Projected blocked state should continue to expose:

- `RuntimeCondition`
- `BlockedReason`
- `GrantRequestID`

If attempt-level fields are introduced later, they should be additive rather than replacing `GrantRequestID`.

### `agent_wait`

`agent_wait` should interpret `grant_request_id` as:

- the stable logical blocked request key

If repeated attempts need to be distinguished in the response shape, they should appear through additional fields rather than a changing request ID.

### RunLedger Durable Mirror

RunLedger should preserve the same semantics:

- `grant_request_id` in the mirror identifies the logical request
- repeated attempts should not require rotating the request ID in durable state

If attempt metadata is later mirrored, it should complement the logical request ID rather than redefine it.

## In Scope

- define `GrantRequestID` policy as stable logical identity
- define repeated-request semantics
- define denial and timeout consequences at the policy level
- define the boundary between logical identity and attempt metadata

## Out of Scope

- redesigning the approval UI
- introducing all possible attempt metadata fields immediately
- redefining built-in capability scope rules
- changing the durable mirror schema beyond what is needed to support the chosen identity policy

## Recommended Implementation Direction

Implementation should likely proceed in two layers:

1. **Contract layer**
   - document that `GrantRequestID` is stable per `(run, tool)`
   - document repeated-request and denial semantics

2. **Runtime/storage layer**
   - keep current stable ID behavior where possible
   - add explicit attempt metadata only where it provides concrete operator value

That avoids unnecessary churn while making the policy explicit enough for future UI, runtime, and durability work.
