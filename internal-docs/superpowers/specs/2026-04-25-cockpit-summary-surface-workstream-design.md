# Cockpit Summary Surface Workstream Design

## Purpose / Scope

This workstream adds the first cockpit-native summary surface on top of the landed dead-letter cockpit page.

The first slice adds:

- a compact summary strip at the top of the dead-letters page

This workstream directly includes:

- page-top summary strip
- reuse of existing dead-letter backlog rows
- global overview only
- refresh aligned with backlog reloads

This workstream does not directly include:

- top reasons
- top actors
- top dispatch references
- a separate summary pane
- a separate summary page
- a new backend summary service

The goal is the smallest useful cockpit summary surface, not a full summary dashboard.

## Surface Placement

This first slice does not add a new page or pane.

Placement:

- top of the `dead-letters` cockpit page
- above or directly adjacent to the existing filter bar

This keeps the current master-detail layout mostly unchanged while giving operators an immediate overview when they enter the page.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- the existing dead-letter backlog rows already loaded by the cockpit page

The page:

1. loads the backlog rows through the existing cockpit dead-letter path
2. computes a page-local aggregate
3. renders the compact summary strip

This keeps the strip aligned with the same data already driving the table and detail pane.

## Summary Scope

The first slice stays intentionally narrow.

Included summary fields:

- `total dead letters`
- `retryable count`
- `by adjudication`
- `by latest family`

This means operators can immediately see:

- current backlog size
- how much of it is retryable
- `release` vs `refund` distribution
- `retry`, `manual-retry`, and `dead-letter` distribution

This slice intentionally excludes:

- top reasons
- top actors
- top dispatch references
- trend summaries

## Refresh Semantics

The summary strip must refresh whenever the backlog rows change.

Recompute triggers:

- initial page load
- filter apply
- `Ctrl+R` reset
- retry-success refresh

This keeps:

- backlog rows
- summary strip
- selected detail

on the same state transition timeline.

## UI Model

The first slice uses a single-line compact-chip style.

Example shape:

- `dead letters: 12`
- `retryable: 7`
- `release/refund: 9/3`
- `retry/manual/dead: 2/4/6`

This keeps vertical space low and fits the existing cockpit density better than cards or a mini-table.

## Execution / Parallelization Model

This workstream is handled as a larger batch.

Execution model:

- one spec
- one implementation plan
- two workers in parallel

### Worker A

Owns:

- `internal/cli/cockpit/*`
- summary aggregation
- strip rendering
- cockpit tests

### Worker B

Owns:

- `docs/architecture/*`
- `README.md` when necessary
- `openspec/*`

This first slice is cockpit-focused, so two workers are sufficient.

## Implementation Shape

Recommended implementation shape:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add a page-local summary aggregation helper
  - compute summary from the current backlog rows
  - render the compact summary strip above the existing page content

- extend `internal/cli/cockpit/pages/deadletters_test.go`
  - cover summary rendering
  - cover recompute after backlog reload paths

- update docs / OpenSpec
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not redesign the current cockpit dead-letter page or backend read model.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. richer cockpit summary
- top reasons
- top actors
- top dispatch references

2. broader summary surfaces
- separate summary pane
- separate summary page

3. operator summary parity
- richer CLI / cockpit summary alignment
