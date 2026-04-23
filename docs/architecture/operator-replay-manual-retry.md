# Operator Replay / Manual Retry

This page describes the first `operator replay / manual retry` slice for dead-lettered background post-adjudication execution in `knowledge exchange v1`.

## Purpose

This slice adds an operator-facing replay path for dead-lettered post-adjudication execution.

The slice is intentionally narrow:

- only dead-lettered post-adjudication execution can be replayed
- replay requires canonical adjudication to still be present
- replay reuses the existing background post-adjudication dispatch path
- prior dead-letter evidence is preserved
- `manual-retry-requested` evidence is appended

## What Ships

- a new `retry_post_adjudication_execution` meta tool
- replay service with:
  - dead-letter evidence gate
  - canonical adjudication gate
  - current submission resolution
- append-only `manual-retry-requested` evidence
- new background dispatch receipt returned to the operator

## Current Limits

This slice does not yet include:

- inline replay
- arbitrary background task replay
- generic dead-letter queue browsing UI
- richer replay policy
- broader dispute engine behavior
