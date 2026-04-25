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
  - `transaction_global_total_retry_count_min` / `transaction_global_total_retry_count_max` filters
  - `transaction_global_any_match_family` filter
  - `transaction_global_dominant_family` filter
  - `latest_status_subtype_family` filter
  - `any_match_family` filter
  - `dominant_family` filter
  - `sort_by`
  - `offset` / `limit` pagination
  - `count` / `total` / `offset` / `limit` response metadata
  - `latest_dead_lettered_at`
  - `latest_manual_replay_actor`
  - `latest_manual_replay_at`
  - `latest_status_subtype`
  - `manual_retry_count`
  - `total_retry_count`
  - `transaction_global_total_retry_count`
  - `transaction_global_any_match_families`
  - `transaction_global_dominant_family`
  - `submission_breakdown`
  - `latest_status_subtype_family`
  - `any_match_families`
  - `dominant_family`
- `get_post_adjudication_execution_status(transaction_receipt_id)`
  - current canonical snapshot
  - latest retry / dead-letter summary
  - optional `latest_background_task`
    - `task_id`
    - `status`
    - `attempt_count`
    - `next_retry_at`
  - lightweight navigation hints:
    - `is_dead_lettered`
    - `can_retry`
    - `adjudication`
- read model composed from:
  - `transaction receipt`
  - `current submission receipt`
  - `submission receipt trail`
- cockpit read surface
  - dead-letter backlog table
  - selected transaction detail pane
  - selection-driven detail refresh
  - reuses the existing list/detail status surfaces
  - thin filter bar
    - `query`
    - `adjudication`
    - `latest_status_subtype`
    - `latest_status_subtype_family`
    - `any_match_family`
    - `manual_replay_actor`
    - `dead_lettered_after`
    - `dead_lettered_before`
    - `dead_letter_reason_query`
    - `latest_dispatch_reference`
    - `Enter` apply
    - first-row reset after reload
  - detail-pane `Retry` action
    - reuses `retry_post_adjudication_execution`
    - enabled only when `can_retry = true`
    - `r` key binding
    - success/failure status message only
    - first `r` enters inline confirm state
    - second `r` executes replay
    - `Esc`, selection change, and filter apply clear confirm state
    - while replay is in flight, `Retry action` renders `running...`
    - duplicate retry triggers are blocked while replay is running
    - replay failure surfaces the backend error string and returns the action to idle
    - replay success refreshes backlog and selected detail

## Current Limits

This slice does not yet include:

- replay / repair actions
- generic dead-letter browsing for all background tasks
- full event history dump
- richer detail-surface actor/time summaries
- selection preservation after filter changes
- higher-level CLI surfaces
