# Replay-Count / Subtype Filters + Alternate Sort Modes Design

## 1. Purpose / Scope

This design upgrades the current dead-letter backlog list to make it more useful for operator triage.

This slice adds two groups of capabilities:

- new filters
  - `latest_status_subtype`
  - `manual_retry_count_min`
  - `manual_retry_count_max`
- alternate sort modes
  - `latest_dead_lettered_at`
  - `latest_retry_attempt`
  - `latest_manual_replay_at`

The target remains the same read-only surface:

- `list_dead_lettered_post_adjudication_executions`

This slice directly covers:

- replay-count filtering
- subtype filtering
- sort-mode input
- list entry field additions:
  - `manual_retry_count`
  - `latest_manual_replay_at`
  - `latest_status_subtype`

This slice does not directly cover:

- detail view expansion
- custom sort order
- multi-column sort
- total retry count
- subtype families
- raw background-task bridges

This is still a **single backlog-list upgrade**.

## 2. Filter Model

This slice adds three filters to the existing backlog list:

- `latest_status_subtype`
- `manual_retry_count_min`
- `manual_retry_count_max`

Their meaning is intentionally narrow:

- `latest_status_subtype`
  - exact match against the latest retry/dead-letter-related status subtype
- `manual_retry_count_min`
  - lower bound on `manual-retry-requested` event count
- `manual_retry_count_max`
  - upper bound on `manual-retry-requested` event count

All filters remain **AND-composed** with the current list filters.

That means:

- adjudication
- retry attempt range
- query
- actor
- time range
- dead-letter reason
- dispatch reference
- latest status subtype
- manual retry count range

must all match when provided together.

## 3. Sort Model

This slice adds a single `sort_by` input.

Allowed values:

- `latest_dead_lettered_at`
- `latest_retry_attempt`
- `latest_manual_replay_at`

Direction is fixed per mode:

- `latest_dead_lettered_at desc`
- `latest_retry_attempt desc`
- `latest_manual_replay_at desc`

This slice does not add:

- `sort_order`
- multi-column sort
- custom comparator composition

The goal is to expose the most useful operator orderings first without overcomplicating the query surface.

## 4. Response Shape

This slice adds these list-entry fields:

- `manual_retry_count`
- `latest_manual_replay_at`
- `latest_status_subtype`

The reason is straightforward:

- filters should be visible in the row
- sort keys should be visible in the row

Otherwise operators would need to reverse-engineer why an entry matched or why it appears in a given order.

Existing fields remain unchanged.

## 5. Evidence Sources

This slice does not introduce a new store. It continues to derive triage data from the `submission receipt trail`.

- `manual_retry_count`
  - count of `manual-retry-requested` events
- `latest_manual_replay_at`
  - timestamp of the latest `manual-retry-requested` event
- `latest_status_subtype`
  - the latest retry/dead-letter-related status subtype

This preserves the current model:

- transaction receipt = canonical state
- submission receipt trail = operator-facing retry/dead-letter evidence

## 6. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus.DeadLetterListOptions` with:
  - `LatestStatusSubtype`
  - `ManualRetryCountMin`
  - `ManualRetryCountMax`
  - `SortBy`
- extend `RetryDeadLetterSummary` with:
  - `ManualRetryCount`
  - `LatestManualReplayAt`
- extend `DeadLetterBacklogEntry` with:
  - `ManualRetryCount`
  - `LatestManualReplayAt`
  - `LatestStatusSubtype`
- extend event summary extraction to:
  - count manual retry events
  - track the latest manual retry timestamp
  - track the latest relevant subtype
- extend list sorting to switch on `sort_by`
- extend the existing list meta tool to pass through the new filters and `sort_by`

No new tools or stores are introduced.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. richer filters
   - total retry count
   - subtype families
   - latest retry-scheduled window

2. richer sorting
   - custom `sort_order`
   - multi-column sort
   - grouped sort by adjudication/subtype

3. higher-level operator surface
   - cockpit dead-letter table
   - CLI list/detail presentation
