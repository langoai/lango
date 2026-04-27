# Richer Filtering / Pagination Design

## Purpose / Scope

This design defines the first `richer filtering / pagination` slice for the existing dead-letter browsing and post-adjudication status observation surface in `knowledge exchange v1`.

Its job is narrow:

- add practical filtering to the dead-letter backlog list
- add offset/limit pagination plus total count
- add lightweight navigation hints to the detail view

This slice covers:

- `list_dead_lettered_post_adjudication_executions`
  - outcome filter
  - retry attempt range filter
  - text query
  - offset and limit
  - total count
- `get_post_adjudication_execution_status`
  - lightweight navigation hints

This slice does not cover:

- cursor pagination
- actor filter
- time-range filter
- raw background task snapshots
- full event history dump
- user-selectable sort keys

## List Filter Model

The list surface adds three filter families:

- `adjudication outcome`
- `retry attempt range`
- `text query`
  - `transaction_receipt_id`
  - `submission_receipt_id`

Behavior:

- if filters are omitted, the current dead-letter backlog is returned as-is
- filters are optional, not mandatory

This lets operators quickly narrow the backlog without requiring a more complex query language.

## Pagination Model

The first-slice pagination model is:

- `offset`
- `limit`

The list response should include:

- `entries`
- `count`
- `total`
- `offset`
- `limit`

This keeps pagination simple while still letting operators understand the size of the current backlog.

## Sorting Model

The first-slice sort order is fixed:

- `latest_retry_attempt desc`
- tie-breaker: `transaction_receipt_id`

This prioritizes transactions that have already failed more times, which is a reasonable default triage order for dead-letter backlog review.

## Detail Navigation Hints

The detail surface adds lightweight navigation hints:

- `is_dead_lettered`
- `can_retry`
- `adjudication`

These are convenience fields only. They do not create a new canonical state layer.

They allow operators to see quickly:

- whether the transaction is currently dead-lettered
- whether the transaction is currently replay-eligible
- which adjudication branch is canonical

## Canonical Sources

The slice continues to read from the existing receipts substrate:

- `transaction receipt`
  - canonical adjudication
  - current settlement progression
  - current escrow execution state
  - current submission pointer
- `submission receipt trail`
  - dead-letter evidence
  - retry-scheduled evidence
  - manual-retry evidence

No new state store is introduced for filtering or pagination.

## Implementation Shape

Recommended structure:

- extend `internal/postadjudicationstatus` service
- add list query input:
  - outcome
  - retry attempt min/max
  - query
  - offset
  - limit
- add list response metadata:
  - total
  - count
  - offset
  - limit
- add detail response hints:
  - dead-lettered
  - retryable
  - adjudication

The existing meta tool names remain unchanged. Only the read model becomes richer.

## Follow-On Inputs

The next follow-on work after this slice is:

1. `richer filters`
   - actor
   - time range
   - reason substring

2. `alternate sort modes`
   - latest dead-letter time
   - latest dispatch time
   - adjudication branch grouping

3. `higher-level operator surface`
   - cockpit / TUI
   - CLI list and detail presentation
