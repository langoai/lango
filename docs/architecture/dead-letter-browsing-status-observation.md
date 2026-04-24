# Dead-Letter Browsing / Status Observation

This page describes the first `dead-letter browsing / status observation` slice for post-adjudication execution in `knowledge exchange v1`.

## Purpose

This slice gives operators a read-only view into the current dead-letter backlog and the current status of a specific post-adjudication execution transaction.

The slice is intentionally narrow:

- read-only only
- transaction-centered read model
- current dead-letter backlog only
- per-transaction detail only

## What Ships

- `list_dead_lettered_post_adjudication_executions`
  - `adjudication` filter
  - `retry_attempt_min` / `retry_attempt_max` filters
  - `query` over transaction and submission receipt IDs
  - `manual_replay_actor` filter
  - `dead_lettered_after` / `dead_lettered_before` filters
  - `dead_letter_reason_query` filter
  - `latest_dispatch_reference` exact-match filter
  - `latest_status_subtype` filter
  - `manual_retry_count_min` / `manual_retry_count_max` filters
  - `total_retry_count_min` / `total_retry_count_max` filters
  - `latest_status_subtype_family` filter
  - `sort_by`
  - `offset` / `limit` pagination
  - `count` / `total` / `offset` / `limit` response metadata
  - `latest_dead_lettered_at`
  - `latest_manual_replay_actor`
  - `latest_manual_replay_at`
  - `latest_status_subtype`
  - `manual_retry_count`
  - `total_retry_count`
  - `latest_status_subtype_family`
- `get_post_adjudication_execution_status(transaction_receipt_id)`
  - current canonical snapshot
  - latest retry / dead-letter summary
  - lightweight navigation hints:
    - `is_dead_lettered`
    - `can_retry`
    - `adjudication`
- read model composed from:
  - `transaction receipt`
  - `current submission receipt`
  - `submission receipt trail`

## Current Limits

This slice does not yet include:

- replay / repair actions
- generic dead-letter browsing for all background tasks
- raw background task snapshots
- full event history dump
- cockpit / CLI presentation surfaces
- richer detail-surface actor/time summaries
- custom sort order
- multi-column sort
- dominant or any-match family grouping
- cross-submission retry aggregation
