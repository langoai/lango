# Transaction-Global Dominant Family Design

## 1. Purpose / Scope

This design adds a single dominant-family signal across the full submission history of a transaction.

This slice introduces:

- `transaction_global_dominant_family`
- `transaction_global_dominant_family` filter

The goal is to give operators a compact transaction-wide family summary without expanding into a full aggregation model.

This slice directly covers:

- row field:
  - `transaction_global_dominant_family`
- filter:
  - `transaction_global_dominant_family`

This slice does not directly cover:

- family count maps
- weighted dominance
- per-submission breakdown expansion
- detail-view expansion

This remains an **existing backlog list only** slice.

## 2. Dominance Model

This slice aggregates across:

- all submission receipts that belong to the current transaction

Relevant events:

- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

Subtype-to-family mapping remains:

- `retry-scheduled` -> `retry`
- `manual-retry-requested` -> `manual-retry`
- `dead-lettered` -> `dead-letter`

Dominance rule:

1. highest count wins
2. if counts tie, the family of the latest relevant event wins

This mirrors the local dominant-family rule, but uses transaction-global submission scope.

## 3. Filter Model

This slice adds one filter:

- `transaction_global_dominant_family`

Its meaning is straightforward:

- if the derived transaction-global dominant family equals the requested family, the row matches

This filter remains **AND-composed** with the existing backlog filters.

## 4. Response Shape

This slice adds one row field:

- `transaction_global_dominant_family`

The field exists so that operators can directly see the same value they filter on.

This slice does not add:

- family count maps
- weighted dominance scores
- dominance explanations

## 5. Evidence Source

This slice continues to use:

- all submission receipts belonging to the current transaction
- each submission's `submission receipt trail`

This slice does not add:

- aggregation cache
- background precomputation
- background metadata joins

Computation stays on-read.

## 6. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus.DeadLetterBacklogEntry` with:
  - `TransactionGlobalDominantFamily`
- extend `internal/postadjudicationstatus.DeadLetterListOptions` with:
  - `TransactionGlobalDominantFamily`
- extend the transaction-global aggregation helper to:
  - count family hits across all submissions in the transaction
  - apply latest-event tie-break
- extend the list filter matcher with exact matching for `TransactionGlobalDominantFamily`
- extend the existing list meta tool to pass the new filter through

No new tools or stores are introduced.

## 7. Follow-On Inputs

Natural follow-on work after this slice:

1. richer transaction-global family model
   - family count maps
   - weighted dominance
   - dominance explanation

2. per-submission breakdown
   - how each submission contributes to the aggregate

3. operator presentation
   - grouped cockpit views
   - CLI summary tables
