# Any-Match Family Grouping Design

## 1. Purpose / Scope

This design adds the first lightweight family-grouping slice to the dead-letter backlog list.

This slice introduces:

- `any_match_families`
- `any_match_family`

The goal is to help operators understand which retry-lifecycle family buckets a current submission has touched, and to filter the backlog by one family at a time.

This slice directly covers:

- row field:
  - `any_match_families`
- filter:
  - `any_match_family`

This slice does not directly cover:

- `dominant_family`
- multi-select family query
- family counts
- detail-view expansion
- cross-submission grouping

This remains an **existing backlog list only** slice.

## 2. Grouping Model

`any_match_families` is derived from the **current submission trail only**.

Relevant events:

- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

Subtype-to-family mapping:

- `retry-scheduled` -> `retry`
- `manual-retry-requested` -> `manual-retry`
- `dead-lettered` -> `dead-letter`

The output shape is a deduplicated string array.

Example:

- `["retry", "manual-retry", "dead-letter"]`

This means the row reports which retry-lifecycle families were observed at least once in the current submission trail.

## 3. Filter Model

This slice adds one new filter:

- `any_match_family`

Its meaning is simple:

- if the derived `any_match_families` set contains the requested family, the row matches

This filter remains **AND-composed** with the existing backlog filters.

## 4. Response Shape

This slice adds one row field:

- `any_match_families`

Shape:

- deduplicated string array

The reason is straightforward:

- if the filter is available, operators should also see the grouped family context in the row

This slice does not add:

- family count maps
- dominant family
- multi-family query helpers

## 5. Evidence Source

This slice continues to use the **current submission receipt trail only**.

It does not aggregate across:

- previous submissions
- transaction-global history
- background metadata outside the current submission trail

This preserves the current backlog surface boundary.

## 6. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus.RetryDeadLetterSummary` with:
  - `AnyMatchFamilies []string`
- extend `internal/postadjudicationstatus.DeadLetterBacklogEntry` with:
  - `AnyMatchFamilies []string`
- extend `internal/postadjudicationstatus.DeadLetterListOptions` with:
  - `AnyMatchFamily string`
- extend event summary extraction to:
  - walk relevant retry lifecycle events
  - map subtype to family
  - collect a deduplicated set
- extend the list filter matcher with membership matching for `AnyMatchFamily`
- extend the existing list meta tool to pass the new filter through

No new tools or stores are introduced.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. `dominant_family`
   - latest or most-frequent family semantics

2. richer family query
   - multi-select family filter
   - family count map

3. broader grouping
   - cross-submission family grouping
   - transaction-global lifecycle grouping
