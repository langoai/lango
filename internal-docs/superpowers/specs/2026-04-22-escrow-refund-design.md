# Escrow Refund Design

## Purpose / Scope

This design defines the first `escrow refund` slice for `knowledge exchange v1`.

Its job is narrow:

- connect a funded but unreleased escrow to a refund execution path
- allow refund only from the settlement review path
- record success and failure evidence
- keep canonical settlement progression unchanged in this first slice

This slice covers:

- a new `refund_escrow_settlement` meta tool
- transaction-level refund gating
- canonical source resolution from receipts and transaction context
- reuse of the existing escrow runtime
- audit and receipt-trail evidence for success and failure

This slice does not cover:

- release reversal
- dispute-linked refund branching
- a refund-specific terminal settlement state
- human refund UI

## Execution Gate

The canonical input is:

- `transaction_receipt_id`

`refund_escrow_settlement(transaction_receipt_id)` may proceed only when:

- the transaction receipt exists
- a current submission exists
- `escrow_execution_status = funded`
- `settlement_progression_status = review-needed`
- the refund amount can be resolved from canonical transaction context

If any prerequisite is missing, execution is denied.

First-slice deny reasons:

- `missing receipt`
- `no current submission`
- `escrow not funded`
- `not review-needed`
- `amount unresolved`

This gate is not a new policy engine. It is the final execution layer that consumes already-canonical funded escrow and settlement review state.

## Canonical Sources

This slice does not accept refund target or amount as tool inputs.

Canonical sources:

- `transaction receipt`
  - `escrow_execution_status`
  - `settlement_progression_status`
  - `escrow_reference`
  - `price/payment context`
- `current submission`
  - current deliverable linkage
  - refund evidence anchor

The refund amount is resolved from the transaction's existing amount context.

The escrow reference is resolved from the transaction receipt.

That means:

- no refund amount parameter on the tool
- no refund target parameter on the tool

This keeps escrow refund aligned with existing runtime truth instead of allowing a new caller-supplied execution surface.

## Success / Failure Semantics

On success:

- the existing escrow runtime executes refund
- canonical settlement progression remains `review-needed`
- refund success evidence is added to the current submission trail

On failure:

- canonical settlement progression remains `review-needed`
- refund failure is recorded as evidence and audit only

This keeps settlement progression separate from refund execution outcome in the first slice.

The first slice does not define a refund-specific terminal progression state.

Success path:

- `review-needed` remains unchanged
- refund evidence is appended

Failure path:

- `review-needed` remains unchanged
- failure evidence is appended

## Evidence Model

Both success and failure are recorded in:

- `audit log`
- `submission receipt trail`

Minimum success evidence:

- escrow refund executed
- transaction reference
- submission reference
- escrow reference
- resolved amount context

Minimum failure evidence:

- escrow refund failed
- failure reason
- transaction reference
- submission reference
- escrow reference

The ownership model remains:

- `transaction receipt` owns canonical settlement state
- `submission receipt trail` owns append-only execution evidence

That allows later reconstruction of:

- why escrow refund executed
- why escrow refund failed
- which submission was used as the refund anchor

## Implementation Shape

Recommended structure:

- `refund_escrow_settlement` meta tool
- shared `escrow refund service`
- existing escrow runtime reuse

Flow:

1. the tool accepts `transaction_receipt_id`
2. the service loads the transaction receipt
3. the service validates current submission, funded escrow, review-needed state, and amount resolution
4. the existing escrow runtime performs refund
5. on success:
   - settlement progression remains unchanged
   - submission trail is updated
   - audit is updated
6. on failure:
   - settlement progression remains unchanged
   - failure evidence is recorded
   - audit is updated

This keeps:

- the meta tool as a thin entrypoint
- the service as the canonical orchestration layer
- the escrow runtime as the actual refund layer

## Follow-On Inputs

The next follow-on work after this slice is:

1. `refund terminal state`
   - define whether refund closes the transaction into a new terminal progression state

2. `dispute-linked refund branching`
   - review-needed vs dispute-ready branching
   - escrow hold/release/refund adjudication

3. `release-after-refund / refund-after-release safety rules`
   - one-way terminal guards
   - lifecycle consistency rules
