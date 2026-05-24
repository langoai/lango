# Release vs Refund Adjudication

This page describes the current `release vs refund adjudication` slice for `knowledge exchange v1`.

## Purpose

Release vs refund adjudication records the canonical branch decision after dispute hold and moves settlement progression onto the matching release or refund path while preserving dispute lifecycle state for downstream recovery.

The slice is intentionally narrow:

- adjudication only
- canonical input is `transaction_receipt_id`
- escrow must already be `funded`
- settlement progression must already be `dispute-ready`
- dispute hold evidence must already exist
- success records `escrow_adjudication`
- success moves settlement progression atomically to `approved-for-settlement` (`release`) or `review-needed` (`refund`)
- success preserves the active dispute lifecycle marker
- release and refund execution remain separate follow-up tools unless explicitly requested through post-adjudication execution modes

## What Ships

- a receipts-backed `adjudicate_escrow_dispute` meta tool
- transaction-level gating on funded escrow, `dispute-ready`, and recorded hold evidence
- service-local per-transaction serialization so concurrent adjudication requests for the same transaction do not apply the branch update in parallel
- canonical adjudication state on the transaction receipt
- atomic progression update paired with adjudication
- adjudication evidence in the current submission receipt trail
- tool results that include:
  - `settlement_progression_status`
  - `dispute_lifecycle_status`
  - `escrow_reference`
  - `outcome`

## Current Limits

This slice does not yet include:

- adjudication outcomes beyond `release` and `refund`
- config-backed non-manual execution defaults
- richer dispute scoring or policy arbitration
- human adjudication UI
