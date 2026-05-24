# Cockpit Reset / Clear Shortcuts Design

## Purpose / Scope

This design upgrades the landed cockpit dead-letter filter bar so operators can quickly return the page to its default unfiltered state with one shortcut.

This slice adds:

- a single global reset shortcut

The target is the existing cockpit dead-letter page.

This slice directly includes:

- `ctrl+r` as the reset shortcut
- full filter reset
- immediate reload
- confirm-state clear

This slice does not directly include:

- per-field clear
- selection preservation
- advanced filter chips
- reset confirmation prompt
- retry cancellation

## Reset Model

The reset is a full page-filter reset, not a draft-only reset.

Reset scope:

- all text-input draft fields
- all enum/toggle draft fields
- all applied filter fields

Reset result:

- the page returns to the default dead-letter backlog view with no active filters

This gives operators one predictable action for “go back to the default backlog state.”

## Running-State Guard

The reset shortcut is ignored while retry is in the `running` state.

That means:

- retry in flight
- `ctrl+r` becomes a no-op

This slice intentionally does not attempt:

- retry cancellation
- override-confirm reset
- modal warning while running

The first slice keeps the guard simple and safe.

## Reload / Selection Semantics

When reset is triggered:

1. all draft filters reset
2. all applied filters reset
3. retry confirm state clears
4. backlog reloads
5. selection resets to the first row
6. detail reloads from that first row

This keeps reset aligned with the current apply semantics rather than introducing new selection-reconciliation behavior.

## Interaction Model

Interaction is intentionally small:

- `ctrl+r`
  - if retry is not running:
    - reset filters
    - clear confirm state
    - reload page data
  - if retry is running:
    - ignore

This slice does not add:

- reset status toasts
- per-field clear shortcuts
- reset confirmation prompts

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add `ctrl+r` key handling
  - add a helper that resets all draft/applied filter state
  - clear retry confirm state
  - no-op while retry is running
  - reuse the existing backlog/detail reload path

- extend `internal/cli/cockpit/pages/deadletters_test.go`
  - cover full reset behavior
  - cover confirm-state clearing
  - cover running-state no-op

No new backend, bridge contract, or retry execution path is introduced.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer filter UX
- per-field clear
- selection preservation
- reset status messaging

2. richer operator polish
- field navigation improvements
- result highlighting

3. higher-level operator surfaces
- additional CLI views
- broader cockpit workflow polish
