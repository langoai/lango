# Dead-Letter Browsing / Status Observation Design

## Purpose / Scope

This design defines the first `dead-letter browsing / status observation` slice for post-adjudication execution in `knowledge exchange v1`.

Its job is narrow:

- surface the current dead-letter backlog for post-adjudication execution
- expose a per-transaction status view for operators
- provide enough read-only context to decide whether replay is needed

This slice covers:

- read-only meta tools
  - `list_dead_lettered_post_adjudication_executions`
  - `get_post_adjudication_execution_status`
- transaction-centered read model
- current dead-letter backlog view
- per-transaction status view

This slice does not cover:

- replay or repair actions
- generic dead-letter browsing for all background tasks
- full event history dump
- raw background task snapshot surface
- richer filtering or pagination design

## Read Model

The basic unit for this slice is `transaction`.

The operator-facing status model is composed from:

- `transaction receipt`
- `current submission receipt`
- `submission receipt trail`

This keeps the read model aligned with the same canonical/evidence substrate already used by adjudication, retry, and replay.

Important rule:

- background-task raw state is not the primary source of truth in this first slice
- transaction receipt owns current canonical state
- submission receipt trail owns append-only failure/retry evidence

## List Surface

First tool:

- `list_dead_lettered_post_adjudication_executions`

This tool returns **only transactions that are currently dead-lettered**.

It does not return:

- every transaction that has ever had dead-letter history
- every post-adjudication background execution attempt

Minimum response fields:

- `transaction_receipt_id`
- `submission_receipt_id`
- `adjudication`
- `latest_dead_letter_reason`
- `latest_retry_attempt`
- `latest_dispatch_reference`

This gives operators the minimum backlog view needed to decide what needs attention.

## Detail Surface

Second tool:

- `get_post_adjudication_execution_status(transaction_receipt_id)`

This tool returns:

- current canonical snapshot
- latest retry / dead-letter summary

The first slice intentionally excludes:

- full receipt trail dump
- raw background task snapshot dump

So the detail view is still bounded:

- current adjudication
- current settlement progression
- current escrow execution status
- latest retry/dead-letter summary

## Canonical Sources

This slice reads from:

1. `transaction receipt`
   - canonical adjudication
   - current settlement progression
   - current escrow execution status
   - current submission pointer

2. `submission receipt trail`
   - dead-letter evidence
   - retry-scheduled evidence
   - manual-retry evidence
   - related post-adjudication execution evidence

This lets the operator-facing view combine:

- current state
- latest failure and retry story

without introducing a new status store.

## Implementation Shape

Recommended structure:

- read-only status service
- `list_dead_lettered_post_adjudication_executions` meta tool
- `get_post_adjudication_execution_status` meta tool

The service should:

- inspect transaction receipts
- follow the current submission pointer
- summarize the latest dead-letter and retry evidence from the submission trail

This slice is strictly read-only:

- no canonical state mutation
- no replay dispatch
- no executor invocation

## Follow-On Inputs

The next follow-on work after this slice is:

1. `richer list / query controls`
   - filtering
   - pagination
   - grouping by branch or outcome

2. `raw background status bridge`
   - decide whether background-task metadata should later be surfaced directly

3. `operator cockpit / CLI surface`
   - expose this read model in higher-level operator interfaces
