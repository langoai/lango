# Escrow Release

This page describes the first `escrow release` slice for `knowledge exchange v1`.

## Purpose

Escrow release connects a funded escrow and an `approved-for-settlement` transaction to real settlement completion.

The slice is intentionally narrow:

- release only
- canonical input is `transaction_receipt_id`
- escrow must already be `funded`
- settlement progression must already be `approved-for-settlement`
- success closes progression to `settled`
- failure keeps progression at `approved-for-settlement`

## What Ships

- a receipts-backed `release_escrow_settlement` meta tool
- transaction-level gating on funded escrow plus approved settlement state
- amount resolution from canonical transaction context
- existing escrow runtime reuse for release
- success and failure evidence in the current submission receipt trail

## Current Limits

This slice does not yet include:

- refund
- dispute-linked escrow handling
- milestone-aware escrow release
- human settlement UI
