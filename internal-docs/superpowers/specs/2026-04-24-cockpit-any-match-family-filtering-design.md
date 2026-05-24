# Cockpit Any-Match Family Filtering Design

## Purpose / Scope

This design extends the landed cockpit dead-letter filter bar so operators can narrow the backlog by whether a retry lifecycle family has appeared at least once.

The slice adds:

- `any_match_family`

to the existing filter bar.

This slice directly includes:

- any-match family enum toggle
- reuse of the existing draft/apply model
- reuse of the current reload + first-row reset semantics

This slice does not directly include:

- multi-select family filtering
- live filtering
- selection preservation
- advanced filter modal
- richer family grouping

## Filter Surface Extension

The cockpit filter bar keeps the current controls:

- `query`
- `adjudication`
- `latest_status_subtype`
- `latest_status_subtype_family`
- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`

and adds:

- `any_match_family`

Supported values:

- `all`
- `retry`
- `manual-retry`
- `dead-letter`

These values match the landed `any_match_family` model in the dead-letter backlog surface.

## Interaction Model

The interaction model remains unchanged:

- edit draft filter state
- press `Enter`
- apply all current filters

This slice does not add:

- immediate apply on family change
- live filtering
- reset/clear controls
- advanced filter modal
- multi-select chips

## Reload / Selection Semantics

Apply behavior remains unchanged:

1. reload the filtered backlog
2. reset selection to the first row
3. reload the detail pane from that first row

No selection preservation is introduced.

## Data Source Reuse

This slice continues to reuse the existing read surfaces:

- `list_dead_lettered_post_adjudication_executions`
- `get_post_adjudication_execution_status`

The cockpit filter bar simply forwards:

- `query`
- `adjudication`
- `latest_status_subtype`
- `latest_status_subtype_family`
- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`
- `any_match_family`

to the existing backlog list surface.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add any-match family draft state
  - render the any-match family toggle in the existing filter bar
  - handle any-match family key input
- extend the cockpit dead-letter bridge
  - forward `any_match_family`
- keep existing apply/reload semantics unchanged
- update cockpit page tests and bridge tests

This remains an incremental filter-bar extension, not a new filtering subsystem.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer family filtering
- multi-select family filters
- broader grouping

2. better filtering UX
- selection preservation
- reset/clear shortcuts
- live filtering

3. richer recovery feedback
- loading state
- failure detail
- replay result presentation
