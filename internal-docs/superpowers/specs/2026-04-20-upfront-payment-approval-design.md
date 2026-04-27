# Upfront Payment Approval Design

## Purpose

Define the first explicit `upfront payment approval` model for `knowledge exchange v1`.

This design answers the control-plane question that sits before artifact work starts:

- when a leader agent may authorize an upfront payment,
- what inputs that decision must consider,
- what states the decision may return,
- what receipt should be produced,
- and how the result should update transaction-level canonical state.

This document is subordinate to:

- `docs/architecture/master-document.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `docs/architecture/trust-security-policy-audit.md`
- `internal-docs/superpowers/specs/2026-04-20-approval-flow-design.md`
- `internal-docs/superpowers/specs/2026-04-20-dispute-ready-receipts-design.md`

## Scope

This design covers:

- the `upfront payment approval` decision model,
- decision inputs,
- decision states,
- suggested payment modes,
- approval receipt subtype,
- transaction receipt updates,
- append-only event trail for payment approval.

This design does not yet cover:

- actual payment execution gating,
- escrow execution,
- human approval UI,
- full transaction orchestration,
- dispute adjudication for payment approval.

## Problem Statement

Lango now has:

- exportability decisions,
- artifact release approval,
- lite dispute-ready receipts.

What is still missing is the front-end control plane that decides whether the transaction may even open with an upfront payment.

Without this, Lango lacks an explicit model for:

- budget/policy-aware prepayment approval,
- trust-aware escalation,
- recommended payment mode selection,
- and durable receipt evidence that can later explain why a payment was approved, rejected, or escalated.

## Approaches Considered

### Approach A: Minimal Audit-Only

Record the decision only as an audit row.

Pros:

- smallest implementation,
- reuses existing audit storage.

Cons:

- weak linkage to transaction canonical state,
- hard to reuse in later execution gating,
- poor foundation for richer receipts.

### Approach B: Approval Receipt Subtype + Transaction Update

Model the result as a subtype of the approval receipt family and also update transaction-level canonical state.

Pros:

- aligns with the existing approval-flow direction,
- keeps point-in-time evidence and canonical state connected,
- reusable for later execution gates.

Cons:

- more state modeling than audit-only,
- requires explicit receipt/event design.

### Approach C: Full Payment Control Plane

Bundle approval, execution gating, escrow routing, and settlement execution into one slice.

Pros:

- end-to-end transaction opening flow.

Cons:

- too large for the first slice,
- mixes decisioning and execution too early.

## Recommendation

Use **Approach B: Approval Receipt Subtype + Transaction Update**.

This is the smallest model that still preserves:

- policy-aware approval,
- explicit escalation,
- reusable payment mode recommendation,
- and transaction-level canonical state.

## Decision Inputs

The first-slice input set is:

- `amount`
- `counterparty`
- `trust input`
  - `score`
  - `score source`
  - `recent risk flags`
- `requested scope`
- `budget / policy context`
  - `budget cap`
  - `remaining budget`
  - `user max prepay policy`
  - `counterparty-specific exception policy`
  - `transaction mode policy`

This means the first slice is already more than a simple amount threshold. It is a structured decision over trust, budget, and policy.

## Decision States

The decision state set is:

- `approve`
- `reject`
- `escalate`

### Approve

`approve` means:

- upfront payment is allowed under current policy,
- a suggested payment mode is available,
- and the result may be attached to the current transaction receipt.

Minimum associated fields:

- `suggested payment mode`
- `amount class`
- `risk class`

### Reject

`reject` means:

- the transaction may not proceed through the upfront payment path under current policy.

Minimum associated fields:

- `reason`
- `policy code`
- `budget / trust failure detail`

### Escalate

`escalate` is reserved for narrow boundary cases:

- high amount,
- low-trust edge case,
- policy ambiguity.

It is not the default path.

## Suggested Payment Mode

The first slice returns a payment mode recommendation, not actual execution.

Minimum mode set:

- `prepay`
- `escrow`
- `escalate`
- `reject`

`escrow` here is only a recommendation signal that direct prepay is not the right path for this transaction.

## Classification Outputs

The first slice also emits:

- `amount class`
  - `low`
  - `medium`
  - `high`
  - `critical`
- `risk class`

These are not just labels. They are inputs for later execution gating and reporting.

## Receipt Model

This slice uses the existing approval receipt family, not a wholly separate payment-only system.

Flow:

1. create `upfront payment approval receipt`
2. link that receipt to the relevant `transaction receipt`
3. update transaction-level canonical payment approval state

This preserves both:

- point-in-time approval evidence
- transaction-level canonical state

## Transaction Receipt Coupling

The first slice updates at least:

- `canonical decision`
- `settlement hint`
- `current payment approval status`

`current payment approval status` is:

- `pending`
- `approved`
- `rejected`
- `escalated`

This allows the transaction receipt to reflect whether the transaction is still waiting for a payment-opening decision, has cleared it, or has been blocked/escalated.

## Event Trail

This slice also has append-only event history.

It stores:

- payment approval decision state
- reason / policy code
- amount class
- risk class
- related refs

The design intention is the same as elsewhere:

- canonical current state for quick reads
- event trail for later reconstruction

## Non-Goals

This design intentionally does not claim that:

- payment execution is now gated by this decision,
- escrow execution is implemented,
- human approval UI exists,
- or full transaction orchestration is complete.

## Initial Success Criteria

This design is successful if a future implementation can:

- evaluate an upfront payment request from structured trust and budget inputs,
- emit `approve / reject / escalate`,
- include suggested payment mode and amount/risk classes,
- produce an approval receipt subtype,
- update the linked transaction receipt canonical payment approval state,
- and append an event trail entry for later reconstruction.

## Follow-On Planning Inputs

The next implementation plan should define:

1. the domain model for upfront payment approval
2. the receipt subtype shape
3. the transaction receipt fields that get updated
4. the append-only payment approval event model
5. the minimal operator and docs surface for the first slice
