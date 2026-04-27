# Per-Submission Breakdown Design

## 1. Purpose / Scope

This design adds the first compact per-submission breakdown to the dead-letter backlog row.

The goal is to let operators see, within one backlog row:

- which submissions belong to the transaction
- how many retry-lifecycle events each submission accumulated
- which retry-lifecycle families each submission touched

This slice directly covers:

- backlog row field:
  - `submission_breakdown`
- breakdown item fields:
  - `submission_receipt_id`
  - `retry_count`
  - `any_match_families`

This slice does not directly cover:

- per-submission dominant family
- per-submission timestamps
- breakdown-specific filters
- separate detail tool
- current marker decoration

This remains an **existing backlog list row only** slice.

## 2. Breakdown Model

`submission_breakdown` is derived from:

- all submissions that belong to the current transaction

For each submission, the slice computes:

- `retry_count`
  - count of relevant retry-lifecycle events on that submission trail
- `any_match_families`
  - deduplicated family set observed on that submission trail

Relevant events:

- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

Family mapping remains:

- `retry-scheduled` -> `retry`
- `manual-retry-requested` -> `manual-retry`
- `dead-lettered` -> `dead-letter`

This makes the breakdown a compact explanation layer under the existing transaction-global aggregates.

## 3. Ordering Model

This slice orders `submission_breakdown` as:

- `oldest -> newest`

The reason is straightforward:

- submission history is easier to read as lifecycle progression in chronological order

This slice does not add:

- current submission pinning
- newest-first ordering
- explicit current markers

## 4. Response Shape

This slice adds:

- `submission_breakdown`

Each item includes:

- `submission_receipt_id`
- `retry_count`
- `any_match_families`

Illustrative shape:

```json
[
  {
    "submission_receipt_id": "sub-1",
    "retry_count": 2,
    "any_match_families": ["retry", "manual-retry"]
  },
  {
    "submission_receipt_id": "sub-2",
    "retry_count": 1,
    "any_match_families": ["dead-letter"]
  }
]
```

This slice deliberately keeps the breakdown compact and does not add timestamps or dominant-family fields yet.

## 5. Computation Model

This slice uses **on-read computation**.

Computation steps:

1. take the row's `transaction_receipt_id`
2. enumerate all submissions that belong to that transaction
3. read each submission trail
4. count relevant retry-lifecycle events per submission
5. derive a deduplicated family set per submission
6. sort submission summaries `oldest -> newest`

This slice does not add:

- aggregation cache
- breakdown precomputation
- separate submission-history index beyond what already exists

## 6. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus` with:
  - `SubmissionBreakdownItem`
    - `SubmissionReceiptID`
    - `RetryCount`
    - `AnyMatchFamilies`
  - `DeadLetterBacklogEntry.SubmissionBreakdown []SubmissionBreakdownItem`
- reuse the existing transaction submission scan helper
- add a compact per-submission summary helper
- populate `submission_breakdown` during backlog row assembly

No new tools or stores are introduced.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. richer per-submission breakdown
   - dominant family
   - timestamps
   - current submission marker

2. breakdown-aware filters
   - submission retry count
   - submission any-match family

3. separate detail surface
   - richer submission timeline view
