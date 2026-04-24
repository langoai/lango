# Total Retry-Count / Subtype-Family Filters Design

## 1. Purpose / Scope

This design upgrades the current dead-letter backlog list so operators can reason about retry lifecycle density and the current subtype family more quickly.

This slice adds two concepts:

- `total_retry_count`
- `latest_status_subtype_family`

The goal is to make the existing backlog list easier to triage without expanding into a broader observability system.

This slice directly covers:

- filters:
  - `total_retry_count_min`
  - `total_retry_count_max`
  - `latest_status_subtype_family`
- response fields:
  - `total_retry_count`
  - `latest_status_subtype_family`

This slice does not directly cover:

- cross-submission aggregation
- dominant family
- any-match family
- detail view expansion
- custom grouping UI

This remains an **existing backlog list only** slice.

## 2. Filter Model

This slice adds three filters:

- `total_retry_count_min`
- `total_retry_count_max`
- `latest_status_subtype_family`

Their meaning is intentionally narrow:

- `total_retry_count_min`
  - lower bound on relevant retry lifecycle events in the current submission trail
- `total_retry_count_max`
  - upper bound on relevant retry lifecycle events in the current submission trail
- `latest_status_subtype_family`
  - exact match against the family of the latest retry/dead-letter subtype

All filters remain **AND-composed** with the existing backlog filters.

## 3. Family Mapping Model

This slice maps latest subtype to family like this:

- `retry-scheduled` -> `retry`
- `manual-retry-requested` -> `manual-retry`
- `dead-lettered` -> `dead-letter`

This slice does not compute:

- dominant family
- any-match family
- multi-family tagging

Only the family of the latest subtype is exposed.

## 4. Response Shape

This slice adds these backlog row fields:

- `total_retry_count`
- `latest_status_subtype_family`

These fields exist so that the new filters are visible in the response and do not require reverse interpretation by operators.

Existing fields remain unchanged.

## 5. Evidence Source

This slice continues to use the **current submission receipt trail only**.

`total_retry_count` counts relevant `post_adjudication_retry` events on the current submission:

- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

This slice does not aggregate:

- previous submissions
- transaction-global retry totals
- background metadata outside the current submission trail

## 6. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus.RetryDeadLetterSummary` with:
  - `TotalRetryCount`
  - `LatestStatusSubtypeFamily`
- extend `internal/postadjudicationstatus.DeadLetterBacklogEntry` with:
  - `TotalRetryCount`
  - `LatestStatusSubtypeFamily`
- extend `internal/postadjudicationstatus.DeadLetterListOptions` with:
  - `TotalRetryCountMin`
  - `TotalRetryCountMax`
  - `LatestStatusSubtypeFamily`
- extend event summary extraction to:
  - count relevant retry lifecycle events
  - map latest subtype to family
- extend the list filter matcher to:
  - apply total retry count range
  - apply family exact match
- extend the existing list meta tool to pass the new inputs through

No new tools or stores are introduced.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. richer lifecycle grouping
   - dominant family
   - any-match family
   - cross-submission aggregation

2. richer counts
   - transaction-global retry count
   - submission-local vs transaction-global dual counts

3. grouped operator surface
   - family-grouped cockpit view
   - CLI grouping / pivot presentation
