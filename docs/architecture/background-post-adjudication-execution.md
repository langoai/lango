# Background Post-Adjudication Execution

This page describes the first `background post-adjudication execution` slice for `knowledge exchange v1`.

## Purpose

This slice adds an async convenience path after escrow adjudication.

The slice is intentionally narrow:

- `adjudicate_escrow_dispute` accepts optional `background_execute=true`
- successful adjudication may enqueue a background release or refund follow-up
- the background path returns a dispatch receipt, not a synchronous execution result
- release/refund still reuse the same executor gates
- dispatch success is not rolled back when worker execution later fails

## What Ships

- optional `background_execute` on `adjudicate_escrow_dispute`
- execution-mode exclusivity with `auto_execute`
- background dispatch receipt
  - `queued`
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `escrow_reference`
  - `outcome`
  - `dispatch_reference`
- existing background task substrate reuse
- existing release/refund executor reuse from the background path

## Current Limits

This slice does not yet include:

- retry orchestration
- dead-letter handling
- scheduled backoff
- status query surface specialized for post-adjudication execution
- automatic async execution as a default policy
