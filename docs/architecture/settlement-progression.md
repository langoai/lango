# Settlement Progression

This page describes the first transaction-level settlement progression slice for `knowledge exchange v1`.

## Purpose

Settlement progression turns artifact release outcomes into canonical transaction-level state before a full settlement executor or dispute engine exists.

The slice is intentionally narrow:

- `transaction receipt` owns canonical settlement progression state
- `submission receipt` contributes evidence, reasons, and hints
- release approval outcomes map into progression states
- actual money movement remains separate

## What Ships

- transaction-level settlement progression state
- release outcome mapping for `approve`, `request-revision`, `reject`, and `escalate`
- canonical `approved-for-settlement` and `review-needed` progression
- `partial_hint` plumbing for bounded partial-settlement guidance
- `dispute-ready` opening rules tied to review-path disagreements
- a receipts-backed `apply_settlement_progression` meta tool

## Canonical State

The current progression states are:

- `pending`
- `in-progress`
- `review-needed`
- `approved-for-settlement`
- `partially-settled`
- `settled`
- `dispute-ready`

`transaction receipt` keeps the canonical state, reason code, human-readable reason, partial-settlement hint, and dispute-ready marker.

## Current Limits

This slice does not yet include:

- a settlement executor
- actual release or refund execution
- partial-settlement calculation formulas
- dispute orchestration
- human adjudication UI

The current implementation closes the control-plane gap first, not the full settlement lifecycle.
