# Retry / Dead-Letter Handling

This page describes the current `retry / dead-letter handling` slice for background post-adjudication execution in `knowledge exchange v1`.

## Purpose

This slice normalizes bounded retry semantics for the background post-adjudication recovery path and defines the canonical re-escalation behavior when retries exhaust.

The slice is intentionally narrow:

- only post-adjudication background execution is retried
- retry uses the existing background task substrate
- the automatic retry policy is normalized into a runtime policy unit
- retries are bounded to `3` attempts with exponential backoff from a `25ms` base delay
- exhausted retries become terminal dead-letter background failure
- canonical adjudication remains recorded
- exhausted retries re-enter settlement progression at `dispute-ready`
- exhausted retries set `dispute_lifecycle_status = re-escalated`

## What Ships

- normalized retry policy shape on the background manager:
  - `MaxRetryAttempts`
  - `BaseDelay`
  - `ShouldScheduleRetry(attempt_count)`
  - `DelayForAttempt(attempt_count)`
- retry metadata on background tasks
  - `retry_key`
  - `attempt_count`
  - `next_retry_at`
- post-adjudication retry hook on the background manager
- normalized retry identity based on:
  - `transaction_receipt_id`
  - adjudication outcome
- duplicate submissions for the same canonical `retry_key` reuse the existing pending, running, or scheduled task instead of dispatching another background run
- append-only submission receipt trail evidence under `source=post_adjudication_retry`
  - `retry-scheduled`
  - `dead-lettered`
- standardized evidence payloads:
  - retry scheduling records `attempt`, `next_retry_at`, `outcome`, and optional `dispatch_reference`
  - dead-lettering records `attempt`, `outcome`, `dead_lettered_at`, and the terminal failure reason
- canonical dead-letter re-escalation:
  - keeps `escrow_adjudication`
  - sets `settlement_progression_status = dispute-ready`
  - sets `settlement_progression_reason_code = escalate`
  - sets `settlement_progression_reason = post-adjudication execution dead-lettered`
  - sets `dispute_lifecycle_status = re-escalated`
- panic in the background runner fails the task explicitly and keeps the event visible as task failure rather than leaving an orphaned running task

## Current Limits

This slice does not yet include:

- operator-editable retry tuning
- wider non-post-adjudication adoption of the retry policy shape
- a generic recovery substrate for arbitrary background task families
