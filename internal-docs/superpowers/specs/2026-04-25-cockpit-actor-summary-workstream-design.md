# Cockpit Actor Summary Workstream Design

## Purpose / Scope

This workstream extends the landed cockpit top summary strip so operators can quickly see who is most associated with recent manual replay activity on page entry.

The first slice extends the existing dead-letters page top strip with:

- top 5 latest manual replay actors

This workstream directly includes:

- existing strip extension
- latest manual replay actor top-5 aggregation
- compact third-line presentation

This workstream does not directly include:

- top dispatch references
- a separate summary pane
- a separate summary page
- grouped actor families
- configurable top-N
- history-wide actor accumulation

The goal is cockpit summary strip v3, not a full cockpit analytics surface.

## Surface Placement

This first slice does not change placement.

Placement:

- the existing page-top summary strip on the `dead-letters` page

The strip becomes:

1. global overview
2. `reasons: ...`
3. `actors: ...`

This keeps the current master-detail layout intact while increasing summary value.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- the current dead-letter backlog rows already loaded by the cockpit page

The page:

1. loads the backlog rows through the existing cockpit dead-letter path
2. performs a page-local aggregation for latest manual replay actors
3. renders the richer strip

This keeps the strip aligned with the same rows that drive the table and detail pane.

## Actor Summary Scope

The actor summary stays intentionally narrow.

Aggregation basis:

- `latest manual replay actor`

Output shape:

- top `5` actors
- each with:
  - `actor`
  - `count`

This means the cockpit strip reflects the distribution of each transaction's current/latest manual replay actor rather than a full historical actor histogram.

This slice intentionally excludes:

- full actor histogram
- grouped actor families
- history-wide actor accumulation

## UI Model

The strip must remain compact.

Recommended shape:

- first line:
  - global overview chips
- second line:
  - `reasons: ...`
- third line:
  - `actors: ...`

Example:

- `actors: operator:alice(4), operator:bob(2) ...`

This keeps the strip dense without turning it into a pane or a mini-table.

## Refresh Semantics

The actor summary shares the existing summary-strip refresh model.

Recompute triggers:

- initial page load
- filter apply
- `Ctrl+R` reset
- retry-success refresh

This keeps:

- backlog rows
- global overview strip
- top reasons strip
- top actors strip
- selected detail

on the same state transition timeline.

## Execution / Parallelization Model

This workstream is handled as a larger batch.

Execution model:

- one spec
- one implementation plan
- two workers in parallel

### Worker A

Owns:

- `internal/cli/cockpit/*`
- aggregation logic
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
  - extend page-local summary aggregation
  - add top latest manual replay actor aggregation
  - extend summary-strip rendering with a third line

- extend `internal/cli/cockpit/pages/deadletters_test.go`
  - cover actor aggregation
  - cover strip rendering after reload paths

- update docs / OpenSpec
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not redesign the current cockpit dead-letter page or backend read model.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. cockpit dispatch summary
- top latest dispatch references

2. broader summary evolution
- grouped reason / actor / dispatch families
- configurable top-N
- trend/time-window summaries

3. operator summary parity
- richer CLI / cockpit summary alignment
