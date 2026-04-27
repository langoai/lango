# Actual Settlement Execution

This page describes the first direct `actual settlement execution` slice for `knowledge exchange v1`.

## Purpose

Actual settlement execution connects `approved-for-settlement` transaction state to real direct money-moving execution.

The slice is intentionally narrow:

- direct settlement only
- canonical gate input is `transaction_receipt_id`
- amount is resolved from canonical transaction context
- success closes settlement progression to `settled`
- failure keeps settlement progression at `approved-for-settlement`

## What Ships

- a receipts-backed `execute_settlement` meta tool
- transaction-level execution gating on `approved-for-settlement`
- amount resolution from canonical `price_context`
- direct payment runtime reuse for final settlement transfer
- success and failure evidence in the current submission receipt trail

## Canonical Gate

`execute_settlement(transaction_receipt_id)` is allowed only when:

- the transaction receipt exists
- a current submission exists
- settlement progression is `approved-for-settlement`
- the settlement amount resolves from transaction context

Current deny reasons:

- `missing_receipt`
- `no_current_submission`
- `not_approved_for_settlement`
- `amount_unresolved`

## Success / Failure

On success:

- the direct payment runtime executes settlement
- settlement progression closes to `settled`
- success evidence is appended to the current submission trail

On failure:

- settlement progression remains `approved-for-settlement`
- failure evidence is appended to the current submission trail

This keeps settlement policy state separate from execution failure.

## Current Limits

This slice does not yet include:

- escrow release or refund execution
- partial settlement execution
- dispute engine behavior
- human settlement UI
- automatic runtime-wide settlement execution
