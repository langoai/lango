# Background Post-Adjudication Execution

This page describes the first `background post-adjudication execution` slice for `knowledge exchange v1`.

## Purpose

This slice adds an async follow-up mode after escrow adjudication.

The slice is intentionally narrow:

- `adjudicate_escrow_dispute` accepts optional `background_execute=true`
- explicit execution flags still control the immediate mode choice
- when both execution flags are omitted, the runtime keeps only the canonical adjudication and defaults to `manual_recovery`
- successful adjudication may enqueue a background release or refund follow-up
- the background path returns a dispatch receipt, not a synchronous execution result
- release/refund still reuse the same executor gates
- dispatch success is not rolled back when worker execution later fails

## What Ships

- optional `background_execute` on `adjudicate_escrow_dispute`
- unified execution-mode policy for post-adjudication follow-up:
  - `manual_recovery`
  - `inline`
  - `background`
- execution-mode exclusivity with `auto_execute`
- background dispatch receipt
  - `queued`
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `escrow_reference`
  - `outcome`
  - `dispatch_reference`
- shared background dispatch prompt built from canonical adjudication state
- existing background task substrate reuse
- existing release/refund executor reuse from the background path

## Current Limits

This slice does not yet include:

- config-backed non-manual default selection
- operator-editable execution-mode policy
- background execution for arbitrary recovery flows outside post-adjudication follow-up
