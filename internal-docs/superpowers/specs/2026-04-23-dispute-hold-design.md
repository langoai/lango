# Dispute Hold Design

## Purpose / Scope

This design defines the first `dispute hold` slice for `knowledge exchange v1`.

Its job is narrow:

- connect a funded escrow to a dispute-aware hold path
- record that hold state as canonical evidence
- avoid deciding release vs refund in the same slice

This slice covers:

- a new `hold_escrow_for_dispute` meta tool
- transaction-level dispute hold gating
- canonical source resolution from receipts
- dispute hold evidence recording

This slice does not cover:

- release vs refund adjudication
- dispute resolution engine behavior
- human adjudication UI
- a new escrow terminal state

## Execution Gate

The canonical input is:

- `transaction_receipt_id`

`hold_escrow_for_dispute(transaction_receipt_id)` may proceed only when:

- the transaction receipt exists
- a current submission exists
- `escrow_execution_status = funded`
- `settlement_progression_status = dispute-ready`
- an escrow reference exists

If any prerequisite is missing, execution is denied.

First-slice deny reasons:

- `missing receipt`
- `no current submission`
- `escrow not funded`
- `not dispute-ready`
- `escrow reference missing`

This gate is not a new policy engine. It is the first execution-control layer that turns already-canonical dispute-ready funded escrow into a holdable runtime state.

## Canonical Sources

This slice does not accept escrow identifiers or extra hold metadata as tool inputs.

Canonical sources:

- `transaction receipt`
  - `escrow_execution_status`
  - `settlement_progression_status`
  - `escrow_reference`
- `current submission`
  - current deliverable linkage
  - dispute hold evidence anchor

That means:

- the escrow reference is resolved from the transaction receipt
- the current submission is resolved from the transaction receipt
- the tool does not accept a separate escrow identifier

This keeps dispute hold aligned with the same transaction-level truth used by the existing escrow execution slices.

## Success / Failure Semantics

On success:

- `escrow_execution_status` remains `funded`
- `settlement_progression_status` remains `dispute-ready`
- dispute hold success evidence is appended

On failure:

- transaction state remains unchanged
- dispute hold failure evidence is appended

The first slice does not add a new `held` state.

Instead, it records the hold as evidence first, so later release/refund adjudication can consume that evidence without the current slice over-claiming broader lifecycle completion.

## Evidence Model

Both success and failure are recorded in:

- `audit log`
- `submission receipt trail`

Minimum success evidence:

- escrow hold applied
- transaction reference
- submission reference
- escrow reference

Minimum failure evidence:

- escrow hold failed
- failure reason
- transaction reference
- submission reference
- escrow reference

The ownership model remains:

- `transaction receipt` owns canonical transaction state
- `submission receipt trail` owns append-only execution evidence

That allows later reconstruction of:

- why a funded escrow was held for dispute
- why hold application failed
- which submission anchored the hold evidence

## Implementation Shape

Recommended structure:

- `hold_escrow_for_dispute` meta tool
- shared `dispute hold service`
- existing escrow control plane reuse

Flow:

1. the tool accepts `transaction_receipt_id`
2. the service loads the transaction receipt
3. the service validates current submission, funded escrow, dispute-ready state, and escrow reference
4. on success:
   - hold success evidence is appended
   - audit is updated
5. on failure:
   - hold failure evidence is appended
   - audit is updated

This keeps:

- the meta tool as a thin entrypoint
- the service as the canonical orchestration layer
- release/refund adjudication explicitly out of scope for this slice

## Follow-On Inputs

The next follow-on work after this slice is:

1. `release vs refund adjudication`
   - define the decision path after dispute hold

2. `dispute-linked escrow state`
   - decide whether evidence-only hold is enough or whether a later explicit `held` state is needed

3. `dispute engine integration`
   - connect adjudication outcomes to escrow lifecycle transitions
