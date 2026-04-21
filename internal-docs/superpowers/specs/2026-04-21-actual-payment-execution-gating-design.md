# Actual Payment Execution Gating Design

## Purpose

Define the first explicit `actual payment execution gating` model for `knowledge exchange v1`.

This design answers the next control question after:

- exportability,
- artifact release approval,
- dispute-ready receipts,
- upfront payment approval.

Specifically:

- when a direct payment execution may actually proceed,
- what canonical state it must rely on,
- what deny reasons it may return,
- and what evidence must be recorded for both allow and deny outcomes.

This document is subordinate to:

- `docs/architecture/master-document.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `docs/architecture/trust-security-policy-audit.md`
- `docs/superpowers/specs/2026-04-20-upfront-payment-approval-design.md`
- `docs/superpowers/specs/2026-04-20-dispute-ready-receipts-design.md`

## Scope

This design covers:

- direct payment execution gating for `payment_send`
- direct payment execution gating for `p2p_pay`
- a shared gate service
- `allow / deny` decisioning
- deny reason codes
- audit event recording
- receipt event recording

This design does not yet cover:

- escrow execution
- human payment approval UI
- full transaction orchestration
- middleware-wide enforcement for all payment-like tools

## Problem Statement

Lango now has:

- pre-execution payment approval
- approval receipts
- transaction-level canonical payment approval state

But it still lacks the final execution control layer that answers:

- may this payment actually execute now?

Without this layer, payment approval remains advisory. The system can decide that a payment should be allowed, but it still cannot enforce that decision at the actual execution boundary of `payment_send` and `p2p_pay`.

## Approaches Considered

### Approach A: Thin Handler Guard

Put narrow `if` checks directly into each payment tool handler.

Pros:

- fastest to implement,
- very little abstraction.

Cons:

- easy drift between tools,
- duplicate reasoning and event recording,
- poor foundation for later expansion.

### Approach B: Handler-Local Wiring + Shared Gate Service

Keep handlers as the integration point, but centralize gate logic in one shared service.

Pros:

- small enough for the first slice,
- common logic stays consistent,
- handlers remain thin,
- easily extendable later.

Cons:

- slightly more design than a direct inline check.

### Approach C: Middleware Gate

Move payment execution gating into a reusable middleware layer for all payment tools.

Pros:

- most uniform long term.

Cons:

- too large for the next slice,
- risks entangling unrelated tool flows too early.

## Recommendation

Use **Approach B: Handler-Local Wiring + Shared Gate Service**.

This is the smallest model that:

- gives a real execution control surface,
- avoids drift between `payment_send` and `p2p_pay`,
- and preserves a clean path toward broader enforcement later.

## Execution Targets

The first slice gates only:

- `payment_send`
- `p2p_pay`

This is intentional. It keeps the first slice focused on direct payment execution paths only.

Escrow execution is explicitly deferred.

## Gate Inputs

The gate uses:

- `transaction_receipt_id` as the primary lookup key
- optional linked `submission_receipt_id` for deeper validation and traceability
- tool name
- tool-specific execution context

Tool-specific context:

- `payment_send`: `to`, `amount`, `purpose`
- `p2p_pay`: `peer_did`, `amount`, `memo`

If `transaction_receipt_id` is missing:

- the direct payment execution is denied immediately

This makes the first slice explicitly receipt-backed.

## Decision Model

The execution gate is binary:

- `allow`
- `deny`

It is not a new policy engine. It consumes upstream canonical state rather than inventing new approval logic.

## Canonical Source Of Truth

The first read is always the linked `transaction receipt`.

The gate uses:

- `current payment approval status`
- `canonical decision`
- `canonical settlement hint`

If needed, it may follow linked receipt references for deeper validation.

The default rule is:

- use `transaction receipt` canonical state first
- use linked receipt detail only for confirmation and traceability

## Deny Reasons

The first slice supports these minimum deny reason codes:

- `missing_receipt`
- `approval_not_approved`
- `stale_state`
- `execution_mode_mismatch`

### Missing Receipt

No usable transaction receipt exists for this execution request.

### Approval Not Approved

The transaction's canonical payment approval state is not `approved`.

### Stale State

This covers at least:

- the approval has become too old to trust,
- or the transaction receipt's canonical state changed after approval.

### Execution Mode Mismatch

The direct payment execution path only allows approved `prepay`.

If the canonical payment recommendation is:

- `escrow`
- `escalate`
- or `reject`

then direct payment execution is denied as a mode mismatch.

## Execution Mode Rule

This first slice defines a deliberately narrow rule:

- direct payment execution only permits canonical `prepay`

This means:

- `payment_send` and `p2p_pay` may proceed only when the canonical payment approval state and mode match a direct prepay path
- the gate does not attempt to reinterpret an escrow recommendation as permission to send directly

## Placement

The gate is integrated in handlers, but its decision logic lives in a shared service.

That means:

- `payment_send` calls the shared gate service
- `p2p_pay` calls the same shared gate service
- both reuse the same reason codes and recording behavior

## Recording Model

Both `allow` and `deny` are recorded.

### Allow

When execution is allowed, the system records:

- `payment execution authorized`
- linked canonical transaction or receipt references

### Deny

When execution is denied, the system records:

- `payment execution denied`
- deny reason code
- linked canonical transaction or receipt references

These outcomes are written to:

- `audit log`
- `receipt event trail`

## Receipt Trail Coupling

The gate does not rewrite approval outcomes.

Instead, it appends execution-layer evidence to the receipt trail, so later systems can reconstruct:

- approval outcome
- execution authorization or denial
- settlement execution result

as separate but linked layers.

## Non-Goals

This design intentionally does not claim that:

- actual escrow execution is gated here,
- human UI exists for payment execution approval,
- middleware-wide payment gating is implemented,
- or the whole transaction runtime is now centralized.

## Initial Success Criteria

This design is successful if a future implementation can:

- deny any direct payment execution without `transaction_receipt_id`
- allow direct payment execution only for canonical `prepay`
- return structured deny reasons
- emit allow and deny events into both audit and receipt trail
- keep `payment_send` and `p2p_pay` behavior consistent through one shared gate service

## Follow-On Planning Inputs

The next implementation plan should define:

1. the shared gate service interface
2. how `transaction_receipt_id` reaches the two payment tools
3. how stale-state checks are represented in the first slice
4. what audit event payloads and receipt event payloads look like
5. how later escrow execution can plug into the same framework
