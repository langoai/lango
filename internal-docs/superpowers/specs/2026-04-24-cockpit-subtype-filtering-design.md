# Cockpit Subtype Filtering Design

## Purpose / Scope

This design extends the landed cockpit dead-letter filter bar so operators can narrow the backlog by its latest retry/dead-letter phase.

The slice adds:

- `latest_status_subtype`

to the existing thin filter bar.

This slice directly includes:

- subtype enum toggle on the existing filter bar
- reuse of the existing draft/apply model
- existing reload + first-row reset semantics

This slice does not directly include:

- actor/time filters
- family filters
- live filtering
- selection preservation
- advanced filter forms

## Filter Surface Extension

The filter bar now contains three controls.

1. `query`
- text input

2. `adjudication`
- small enum toggle
- `all`
- `release`
- `refund`

3. `latest_status_subtype`
- small enum toggle
- `all`
- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

These subtype values match the existing landed backlog read model exactly.

## Interaction Model

Interaction stays the same as the current filter bar:

- edit draft filter state
- press `Enter`
- apply the current draft

This slice does not add:

- immediate apply on subtype change
- live filtering
- separate apply button

## Reload / Selection Semantics

Apply behavior remains unchanged:

1. reload the filtered backlog
2. reset selection to the first row
3. reload the detail pane from that first row

No selection preservation is introduced in this slice.

## Data Source Reuse

This slice continues to reuse the existing read surfaces:

- `list_dead_lettered_post_adjudication_executions`
- `get_post_adjudication_execution_status`

The cockpit filter bar simply forwards:

- `query`
- `adjudication`
- `latest_status_subtype`

to the existing backlog list surface.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add subtype draft state
  - add subtype toggle rendering
  - add subtype key handling
- extend the cockpit dead-letter bridge
  - forward `latest_status_subtype`
- keep apply semantics unchanged
- update cockpit page tests and bridge tests

This remains an incremental page-local filter-state extension.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer cockpit filters
- actor/time
- family filters

2. better filtering UX
- selection preservation
- reset/clear shortcuts
- live filtering

3. richer recovery feedback
- loading state
- failure detail
- replay result presentation
