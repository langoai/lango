# Escrow Refund

This page describes the first `escrow refund` slice for `knowledge exchange v1`.

## Purpose

Escrow refund connects a funded but unreleased escrow to a direct refund execution path.

The slice is intentionally narrow:

- refund only
- canonical input is `transaction_receipt_id`
- escrow must already be `funded`
- settlement progression must already be `review-needed`
- success keeps settlement progression unchanged
- failure also keeps settlement progression unchanged

## What Ships

- a receipts-backed `refund_escrow_settlement` meta tool
- transaction-level gating on funded escrow plus `review-needed` settlement state
- amount resolution from canonical transaction context
- existing escrow runtime reuse for refund
- success and failure evidence in the current submission receipt trail

## Current Limits

This slice does not yet include:

- a refund-specific terminal settlement state
- dispute-linked refund branching
- release reversal
- human refund UI
