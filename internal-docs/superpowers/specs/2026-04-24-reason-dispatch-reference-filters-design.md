# Reason / Dispatch Reference Filters Design

## 1. Purpose / Scope

This design extends the current dead-letter backlog list with two more triage-oriented filters:

- dead-letter reason substring
- dispatch reference exact match

The goal is to make `list_dead_lettered_post_adjudication_executions` more useful for operator backlog narrowing without widening the scope into a larger observability system.

This slice directly covers:

- `dead_letter_reason_query`
- `latest_dispatch_reference`

This slice does not directly cover:

- detail view changes
- response field additions
- subtype filters
- replay count filters
- sort-mode expansion

This remains an **existing backlog list only** slice.

## 2. Filter Model

This slice adds two filters to the existing dead-letter backlog list:

- `dead_letter_reason_query`
- `latest_dispatch_reference`

Existing filters remain in place:

- `adjudication`
- `retry_attempt_min`
- `retry_attempt_max`
- `query`
- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`
- `offset`
- `limit`

Their meaning is intentionally narrow:

- `dead_letter_reason_query`
  - substring match against `latest_dead_letter_reason`
- `latest_dispatch_reference`
  - exact match against `latest_dispatch_reference`

This keeps the first slice operator-focused and predictable.

## 3. Matching Semantics

Matching rules for the new filters:

- `dead_letter_reason_query`
  - case-insensitive substring match
  - only against `latest_dead_letter_reason`

- `latest_dispatch_reference`
  - exact match
  - dispatch references are identifier-like and should not use substring semantics here

Like the existing filters, these are combined with **`AND`** semantics.

That means:

- adjudication
- retry attempt range
- query
- actor
- time range
- dead-letter reason query
- dispatch reference

must all match when they are supplied together.

## 4. Response Shape

This slice does **not** add new response fields.

The reason is simple:

- `latest_dead_letter_reason`
- `latest_dispatch_reference`

already exist on each backlog entry.

So this slice expands the query surface only. The list page shape and entry shape remain unchanged.

## 5. Implementation Shape

Recommended implementation:

- extend `internal/postadjudicationstatus.DeadLetterListOptions` with:
  - `DeadLetterReasonQuery`
  - `LatestDispatchReference`
- extend the existing list filter matcher with:
  - reason substring matching
  - dispatch reference exact matching
- extend `list_dead_lettered_post_adjudication_executions` to pass the new filter inputs into the status service

This means:

- no new store
- no new response type
- no new meta tool

Only the current read model and list tool are extended.

## 6. Follow-On Inputs

Natural follow-on work after this slice:

1. richer backlog filters
   - status subtype
   - replay count
   - latest retry scheduled window

2. alternate sort modes
   - latest dead-letter time
   - latest manual replay time
   - latest dispatch time

3. higher-level operator surface
   - cockpit / CLI dead-letter views
   - grouped triage presentation
