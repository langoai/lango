# Settlement Progression Design

## Purpose

This document defines the settlement progression framing for `knowledge exchange v1`.

Its purpose is not to define the full money-moving execution layer. Instead, it establishes the transaction-level progression model that begins after artifact release approval and determines:

- when settlement should move forward automatically,
- when settlement should pause for review,
- when a transaction becomes dispute-ready,
- what belongs to progression state versus execution state.

This design is intended as a follow-on design input to the broader `knowledge exchange runtime` work.

## Scope

This design covers:

- transaction-level settlement progression state,
- mapping release-approval outcomes into settlement progression,
- review-needed and higher-approval-needed branches,
- partial-settlement state and hint handling,
- dispute-ready handoff.

This design does not directly define:

- actual fund movement calculations,
- settlement executor internals,
- human adjudication UI,
- a full dispute engine,
- full escrow lifecycle completion.

This is a state-progression design, not the full settlement runtime implementation.

## Recommended Model

The preferred model is **Transaction-Level Settlement Progression**.

The key idea is:

- settlement progression is owned by the transaction,
- submission receipts provide evidence and reasoning,
- progression state is separate from the actual money-moving executor.

This is preferred over a submission-first or settlement-record-first model because the currently landed runtime and receipt slices already center on transaction receipts for canonical control-plane state.

## Canonical Progression Model

The progression unit is the `transaction`.

`transaction receipt` should own the canonical settlement progression state.

The core state set is:

- `pending`
- `in-progress`
- `review-needed`
- `approved-for-settlement`
- `partially-settled`
- `settled`
- `dispute-ready`

These states mean:

- `pending` — progression has not yet begun
- `in-progress` — progression has started but is not complete
- `review-needed` — the transaction cannot move straight to settlement completion
- `approved-for-settlement` — the executor is allowed to perform settlement work
- `partially-settled` — only part of the settlement has been completed
- `settled` — settlement is complete
- `dispute-ready` — the transaction has crossed from review into formal disagreement territory

This keeps the transaction receipt as the canonical answer to “where is this deal in settlement progression?”

## Outcome Mapping

Release approval outcomes should map into settlement progression as follows.

### Approve

`approve` moves the transaction into:

- `approved-for-settlement`

From there, automatic progression may move it toward:

- `in-progress`
- `settled`

The exact fund movement still belongs to a separate executor, but the progression state already allows that execution to proceed.

### Request Revision

`request-revision` moves the transaction into:

- `review-needed`

Progression pauses there while the same transaction waits for a new submission.

Revision therefore does not create a new transaction. It keeps the current transaction open and expects a new current submission.

### Reject

`reject` also moves the transaction into:

- `review-needed`

It does not immediately become `dispute-ready`.

This is intentional. A rejected submission may still resolve through operator review, clarification, or a subsequent submission without immediately becoming a formal dispute.

### Escalate

`escalate` means the transaction requires:

- higher approval,
- adjudication,
- or stronger operator review.

At the canonical state level, this still maps into:

- `review-needed`

But the reason layer attached to the transaction should preserve that this is specifically a higher-approval or adjudication-needed situation rather than a simple revision request.

## State Ownership

### Transaction Receipt

The transaction receipt should own:

- canonical settlement progression state,
- progression reason,
- partial-settlement hint,
- dispute-ready readiness.

### Submission Receipt

The submission receipt should not own progression itself.

Instead, it should supply:

- release-approval evidence,
- fulfillment assessment,
- settlement hints,
- escalation or rejection reasoning,
- material later used for dispute handoff.

This keeps the transaction receipt responsible for the deal-level progression while submission receipts remain the evidence layer for particular deliverables.

## Automatic vs Execution Boundaries

Automatic progression should stop at state movement.

That means the progression layer is responsible for:

- converting release-approval outcomes into progression states,
- opening or pausing settlement paths,
- deciding whether the transaction is review-needed or dispute-ready.

The progression layer is **not** responsible for:

- actually moving funds,
- actually releasing or refunding funds,
- executing settlement-side transactions.

Those belong to a separate settlement executor.

This distinction matters because reasoning failures and fund-movement failures are different classes of failure and should not be collapsed into one layer.

## Partial Settlement

This design does not yet define a full calculation model for partial settlement.

It only locks the idea that:

- `partially-settled` is a valid transaction-level state,
- partial settlement should be driven by explicit hints and evidence,
- the actual amount or ratio calculation remains follow-on work.

This preserves space for later executor and adjudication work without over-specifying a premature formula.

## Dispute-Ready Handoff

`dispute-ready` should not open on every `reject` or every `escalate`.

It should open only when:

- the transaction is already in `review-needed`,
- or in a higher-approval-needed condition,
- and a concrete disagreement is confirmed.

The disagreement classes should include at least:

- settlement disagreement,
- fulfillment disagreement,
- policy disagreement.

This makes `dispute-ready` narrower than generic rejection and ensures that formal disputes remain a specific handoff state rather than a synonym for “the reviewer said no.”

## Follow-On Implementation Inputs

This design should feed directly into:

### 1. Transaction Receipt Extension

The runtime will need explicit transaction-level fields for:

- settlement progression state,
- progression reason,
- partial-settlement hint,
- dispute-ready marker.

### 2. Release Outcome Mapper

The runtime will need a mapper that turns:

- `approve`
- `request-revision`
- `reject`
- `escalate`

into the canonical settlement progression model described here.

### 3. Settlement Executor Boundary

The runtime will need a distinct executor boundary for actual settlement work once a transaction reaches `approved-for-settlement`.

### 4. Disagreement Classifier

The runtime will need a small disagreement-classification layer to decide when a review-needed transaction actually becomes dispute-ready.

## Deliverable Expectation

The final document that follows from this design should remain a design-level description of settlement progression for `knowledge exchange v1`.

It should not yet expand into the full settlement executor or dispute engine implementation plan.
