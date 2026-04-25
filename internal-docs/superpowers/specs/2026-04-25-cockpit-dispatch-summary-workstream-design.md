# Cockpit Dispatch Summary Workstream Design

## Purpose / Scope

This workstream extends the landed cockpit top summary strip so operators can quickly see which dispatch references are most associated with the current dead-letter backlog on page entry.

The first slice extends the existing dead-letters page top strip with:

- top 5 latest dispatch references

This workstream directly includes:

- existing strip extension
- latest dispatch reference top-5 aggregation
- compact fourth-line presentation

This workstream does not directly include:

- grouped dispatch families
- trend/time-window views
- a separate summary pane
- a separate summary page
- configurable top-N

The goal is cockpit summary strip v4, not a full cockpit analytics surface.

## Surface Placement

This first slice does not change placement.

Placement:

- the existing page-top summary strip on the `dead-letters` page

The strip becomes:

1. global overview
2. `reasons: ...`
3. `actors: ...`
4. `dispatch: ...`

This keeps the current master-detail layout intact while increasing summary value.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- the current dead-letter backlog rows already loaded by the cockpit page

The page:

1. loads the backlog rows through the existing cockpit dead-letter path
2. performs a page-local aggregation for latest dispatch references
3. renders the richer strip

This keeps the strip aligned with the same rows that drive the table and detail pane.

## Dispatch Summary Scope

The dispatch summary stays intentionally narrow.

Aggregation basis:

- `latest dispatch reference`

Output shape:

- top `5` dispatch references
- each with:
  - `dispatch_reference`
  - `count`

This means the cockpit strip reflects the distribution of each transaction's current/latest dispatch reference rather than a full historical dispatch histogram.

This slice intentionally excludes:

- grouped dispatch families
- trend/time-window summaries
- history-wide dispatch accumulation

## UI Model

The strip must remain compact.

Recommended shape:

- first line:
  - global overview chips
- second line:
  - `reasons: ...`
- third line:
  - `actors: ...`
- fourth line:
  - `dispatch: ...`

Example:

- `dispatch: dispatch-a(4), dispatch-b(2) ...`

This keeps the strip dense without turning it into a pane or a mini-table.

## Refresh Semantics

The dispatch summary shares the existing summary-strip refresh model.

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
- top dispatch strip
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
  - add top latest dispatch reference aggregation
  - extend summary-strip rendering with a fourth line

- extend `internal/cli/cockpit/pages/deadletters_test.go`
  - cover dispatch aggregation
  - cover strip rendering after reload paths

- update docs / OpenSpec
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not redesign the current cockpit dead-letter page or backend read model.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. broader summary evolution
- grouped reason / actor / dispatch families
- configurable top-N
- trend/time-window summaries

2. broader cockpit summary surfaces
- separate summary pane
- separate summary page

3. operator summary parity
- richer CLI / cockpit summary alignment
