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
  - `offset` / `limit` pagination
  - `count` / `total` / `offset` / `limit` response metadata
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
- actor or time-range filters
- alternate sort modes
- cockpit / CLI presentation surfaces
