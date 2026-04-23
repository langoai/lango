# Release vs Refund Adjudication

This page describes the first `release vs refund adjudication` slice for `knowledge exchange v1`.

## Purpose

Release vs refund adjudication records the first canonical branch decision after dispute hold.

The slice is intentionally narrow:

- adjudication only
- canonical input is `transaction_receipt_id`
- escrow must already be `funded`
- settlement progression must already be `dispute-ready`
- dispute hold evidence must already exist
- success keeps escrow and settlement state unchanged
- actual release and refund execution remain separate tools

## What Ships

- a receipts-backed `adjudicate_escrow_dispute` meta tool
- transaction-level gating on funded escrow, `dispute-ready`, and recorded hold evidence
- canonical adjudication state on the transaction receipt
- adjudication evidence in the current submission receipt trail

## Current Limits

This slice does not yet include:

- automatic release or refund execution
- keep-hold or re-escalation outcomes
- richer dispute scoring or policy arbitration
- human adjudication UI
