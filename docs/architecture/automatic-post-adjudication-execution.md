# Automatic Post-Adjudication Execution

This page describes the first `automatic post-adjudication execution` slice for `knowledge exchange v1`.

## Purpose

This slice adds an inline execution mode after escrow adjudication.

The slice is intentionally narrow:

- `adjudicate_escrow_dispute` accepts optional `auto_execute=true`
- explicit execution flags win over runtime defaults
- when neither execution flag is set, the runtime defaults to `manual_recovery`
- successful adjudication may immediately call the existing release or refund executor
- release/refund still go through the same executor gates
- adjudication success is never rolled back if nested execution fails
- no new lifecycle state is introduced

## What Ships

- optional `auto_execute` on `adjudicate_escrow_dispute`
- shared execution-mode resolution:
  - `auto_execute=true` => `inline`
  - `background_execute=true` => `background`
  - omitted flags => `manual_recovery`
  - `auto_execute` and `background_execute` are mutually exclusive
- inline handler orchestration
  - `release` adjudication routes to `release_escrow_settlement`
  - `refund` adjudication routes to `refund_escrow_settlement`
- combined return shape
  - adjudication result
  - nested execution result when requested

## Current Limits

This slice does not yet include:

- config-backed non-manual default selection
- policy editing for execution-mode defaults
- richer dispute engine behavior
