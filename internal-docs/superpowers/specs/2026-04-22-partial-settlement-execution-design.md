# Partial Settlement Execution Design

## Purpose / Scope

This design defines the first `partial settlement execution` slice for `knowledge exchange v1`.

Its job is narrow:

- enable partial execution only on the direct settlement path
- execute only when a canonical partial hint exists
- move canonical progression to `partially-settled` on success
- keep progression unchanged on failure while recording evidence

This slice covers:

- a new `execute_partial_settlement` meta tool
- transaction-level execution gating
- canonical partial-hint parsing
- reuse of the existing direct payment runtime
- success and failure evidence
- remaining-amount canonicalization

This slice does not cover:

- escrow partial release
- percentage-based partial hints
- free-form hint parsing
- multi-round partial execution
- dispute engine behavior
- human settlement UI

## Execution Gate

The canonical input is:

- `transaction_receipt_id`

`execute_partial_settlement(transaction_receipt_id)` may proceed only when:

- the transaction receipt exists
- a current submission exists
- `settlement_progression_status = approved-for-settlement`
- a canonical `partial_settlement_hint` exists
- the hint is parseable as an absolute amount
- the transaction has not already been partially settled in this slice

If any prerequisite is missing, execution is denied.

First-slice deny reasons:

- `missing receipt`
- `no current submission`
- `not approved-for-settlement`
- `partial hint missing`
- `partial hint invalid`
- `already partially-settled`

This gate does not create new policy judgment. It only turns canonical partial-hint state into executable settlement behavior.

## Canonical Hint Model

This slice does not accept the partial amount as a tool parameter.

The amount must come only from:

- `transaction receipt.partial_settlement_hint`

The first slice supports only one hint form:

- `settle:0.40-usdc`

That means:

- absolute amount only
- no percentage-based hints yet
- no free-form text parsing

This keeps partial execution grounded in canonical state rather than caller-supplied execution inputs.

## Success / Failure Semantics

On success:

- the direct payment runtime executes the partial amount
- `settlement_progression_status = partially-settled`
- a new absolute remaining-amount hint is canonicalized
- partial settlement success evidence is appended to the current submission trail

On failure:

- progression remains `approved-for-settlement`
- failure evidence is appended

This keeps settlement policy state separate from money-moving execution failure.

The first slice allows only one partial execution per transaction.

That means:

- `approved-for-settlement -> partially-settled` is allowed
- repeated partial execution from `partially-settled` is a follow-on slice

## Evidence Model

Both success and failure are recorded in:

- `audit log`
- `submission receipt trail`

Minimum success evidence:

- partial settlement executed
- transaction reference
- submission reference
- executed partial amount
- remaining amount

Minimum failure evidence:

- partial settlement execution failed
- failure reason
- transaction reference
- submission reference
- attempted partial amount

This allows later reconstruction of:

- why only part of the settlement was executed
- how much executed
- what remains
- why execution failed

## Remaining Amount Canonicalization

The service calculates:

- total amount from `price_context`
- partial amount from `partial_settlement_hint`

It then writes the remaining amount back as a new absolute hint.

Example:

- total: `quote:1.00-usdc`
- partial hint: `settle:0.40-usdc`
- remaining hint after success: `settle:0.60-usdc`

The remaining amount is therefore a canonical state update, not free-form operator input.

## Implementation Shape

Recommended structure:

- `execute_partial_settlement` meta tool
- shared `partial settlement execution service`
- existing direct payment runtime reuse

Flow:

1. the tool accepts `transaction_receipt_id`
2. the service loads the transaction receipt
3. the service validates current submission, progression state, and partial hint
4. the service parses total amount and partial amount
5. the direct payment runtime executes the partial settlement
6. on success:
   - progression moves to `partially-settled`
   - remaining hint is canonicalized
   - submission trail is updated
   - audit is updated
7. on failure:
   - progression remains unchanged
   - failure evidence is recorded
   - audit is updated

This keeps:

- the meta tool as a thin entrypoint
- the service as the canonical orchestration layer
- the payment runtime as the actual transfer layer

## Follow-On Inputs

The next follow-on work after this slice is:

1. `multi-round partial settlement`
   - repeated partial execution from `partially-settled`
   - remaining amount exhaustion
   - final closeout from partial state

2. `percentage-based hint model`
   - forms like `settle:40%`
   - percent-to-amount canonicalization

3. `escrow partial release`
   - partial release on escrow-funded paths
   - escrow remaining balance semantics
