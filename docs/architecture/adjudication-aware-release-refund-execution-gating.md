# Adjudication-Aware Release/Refund Execution Gating

This page describes the first `adjudication-aware release/refund execution gating` slice for `knowledge exchange v1`.

## Purpose

This slice connects canonical escrow adjudication to the existing release and refund executors.

The slice is intentionally narrow:

- `release_escrow_settlement` now requires `escrow_adjudication = release`
- `refund_escrow_settlement` now requires `escrow_adjudication = refund`
- adjudication success also moves settlement progression atomically
- opposite-branch execution evidence blocks the executor
- automatic post-adjudication execution is still out of scope

## What Ships

- atomic adjudication write in the receipt store
  - adjudication field update
  - progression transition
  - adjudication evidence append
- stricter escrow release gate
  - funded escrow
  - `approved-for-settlement`
  - `escrow_adjudication = release`
- stricter escrow refund gate
  - funded escrow
  - `review-needed`
  - `escrow_adjudication = refund`
- one-way branch safety against opposite-branch execution evidence

## Current Limits

This slice does not yet include:

- automatic release or refund after adjudication
- keep-hold or re-escalation states
- broader dispute engine behavior
- human adjudication UI
