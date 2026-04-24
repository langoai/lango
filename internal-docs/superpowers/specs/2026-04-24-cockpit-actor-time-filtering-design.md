# Cockpit Actor / Time Filtering Design

## Purpose / Scope

This design extends the landed cockpit dead-letter filter bar so operators can filter the backlog by the latest manual replay actor and the latest dead-letter time window.

The slice adds:

- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`

to the existing filter bar.

This slice directly includes:

- actor text input
- time-range text inputs
- reuse of the existing draft/apply model
- reuse of the current reload + first-row reset semantics

This slice does not directly include:

- actor picker
- date/time widget
- live filtering
- selection preservation
- advanced filter modal
- family filters

## Filter Surface Extension

The cockpit filter bar keeps the current controls:

- `query`
- `adjudication`
- `latest_status_subtype`

and adds:

- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`

This lets the operator narrow the backlog by:

- receipt ID
- branch
- latest subtype
- latest manual replay actor
- latest dead-letter time window

## Input Model

The new inputs are intentionally simple:

- `manual_replay_actor`
  - free-text input
- `dead_lettered_after`
  - RFC3339 string text input
- `dead_lettered_before`
  - RFC3339 string text input

No picker or date widget is introduced in this slice.

## Interaction / Apply Model

The interaction model stays the same:

- edit draft filter state
- press `Enter`
- apply all current filters

This slice does not add:

- live filtering
- separate apply controls
- reset/clear shortcuts

## Reload Semantics

Apply behavior remains unchanged:

1. reload the filtered backlog
2. reset selection to the first row
3. reload the detail pane from that first row

No selection preservation is introduced.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add actor/time draft state
  - render the new inputs in the existing filter bar
  - handle text editing for the new fields
- extend the cockpit dead-letter bridge
  - forward:
    - `manual_replay_actor`
    - `dead_lettered_after`
    - `dead_lettered_before`
- keep current apply/reload semantics unchanged
- update cockpit page tests and bridge tests

This remains an incremental filter-bar extension rather than a new filtering subsystem.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer cockpit filters
- family filters
- reason/dispatch filters

2. better filtering UX
- selection preservation
- reset/clear shortcuts
- richer invalid-time feedback

3. richer recovery feedback
- loading state
- failure detail
- replay result presentation
