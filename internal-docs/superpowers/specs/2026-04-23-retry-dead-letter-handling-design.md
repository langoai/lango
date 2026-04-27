# Retry / Dead-Letter Handling Design

## Purpose / Scope

This design defines the first `retry / dead-letter handling` slice for background post-adjudication execution in `knowledge exchange v1`.

Its job is narrow:

- retry background post-adjudication execution on worker failure
- apply bounded exponential backoff
- stop automatic retries after a fixed retry budget
- record terminal dead-letter failure when retries are exhausted

This slice covers:

- the post-adjudication background worker wrapper
- retry scheduling
- exponential backoff
- dead-letter terminal failure
- retry and dead-letter evidence

This slice does not cover:

- generic background-manager-wide retry
- queue-wide dead-letter infrastructure
- operator replay UI
- broader recovery engine behavior
- dispute engine behavior changes

## Retry Identity

The canonical retry identity for this slice is:

- `transaction_receipt_id`
- `adjudication outcome`

This means:

- `release` and `refund` retries are distinct for the same transaction
- `task_id` is an execution-attempt identifier, not the canonical retry identity

That keeps retries tied to the business branch rather than to one transient background task instance.

## Retry Policy

The first-slice retry policy is:

- maximum `3` retries
- exponential backoff

This means the worker wrapper may schedule successive attempts with progressively larger delays, but it stops automatically after the configured retry budget is exhausted.

Retry bookkeeping lives in existing background-task metadata, not in canonical transaction state.

Minimum metadata expectations:

- `attempt_count`
- `next_retry_at`
- `transaction_receipt_id`
- `outcome`

## Dead-Letter Semantics

In this first slice, dead-letter means:

- automatic retry has been exhausted
- the async execution path is now a terminal background failure

Dead-letter does **not** mean:

- dispute escalation
- settlement progression rollback
- refund/release branch reversal

Canonical adjudication remains intact. The background execution path simply stops retrying automatically.

## Evidence Model

This slice records evidence in two layers:

1. background-task metadata
   - attempt count
   - next retry timing
   - terminal dead-letter status

2. submission receipt trail
   - retry scheduled
   - retry failed
   - retry exhausted / dead-lettered
   - eventual success if a later retry succeeds

Important rule:

- retry success does not erase prior failure evidence
- evidence remains append-only

This allows later reconstruction of:

- how many attempts failed
- when retries were scheduled
- when the execution ultimately succeeded
- or when it became dead-lettered

## Implementation Shape

Recommended structure:

- background dispatch
- retrying worker wrapper
- existing release/refund services

This means:

- the existing background manager is not turned into a generic retry engine yet
- a post-adjudication-specific worker wrapper owns retry count, exponential backoff, and dead-letter semantics
- actual execution continues to reuse existing `release_escrow_settlement` and `refund_escrow_settlement` services

This keeps scope limited to the post-adjudication path while preserving existing service layering.

## Follow-On Inputs

The next follow-on work after this slice is:

1. `operator replay / manual retry`
   - allow dead-lettered execution to be re-run explicitly

2. `generic async execution policy`
   - decide whether retry behavior should later generalize beyond post-adjudication execution

3. `status observation surface`
   - add a clearer query surface for retry progress and dead-letter state
