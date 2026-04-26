# Operator Surface Consolidation Workstream Design

## Purpose / Scope

This workstream consolidates the remaining dead-letter operator-surface follow-up work into one larger batch so the project can stop paying repeated micro-slice overhead and move faster toward runtime and reputation work.

The target surfaces are:

- `lango status dead-letter-summary`
- `lango status dead-letters`
- `lango status dead-letter retry <transaction-receipt-id>`
- cockpit `dead-letters` page

This workstream directly includes:

- grouped dispatch-family summaries
- dead-letter CLI `any_match_family`
- summary evolution for richer top-N and trend / time-window views
- recovery follow-up UX
  - polling
  - follow-up refresh
  - richer structured retry results

This workstream does not directly include:

- generic async retry policy redesign
- policy-driven defaults
- replay substrate normalization
- dispute runtime completion
- reputation model changes

The goal is to materially reduce the remaining operator-surface backlog in one workstream so the next major stage can shift to runtime and dispute-core work.

## Current Baseline

The landed operator surface already includes:

- dead-letter CLI list / detail / retry
- dead-letter CLI summary
- cockpit dead-letter master-detail surface
- grouped latest reason-family summaries
- grouped latest actor-family summaries
- raw top latest reasons / actors / dispatch references
- retry confirm / running / success / failure UX
- selection preservation and reset behavior

The remaining gaps are now concentrated rather than broad:

- grouped dispatch families
- dead-letter CLI `any_match_family`
- richer summary controls and trend views
- richer post-retry follow-up UX

This makes a consolidation workstream feasible because the remaining tasks all live in the operator layer and mostly share CLI / cockpit / docs boundaries.

## Workstream Scope

This workstream is intentionally broad but still operator-surface only.

### 1. Grouped Dispatch Families

Add grouped dispatch-family summaries across:

- CLI summary output
- cockpit top summary strip

The first slice should follow the same additive model as grouped reason and actor families:

- keep raw top latest dispatch references
- add grouped dispatch-family buckets

The taxonomy can remain small and heuristic in the first implementation.

### 2. Dead-Letter CLI `any_match_family`

Extend `lango status dead-letters` with:

- `--any-match-family`

This closes the explicit CLI filter parity gap called out in current docs and OpenSpec.

The first implementation should:

- validate supported values
- reuse the existing dead-letter list bridge
- keep detail and retry surfaces unchanged

### 3. Summary Evolution

Add the next summary-operator improvements without redesigning the whole surface.

In scope:

- richer top-N support
- trend / time-window summaries

Recommended first implementation direction:

- keep current summary commands and strip in place
- make top-N richer without inventing a new command tree
- add a compact trend / time-window representation that can be read in both CLI and cockpit

This should remain additive rather than replacing the existing overview surfaces.

### 4. Recovery Follow-Up UX

Extend the retry follow-up experience after request acceptance.

In scope:

- polling
- follow-up refresh
- richer structured retry results

CLI direction:

- preserve the current request flow
- add optional follow-up polling / structured result refinement after request acceptance

Cockpit direction:

- preserve the current retry action semantics
- make follow-up refresh and status interpretation more legible after request acceptance

This remains surface-level UX work, not retry substrate redesign.

## Architectural Shape

This workstream should preserve the current layering.

### CLI Boundary

Primary ownership:

- `internal/cli/status/*`

This boundary should own:

- dead-letter list filter parity
- summary output model evolution
- retry follow-up CLI UX

### Cockpit Boundary

Primary ownership:

- `internal/cli/cockpit/*`

This boundary should own:

- page-top summary strip evolution
- retry follow-up cockpit UX
- dead-letter page rendering / interaction changes

### Shared Operator Helpers

Shared operator-summary logic should live in focused internal helpers when duplication becomes real.

Likely candidates:

- dispatch-family classifier helper
- shared summary-bucket ordering helpers
- trend / time-window aggregation helpers

Avoid creating a premature general analytics subsystem. Add shared helpers only where CLI and cockpit would otherwise drift.

### Docs / OpenSpec Boundary

Primary ownership:

- `docs/cli/*`
- `docs/architecture/*`
- `README.md`
- `openspec/*`

Docs should continue to describe only behavior that is actually wired in the codebase.

## Execution / Parallelization Model

This workstream should be executed as a single larger batch rather than more micro-workstreams.

Execution model:

- one spec
- one implementation plan
- three to four workers in parallel

### Worker A

Owns:

- `internal/cli/status/*`
- dead-letter CLI `any_match_family`
- CLI summary evolution
- CLI retry follow-up UX
- CLI tests

### Worker B

Owns:

- `internal/cli/cockpit/*`
- cockpit summary evolution
- cockpit retry follow-up UX
- cockpit tests

### Worker C

Owns:

- `docs/cli/*`
- `docs/architecture/*`
- `README.md`
- `openspec/*`

### Worker D (optional)

Owns:

- shared operator helper extraction
- focused implementation review
- summary-classifier or trend helper consolidation

Worker D should be used only if duplication or review pressure justifies it.

## Implementation Strategy

This workstream should be implemented in one plan with several grouped task bands, not as separate workstream artifacts.

Recommended task bands:

1. shared operator helper additions needed by both CLI and cockpit
2. CLI list filter parity (`any_match_family`)
3. CLI summary evolution and retry follow-up UX
4. cockpit summary evolution and retry follow-up UX
5. docs / OpenSpec truth alignment
6. final integrated verification

This sequence preserves useful dependency order while still allowing broad parallel execution inside the workstream.

## Completion Criteria

The workstream is complete when:

- CLI supports `--any-match-family`
- grouped dispatch-family summaries are visible where raw dispatch summaries already exist
- summary evolution for richer top-N and trend / time-window views is landed in the chosen additive form
- retry follow-up UX is richer in both CLI and cockpit
- docs / README / OpenSpec reflect the landed operator behavior
- `go build ./...`
- `go test ./...`
- `.venv/bin/zensical build`
- docs-only OpenSpec validation
all pass

## Follow-On Inputs

Natural follow-on work after this workstream:

1. Replay / Recovery Policy Runtime Workstream
- policy-driven defaults
- generic async retry policy
- replay / recovery substrate normalization

2. Dispute Runtime Completion Workstream
- keep-hold / re-escalation
- broader dispute engine completion
- richer settlement progression
- escrow lifecycle completion

3. Reputation V2 + Runtime Integration Workstream
- reputation model v2
- trust-entry contract strengthening
- deeper runtime integration
