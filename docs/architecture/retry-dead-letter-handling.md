# Retry / Dead-Letter Handling

This page describes the first `retry / dead-letter handling` slice for background post-adjudication execution in `knowledge exchange v1`.

## Purpose

This slice adds bounded retry semantics to the background post-adjudication path.

The slice is intentionally narrow:

- only post-adjudication background execution is retried
- retry uses the existing background task substrate
- retries are bounded to `3` attempts with exponential backoff
- exhausted retries become terminal dead-letter background failure
- canonical adjudication remains unchanged

## What Ships

- retry metadata on background tasks
  - `retry_key`
  - `attempt_count`
  - `next_retry_at`
- post-adjudication retry hook on the background manager
- append-only submission receipt trail evidence for:
  - retry scheduled
  - dead-lettered
- retry identity based on:
  - `transaction_receipt_id`
  - adjudication outcome

## Current Limits

This slice does not yet include:

- operator replay or manual retry UI
- generic background-manager-wide retry policy
- dead-letter queue browsing surface
- scheduled backoff configuration by policy
