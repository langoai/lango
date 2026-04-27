# Richer Cockpit Summary Workstream Design

## Purpose / Scope

This workstream extends the landed cockpit dead-letter summary strip so operators can understand backlog character immediately on page entry.

The first slice extends the existing dead-letters page top summary strip with:

- top 5 latest dead-letter reasons

This workstream directly includes:

- existing summary-strip extension
- latest dead-letter reason top-5 aggregation
- strip-level compact presentation

This workstream does not directly include:

- top actors
- top dispatch references
- a separate summary pane
- a separate summary page
- grouped reason families
- configurable top-N

The goal is cockpit summary strip v2, not a full cockpit analytics surface.

## Surface Placement

This first slice does not change placement.

Placement:

- the existing page-top summary strip on the `dead-letters` page

The global overview chips remain. The new reason summary is added additively to that strip rather than moving summary content into a separate pane or page.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- the current dead-letter backlog rows already loaded by the cockpit page

The page:

1. loads the backlog rows through the existing cockpit dead-letter path
2. performs a page-local aggregation for latest reasons
3. renders the richer strip

This keeps the strip aligned with the same rows that drive the table and detail pane.

## Reason Summary Scope

The reason summary stays intentionally narrow.

Aggregation basis:

- `latest dead-letter reason`

Output shape:

- top `5` reasons
- each with:
  - `reason`
  - `count`

This means the cockpit strip reflects the distribution of each transaction's current/latest dead-letter reason rather than a full historical histogram.

This slice intentionally excludes:

- full reason histogram
- grouped reason families
- history-wide accumulation

## UI Model

The strip must remain compact.

Recommended shape:

- keep the existing global overview chips
- add a second compact line for top reasons

Example:

- `reasons: worker exhausted(6), invalid receipt(3), timeout(2) ...`

This keeps summary density high without turning the strip into a large table or card layout.

## Refresh Semantics

The richer strip shares the existing summary-strip refresh model.

Recompute triggers:

- initial page load
- filter apply
- `Ctrl+R` reset
- retry-success refresh

This keeps:

- backlog rows
- global overview strip
- top reasons strip
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
  - add top latest dead-letter reason aggregation
  - extend summary-strip rendering

- extend `internal/cli/cockpit/pages/deadletters_test.go`
  - cover reason aggregation
  - cover strip rendering after reload paths

- update docs / OpenSpec
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not redesign the current cockpit dead-letter page or backend read model.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. richer cockpit summary
- top actors
- top dispatch references

2. broader summary surfaces
- separate summary pane
- separate summary page

3. operator summary parity
- richer CLI / cockpit summary alignment
