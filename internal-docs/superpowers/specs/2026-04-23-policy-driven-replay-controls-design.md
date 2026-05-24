# Policy-Driven Replay Controls Design

## Purpose / Scope

This design defines the first `policy-driven replay controls` slice for `retry_post_adjudication_execution` in `knowledge exchange v1`.

Its job is narrow:

- add an actor-based authorization gate on top of the canonical replay gate
- resolve the actor from existing runtime context
- allow replay only when the actor is permitted for the current replay outcome

This slice covers:

- a replay-service-local policy gate
- actor resolution from existing session / approval context
- outcome-aware replay allowlists
- fail-closed deny behavior

This slice does not cover:

- human approval UI
- org-level policy editing
- richer policy engines
- per-transaction policy snapshots
- amount-tier replay rules

## Policy Model

The policy source for this first slice is current configuration.

Minimum policy shape:

- `replay.allowed_actors`
- `replay.release_allowed_actors`
- `replay.refund_allowed_actors`

Meaning:

- `allowed_actors`
  - broad replay allowlist
- `release_allowed_actors`
  - actors allowed to replay `release` outcomes
- `refund_allowed_actors`
  - actors allowed to replay `refund` outcomes

This first slice therefore evaluates:

- actor identity
- replay outcome

and does not yet inspect amount tiers or richer policy classes.

## Actor Resolution

The actor is not passed as a tool parameter.

Instead, the replay service resolves it from:

- the current session context
- or approval-related runtime context already carried through the request

This first slice is fail-closed:

- if the actor cannot be resolved, replay is denied
- if the actor is resolved but not allowed, replay is denied

This avoids ambiguity on a dead-letter recovery path.

## Replay Gate Extension

Before this slice, replay is allowed only when:

- dead-letter evidence exists
- canonical adjudication still exists
- the current submission still exists

After this slice, replay additionally requires:

- actor resolved from runtime context
- actor allowed for the current replay outcome

The replay service therefore checks:

1. dead-letter evidence exists
2. canonical adjudication still exists
3. current submission still exists
4. actor can be resolved
5. actor is allowed for the replay outcome

This keeps the canonical replay gate and authorization gate in one service layer.

## Failure Semantics

New deny reasons for this slice are:

- `actor_unresolved`
- `replay_not_allowed`

`actor_unresolved` means:

- the replay service could not derive an actor identity from runtime context

`replay_not_allowed` means:

- the actor was resolved
- but the allowlist policy does not permit replay for the current outcome

Existing canonical replay deny cases remain unchanged:

- dead-letter missing
- adjudication missing
- current submission missing

## Configuration Shape

Example first-slice configuration:

- `replay.allowed_actors = ["operator:alice", "operator:bob"]`
- `replay.release_allowed_actors = ["operator:alice"]`
- `replay.refund_allowed_actors = ["operator:alice", "operator:bob"]`

The evaluation can remain intentionally simple:

- actor must be present in `allowed_actors`
- actor must also be present in the outcome-specific allowlist for the replay outcome

This keeps authorization explicit and easy to audit.

## Follow-On Inputs

The next follow-on work after this slice is:

1. `policy-driven replay controls v2`
   - amount-tier rules
   - failure-type rules
   - richer actor classes

2. `operator management`
   - UI or config editing surface for replay policy

3. `per-transaction policy snapshots`
   - decide whether replay-time policy should be persisted alongside replay evidence
