# Actual Settlement Execution Design

## Purpose / Scope

This design defines the first `actual settlement execution` slice for `knowledge exchange v1`.

Its job is narrow:

- connect `approved-for-settlement` canonical state to real money-moving execution
- support only the direct settlement path
- record success and failure evidence
- close canonical settlement progression on success

This slice covers:

- a new `execute_settlement` meta tool
- transaction-level execution gating
- canonical source resolution from receipts and transaction context
- reuse of the existing direct payment runtime
- audit and receipt-trail evidence for success and failure

This slice does not cover:

- escrow release or refund execution
- partial settlement calculation or execution
- dispute engine behavior
- automatic runtime-wide settlement execution
- human settlement UI

## Execution Gate

The canonical input is:

- `transaction_receipt_id`

`execute_settlement(transaction_receipt_id)` may proceed only when:

- the transaction receipt exists
- a current submission exists
- `settlement_progression_status = approved-for-settlement`
- the settlement amount can be resolved from canonical transaction context

If any prerequisite is missing, execution is denied.

First-slice deny reasons:

- `missing receipt`
- `no current submission`
- `not approved-for-settlement`
- `amount unresolved`

This gate is not a new policy decision engine. It is the final execution layer that consumes already-canonical settlement progression.

## Canonical Sources

This slice does not accept settlement target or amount as tool inputs.

Canonical sources:

- `transaction receipt`
  - settlement progression state
  - counterparty baseline
  - price and payment context
- `current submission`
  - current deliverable linkage
  - settled evidence anchor

The settlement amount is resolved from the transaction's existing `price/payment context`.

The settlement target is resolved from the transaction receipt and current submission context.

That means:

- no `to` parameter on the tool
- no `amount` parameter on the tool

This keeps actual settlement execution aligned with existing runtime truth instead of allowing a new caller-supplied execution surface.

## Success / Failure Semantics

On success:

- the existing direct payment runtime executes final settlement
- `transaction receipt.settlement_progression_status = settled`
- settled evidence is added to the current submission trail

On failure:

- canonical settlement progression remains `approved-for-settlement`
- execution failure is recorded as evidence and audit only

This keeps settlement policy state separate from money-moving execution failure.

The first slice does not add an intermediate execution status.

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

- settlement executed
- transaction reference
- submission reference
- resolved amount context

Minimum failure evidence:

- settlement execution failed
- deny or execution failure reason
- transaction reference
- submission reference

The ownership model remains:

- `transaction receipt` owns canonical settlement state
- `submission receipt trail` owns append-only execution evidence

That allows later reconstruction of:

- why settlement executed
- why settlement failed
- which submission was used as the settlement anchor

## Implementation Shape

Recommended structure:

- `execute_settlement` meta tool
- shared `settlement execution service`
- existing direct payment runtime reuse

Flow:

1. the tool accepts `transaction_receipt_id`
2. the service loads the transaction receipt
3. the service validates current submission, `approved-for-settlement`, and amount resolution
4. the existing direct payment runtime performs the settlement transfer
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
- the payment runtime as the actual transfer layer

## Follow-On Inputs

The next follow-on work after this slice is:

1. `partial settlement`
   - amount split rules
   - `partially-settled` execution path
   - remaining balance semantics

2. `escrow lifecycle completion`
   - release
   - refund
   - dispute-linked escrow handling

3. `dispute handoff integration`
   - distinction between settlement execution failure and dispute readiness
   - disagreement-based dispute opening after settlement review
