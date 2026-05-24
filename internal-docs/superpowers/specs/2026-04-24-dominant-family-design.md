# Dominant Family Design

## 1. Purpose / Scope

This design adds a single dominant-family signal to the current dead-letter backlog list.

This slice introduces:

- `dominant_family`
- `dominant_family` filter

The goal is to give operators a compact summary of which retry-lifecycle family most strongly characterizes the current submission trail.

This slice directly covers:

- row field:
  - `dominant_family`
- filter:
  - `dominant_family`

This slice does not directly cover:

- family count maps
- weighted dominance
- cross-submission dominance
- detail-view expansion

This remains an **existing backlog list only** slice.

## 2. Dominance Model

`dominant_family` is derived from the **current submission trail only**.

Relevant events:

- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

Subtype-to-family mapping:

- `retry-scheduled` -> `retry`
- `manual-retry-requested` -> `manual-retry`
- `dead-lettered` -> `dead-letter`

Decision rule:

1. the family with the highest count wins
2. if counts tie, the family of the latest relevant event wins

This gives the first slice a simple and operator-friendly summary rule:

- count-first
- latest-event tie-break

## 3. Filter Model

This slice adds one filter:

- `dominant_family`

Its meaning is straightforward:

- if the derived `dominant_family` equals the requested family, the row matches

This filter remains **AND-composed** with the existing backlog filters.

## 4. Response Shape

This slice adds one row field:

- `dominant_family`

The reason is simple:

- if the filter exists, operators should also see the derived family directly in the row

This slice does not add:

- family counts
- dominance score
- weighted breakdown

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
  - `DominantFamily`
- extend `internal/postadjudicationstatus.DeadLetterBacklogEntry` with:
  - `DominantFamily`
- extend `internal/postadjudicationstatus.DeadLetterListOptions` with:
  - `DominantFamily`
- extend event summary extraction to:
  - count family hits across relevant events
  - apply latest-event tie-break
- extend the list filter matcher with exact matching for `DominantFamily`
- extend the existing list meta tool to pass the new filter through

No new tools or stores are introduced.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. richer family aggregation
   - family count maps
   - weighted dominance
   - dominance score

2. broader scope
   - cross-submission dominance
   - transaction-global lifecycle dominance

3. grouped operator surface
   - family-grouped cockpit / CLI views
