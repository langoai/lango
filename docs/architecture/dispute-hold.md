# Dispute Hold

This page describes the first `dispute hold` slice for `knowledge exchange v1`.

## Purpose

Dispute hold connects a funded escrow in canonical `dispute-ready` state to a hold evidence path.

The slice is intentionally narrow:

- hold only
- canonical input is `transaction_receipt_id`
- escrow must already be `funded`
- settlement progression must already be `dispute-ready`
- success keeps escrow and settlement state unchanged
- failure also keeps escrow and settlement state unchanged

## What Ships

- a receipts-backed `hold_escrow_for_dispute` meta tool
- transaction-level gating on funded escrow plus `dispute-ready` settlement state
- escrow reference resolution from canonical transaction context
- dispute hold success and failure evidence in the current submission receipt trail

## Current Limits

This slice does not yet include:

- release vs refund adjudication
- a new escrow `held` terminal or lifecycle state
- dispute resolution engine behavior
- human adjudication UI
