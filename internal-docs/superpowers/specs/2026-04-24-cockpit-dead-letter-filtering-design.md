# Cockpit Dead-Letter Filtering Design

## Purpose / Scope

This design extends the landed cockpit dead-letter master-detail read surface with a first usable filtering step.

The slice adds a thin filter bar above the backlog table with:

- `query`
- `adjudication`
- `Enter` to apply

The goal is to let operators narrow the backlog without introducing a full filter form or any write actions.

This slice directly covers:

- cockpit filter bar
- `query` text input
- `adjudication` toggle
- filtered backlog reload
- first-row reset semantics
- detail reload after apply

This slice does not directly cover:

- live filtering
- selection preservation
- subtype / actor / time filters
- replay / write controls
- advanced filter UI

## Filter Surface Model

The filter bar contains two controls.

1. `query`
- free-text input
- used for transaction/submission receipt ID filtering

2. `adjudication`
- small enum toggle
- values:
  - `all`
  - `release`
  - `refund`

This is the smallest useful filter subset for the cockpit dead-letter page.

## Interaction Model

The interaction stays intentionally simple.

- user edits filter draft state
- user presses `Enter`
- page applies the current draft

This slice does not add:

- typing-as-you-filter
- explicit apply button
- clear chips
- modal filter dialogs

## Reload / Selection Semantics

When the user applies filters:

1. reload the filtered backlog
2. reset selection to the first row
3. reload the detail pane for that first row

If the filtered result is empty:

- show an empty backlog state
- clear the selected detail pane

This slice intentionally does not attempt selection preservation.

## Data Source Reuse

The cockpit page continues to reuse the existing meta-tool-backed read surfaces.

Backlog source:

- `list_dead_lettered_post_adjudication_executions`

Detail source:

- `get_post_adjudication_execution_status`

The cockpit filter bar only forwards:

- `query`
- `adjudication`

to the existing list surface.

No new backend endpoint or cockpit-specific read API is introduced.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add local filter draft state
    - `query`
    - `adjudication`
  - add key handling for:
    - text editing
    - adjudication toggle
    - `Enter` apply
- extend the injected list loader to accept filter params
- on apply:
  - reload backlog
  - reset cursor to the first row
  - reload detail

This remains a page-local state machine extension on top of the already landed cockpit dead-letter page.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer cockpit filters
- `latest_status_subtype`
- actor/time filters
- family filters

2. better interaction semantics
- selection preservation
- live filtering
- reset/clear shortcuts

3. cockpit actions
- replay / repair controls
