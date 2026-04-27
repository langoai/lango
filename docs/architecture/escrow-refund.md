# Escrow Refund

This page describes the first `escrow refund` slice for `knowledge exchange v1`.

## Purpose

Escrow refund connects a funded, refund-adjudicated escrow to a direct refund execution path.

The slice is intentionally narrow:

- refund only
- canonical input is `transaction_receipt_id`
- escrow must already be `funded`
- settlement progression must already be `review-needed`
- `escrow_adjudication` must already be `refund`
- opposite-branch release evidence blocks execution
- success keeps settlement progression unchanged and clears the active dispute lifecycle marker
- failure also keeps settlement progression unchanged

## What Ships

- a receipts-backed `refund_escrow_settlement` meta tool
- transaction-level gating on funded escrow plus `review-needed` settlement state
- service-local per-transaction serialization so concurrent refund requests for the same transaction do not enter the refund runtime at the same time
- matching `escrow_adjudication = refund`
- one-way branch safety against opposite release evidence
- amount resolution from canonical transaction context
- existing escrow runtime reuse for refund
- success and failure evidence in the current submission receipt trail

## Current Limits

This slice does not yet include:

- a refund-specific terminal settlement state
- release reversal
- config-backed default execution-mode policy
- human refund UI
