# Dispute Hold

This page describes the current `dispute hold` slice for `knowledge exchange v1`.

## Purpose

Dispute hold connects a funded escrow in canonical `dispute-ready` state to explicit hold evidence and the active dispute lifecycle marker used by downstream adjudication and recovery.

The slice is intentionally narrow:

- hold only
- canonical input is `transaction_receipt_id`
- escrow must already be `funded`
- settlement progression must already be `dispute-ready`
- success keeps escrow and settlement progression unchanged at `dispute-ready`
- success sets `dispute_lifecycle_status = hold-active`
- failure keeps escrow, settlement progression, and dispute lifecycle state unchanged

## What Ships

- a receipts-backed `hold_escrow_for_dispute` meta tool
- transaction-level gating on funded escrow plus `dispute-ready` settlement state
- service-local per-transaction serialization so concurrent hold requests for the same transaction do not enter the hold runtime at the same time
- escrow reference resolution from canonical transaction context
- canonical `dispute_lifecycle_status = hold-active` on the transaction receipt after hold success
- dispute hold success and failure evidence in the current submission receipt trail
- tool results that include:
  - `settlement_progression_status`
  - `dispute_lifecycle_status`
  - `escrow_reference`
  - `runtime_reference`

## Current Limits

This slice does not yet include:

- a separate escrow `held` terminal state
- richer adjudication policy or scoring
- automatic hold looping distinct from the current `hold-active` / `re-escalated` lifecycle markers
- human adjudication UI
