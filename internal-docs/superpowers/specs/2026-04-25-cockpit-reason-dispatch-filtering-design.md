# Cockpit Reason / Dispatch Filtering Design

## Purpose / Scope

This design extends the landed cockpit dead-letter filter bar so operators can narrow the backlog using the two remaining high-value text filters that already exist on the backlog read model.

This slice adds:

- `dead_letter_reason_query`
- `latest_dispatch_reference`

The target is the existing cockpit dead-letter page filter bar.

This slice directly includes:

- a reason-query text input
- a dispatch-reference text input
- existing `Enter` apply semantics
- existing reload and first-row reset semantics

This slice does not directly include:

- reset or clear shortcuts
- selection preservation
- advanced filter modal UX
- dispatch picker UX
- result highlighting

## Filter Surface Extension

The current filter bar already carries:

- `query`
- `adjudication`
- `latest_status_subtype`
- `latest_status_subtype_family`
- `any_match_family`
- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`

This slice adds:

- `dead_letter_reason_query`
- `latest_dispatch_reference`

After this slice, cockpit filtering reaches near-parity with the currently landed dead-letter backlog read model on the operator-facing surface.

## Input Model

The input model stays intentionally small:

- `dead_letter_reason_query`
  - free-text input
  - same substring-style intent as the existing backlog tool surface

- `latest_dispatch_reference`
  - free-text input
  - same exact-match intent as the existing backlog tool surface

This slice does not add suggestions, pickers, or structured lookup helpers.

## Interaction / Apply Model

Interaction stays aligned with the current cockpit filter bar:

- operator edits draft state
- `Enter` applies all current filters

This slice does not change:

- field-draft semantics
- text-edit behavior
- per-filter immediate apply behavior

The goal is a small incremental extension of the current page-local filter workflow, not a new filter subsystem.

## Reload Semantics

Apply semantics remain unchanged.

After apply:

1. filtered backlog reload
2. selection reset to the first row
3. selected detail reload from that first row

This slice intentionally does not add selection preservation or stale-selection reconciliation.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add reason and dispatch draft state
  - add rendering in the existing filter bar
  - add text-edit handling for the new fields

- extend `internal/cli/cockpit/deps.go`
  - forward:
    - `dead_letter_reason_query`
    - `latest_dispatch_reference`

- extend tests:
  - `internal/cli/cockpit/pages/deadletters_test.go`
  - `internal/cli/cockpit/deps_test.go`

No new backend endpoint, page, or write path is introduced.

## Follow-On Inputs

Natural follow-on work after this slice:

1. filter UX polish
- reset / clear shortcuts
- selection preservation
- richer field navigation

2. richer matching UX
- result highlighting for reason query
- dispatch-reference copy affordance

3. broader cockpit polish
- higher-level CLI surfaces
- further operator recovery workflow improvements
