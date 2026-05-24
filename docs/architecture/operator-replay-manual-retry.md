# Operator Replay / Manual Retry

This page describes the first `operator replay / manual retry` slice for dead-lettered background post-adjudication execution in `knowledge exchange v1`.

## Purpose

This slice adds an operator-facing replay path that reuses the same recovery substrate as automatic retry and dead-letter handling.

The slice is intentionally narrow:

- only dead-lettered post-adjudication execution can be replayed
- replay requires canonical adjudication to still be present
- replay requires the current submission trail to contain canonical dead-letter evidence from the shared recovery source
- replay also requires policy authorization for the current actor and outcome
- replay reuses the existing background post-adjudication dispatch path
- prior dead-letter evidence is preserved
- `manual-retry-requested` evidence is appended

## What Ships

- a new `retry_post_adjudication_execution` meta tool
- replay service with:
  - dead-letter evidence gate
  - canonical adjudication gate
  - current submission resolution
- shared recovery policy semantics:
  - `source=post_adjudication_retry`
  - `retry-scheduled`
  - `dead-lettered`
  - `manual-retry-requested`
- append-only `manual-retry-requested` evidence including actor and timestamp when available
- new background dispatch receipt returned to the operator
- replay dispatch reuse of the same canonical prompt and retry-key substrate as background recovery

## Current Limits

This slice does not yet include:

- inline replay
- arbitrary background task replay
- per-transaction recovery policy snapshots
- broader dispute engine behavior
