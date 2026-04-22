# Escrow Release Design

## Purpose / Scope

This design defines the first `escrow release` slice for `knowledge exchange v1`.

Its job is narrow:

- connect a funded escrow to real settlement completion
- support only the release path
- record success and failure evidence
- close canonical settlement progression on success

This slice covers:

- a new `release_escrow_settlement` meta tool
- transaction-level escrow release gating
- canonical source resolution from receipts and transaction context
- reuse of the existing escrow runtime
- audit and receipt-trail evidence for success and failure

This slice does not cover:

- escrow refund
- dispute-linked escrow handling
- human approval UI
- broader escrow lifecycle orchestration

## Execution Gate

The canonical input is:

- `transaction_receipt_id`

`release_escrow_settlement(transaction_receipt_id)` may proceed only when:

- the transaction receipt exists
- a current submission exists
- `escrow_execution_status = funded`
- `settlement_progression_status = approved-for-settlement`
- the settlement amount can be resolved from canonical transaction context

If any prerequisite is missing, execution is denied.

First-slice deny reasons:

- `missing receipt`
- `no current submission`
- `escrow not funded`
- `not approved-for-settlement`
- `amount unresolved`

This gate is not a new policy engine. It is the final execution layer that consumes already-canonical escrow and settlement progression state.

## Canonical Sources

This slice does not accept release target or amount as tool inputs.

Canonical sources:

- `transaction receipt`
  - `escrow_execution_status`
  - `settlement_progression_status`
  - `escrow_reference`
  - `price/payment context`
- `current submission`
  - current deliverable linkage
  - settled evidence anchor

The release amount is resolved from the transaction's existing amount context.

The escrow reference is resolved from the transaction receipt.

That means:

- no release amount parameter on the tool
- no release target parameter on the tool

This keeps escrow release aligned with existing runtime truth instead of opening a new caller-supplied execution surface.

## Success / Failure Semantics

On success:

- the existing escrow runtime executes release
- `transaction receipt.settlement_progression_status = settled`
- settled evidence is added to the current submission trail

On failure:

- canonical settlement progression remains `approved-for-settlement`
- escrow release failure is recorded as evidence and audit only

This keeps settlement policy state separate from execution failure.

The first slice does not add an intermediate release status.

Success path:

- `approved-for-settlement -> settled`

Failure path:

- `approved-for-settlement` remains unchanged
- failure evidence is appended

## Evidence Model

Both success and failure are recorded in:

- `audit log`
- `submission receipt trail`

Minimum success evidence:

- escrow release executed
- transaction reference
- submission reference
- escrow reference
- resolved amount context

Minimum failure evidence:

- escrow release failed
- failure reason
- transaction reference
- submission reference
- escrow reference

The ownership model remains:

- `transaction receipt` owns canonical settlement state
- `submission receipt trail` owns append-only execution evidence

That allows later reconstruction of:

- why escrow release executed
- why escrow release failed
- which submission was used as the settlement anchor

## Implementation Shape

Recommended structure:

- `release_escrow_settlement` meta tool
- shared `escrow release service`
- existing escrow runtime reuse

Flow:

1. the tool accepts `transaction_receipt_id`
2. the service loads the transaction receipt
3. the service validates current submission, funded escrow, approved-for-settlement, and amount resolution
4. the existing escrow runtime performs release
5. on success:
   - settlement progression moves to `settled`
   - submission trail is updated
   - audit is updated
6. on failure:
   - settlement progression remains unchanged
   - failure evidence is recorded
   - audit is updated

This keeps:

- the meta tool as a thin entrypoint
- the service as the canonical orchestration layer
- the escrow runtime as the actual release layer

## Follow-On Inputs

The next follow-on work after this slice is:

1. `refund`
   - reject and reversal path
   - escrow refund semantics

2. `dispute-linked escrow handling`
   - funded but contested transaction
   - release vs hold vs refund branching

3. `broader escrow lifecycle completion`
   - release/refund/dispute orchestration
   - milestone-aware escrow release
