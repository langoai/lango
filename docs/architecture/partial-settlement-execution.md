# Partial Settlement Execution

This page describes the first direct `partial settlement execution` slice for `knowledge exchange v1`.

## Purpose

Partial settlement execution connects canonical `partial_settlement_hint` state to a one-shot direct settlement execution path.

The slice is intentionally narrow:

- direct settlement only
- canonical input is `transaction_receipt_id`
- canonical partial amount comes only from `partial_settlement_hint`
- success moves progression to `partially-settled`
- failure keeps progression at `approved-for-settlement`

## What Ships

- a receipts-backed `execute_partial_settlement` meta tool
- transaction-level gating on `approved-for-settlement`
- canonical absolute partial-hint parsing in form `settle:<amount>-usdc`
- direct payment runtime reuse for one-shot partial execution
- canonical remaining-amount hint updates after success
- success and failure evidence in the current submission receipt trail

## Canonical Hint Model

The slice accepts canonical absolute hints of the form `settle:<amount>-usdc`.

That means:

- absolute amount only
- no percentage-based hints
- no free-form parsing
- no multi-round partial execution

## Success / Failure

On success:

- the direct payment runtime executes the partial amount
- settlement progression moves to `partially-settled`
- the remaining amount is canonicalized back to a new absolute hint
- success evidence is appended to the current submission trail

On failure:

- settlement progression remains `approved-for-settlement`
- failure evidence is appended to the current submission trail

## Current Limits

This slice does not yet include:

- repeated partial execution from `partially-settled`
- percentage-based partial hints
- escrow partial release
- dispute engine behavior
- human settlement UI
