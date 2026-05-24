# Actor / Time-Based Filters Design

## 1. Purpose / Scope

This design upgrades the existing dead-letter backlog list with two operator-facing filter axes:

- who requested the latest manual replay
- when the latest dead-letter happened

The goal is to make `list_dead_lettered_post_adjudication_executions` more useful for real backlog triage without widening scope into a broader observability system.

This slice directly covers:

- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`
- list response additions:
  - `latest_dead_lettered_at`
  - `latest_manual_replay_actor`

This slice does not directly cover:

- detail-view changes
- actor classes or policy joins
- reason substring filters
- alternate sort modes
- raw background task bridges

This is an **existing backlog list upgrade only**.

## 2. Filter Model

This slice adds three new filters to the existing dead-letter backlog list:

- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`

Existing filters remain in place:

- `adjudication`
- `retry_attempt_min`
- `retry_attempt_max`
- `query`
- `offset`
- `limit`

This means operators can narrow the backlog with combinations such as:

- only `refund` branch entries
- only retry attempt `>= 3`
- only entries replayed by one operator
- only entries dead-lettered after a given time

## 3. Evidence Sources

This slice does not introduce a new store. It continues to derive status from the current read model.

- `latest_manual_replay_actor`
  - read from the latest `manual-retry-requested` evidence in the `submission receipt trail`

- `latest_dead_lettered_at`
  - read from the latest `dead-lettered` evidence timestamp in the `submission receipt trail`

This keeps the current split intact:

- canonical transaction state stays on the transaction receipt
- operator-facing retry / dead-letter evidence stays on the submission trail

## 4. Response Shape

Each backlog entry now also includes:

- `latest_dead_lettered_at`
- `latest_manual_replay_actor`

Time values use:

- `RFC3339 UTC string`

Existing fields remain:

- `transaction_receipt_id`
- `submission_receipt_id`
- `adjudication`
- `latest_dead_letter_reason`
- `latest_retry_attempt`
- `latest_dispatch_reference`

This ensures operators can see why an entry matched the actor or time filter without opening another surface.

## 5. Combination Semantics

All list filters are combined with **`AND`** semantics.

That includes:

- adjudication
- retry attempt range
- query
- manual replay actor
- dead-letter time range

This keeps the first slice predictable and aligned with backlog narrowing workflows. No OR query language is added here.

## 6. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus` event summary extraction with:
  - latest dead-letter timestamp
  - latest manual replay actor
- extend the list filter matcher with:
  - actor matching
  - time-window matching
- extend the list response entry with:
  - `latest_dead_lettered_at`
  - `latest_manual_replay_actor`

The meta tool surface stays the same:

- `list_dead_lettered_post_adjudication_executions`

Only its query surface and response entry shape become richer.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. richer filters
   - dead-letter reason substring
   - dispatch reference
   - replay count

2. alternate sort modes
   - latest dead-letter time
   - latest manual replay time

3. detail surface expansion
   - richer actor/time summaries in `get_post_adjudication_execution_status`
