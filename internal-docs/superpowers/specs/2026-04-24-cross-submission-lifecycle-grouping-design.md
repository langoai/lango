# Cross-Submission Lifecycle Grouping Design

## 1. Purpose / Scope

This design extends the dead-letter backlog list with a small transaction-history aggregation layer.

This slice introduces:

- `transaction_global_total_retry_count`
- `transaction_global_any_match_families`
- matching filters for both

The goal is to let operators see a compact transaction-wide retry-lifecycle signal without changing canonical transaction state.

This slice directly covers:

- row fields:
  - `transaction_global_total_retry_count`
  - `transaction_global_any_match_families`
- filters:
  - `transaction_global_total_retry_count_min`
  - `transaction_global_total_retry_count_max`
  - `transaction_global_any_match_family`

This slice does not directly cover:

- transaction-global dominant family
- per-submission breakdown
- time-aware lifecycle timeline
- canonical transaction mutations
- aggregation cache

This is a **read-only transaction-history aggregation** slice.

## 2. Aggregation Model

This slice aggregates over:

- all submission receipts that belong to the current transaction

That means:

- current submission included
- historical submissions included
- submissions from other transactions excluded

Relevant events remain:

- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

Family mapping remains:

- `retry-scheduled` -> `retry`
- `manual-retry-requested` -> `manual-retry`
- `dead-lettered` -> `dead-letter`

Computed values:

- `transaction_global_total_retry_count`
  - total relevant retry lifecycle events across all submission trails in the transaction
- `transaction_global_any_match_families`
  - deduplicated family set observed across all submission trails in the transaction

## 3. Filter Model

This slice adds:

- `transaction_global_total_retry_count_min`
- `transaction_global_total_retry_count_max`
- `transaction_global_any_match_family`

Their meaning is straightforward:

- total count range:
  - lower and upper bounds on transaction-global retry lifecycle count
- any-match family:
  - the requested family must exist in the derived transaction-global family set

These filters remain **AND-composed** with the existing backlog filters.

## 4. Response Shape

This slice adds:

- `transaction_global_total_retry_count`
- `transaction_global_any_match_families`

Shape:

- total count = integer
- any-match families = deduplicated string array

These fields exist so the new filters are visible in the row and do not require extra interpretation by operators.

## 5. Computation Model

This first slice uses **on-read aggregation**.

Computation steps:

1. take the current row's `transaction_receipt_id`
2. find all submission receipts belonging to that transaction
3. read each submission trail
4. count relevant retry lifecycle events
5. collect deduplicated family set

This slice does not add:

- aggregation cache
- background precomputation
- transaction-to-submission index beyond what already exists

## 6. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus.DeadLetterBacklogEntry` with:
  - `TransactionGlobalTotalRetryCount`
  - `TransactionGlobalAnyMatchFamilies`
- extend `internal/postadjudicationstatus.DeadLetterListOptions` with:
  - `TransactionGlobalTotalRetryCountMin`
  - `TransactionGlobalTotalRetryCountMax`
  - `TransactionGlobalAnyMatchFamily`
- extend the receipts access path used by the status service to scan submissions for a transaction
- aggregate relevant retry lifecycle evidence across those submission trails
- extend the list filter matcher with:
  - global count range
  - global family membership
- extend the existing list meta tool to pass the new filters through

No new tools or stores are introduced.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. richer transaction-global lifecycle view
   - dominant family
   - family counts
   - latest family across transaction

2. per-submission breakdown
   - show which submission contributed what

3. cached/indexed aggregation
   - if transaction submission fanout grows large
