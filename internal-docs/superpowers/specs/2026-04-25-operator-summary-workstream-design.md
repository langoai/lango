# Operator Summary Workstream Design

## Purpose / Scope

This workstream adds the first operator-facing summary surface on top of the landed dead-letter operator views.

The first slice of this workstream adds:

- `lango status dead-letter-summary`

This workstream directly includes:

- dead-letter backlog global overview
- existing backlog read-model reuse
- default `table` output
- optional `json` output

This workstream does not directly include:

- top dead-letter reasons
- actor / dispatch breakdown
- cockpit summary pane
- a new backend summary service
- bulk operator actions

The goal is a first operator summary surface, not a full analytics layer.

## CLI Surface

This first slice adds one CLI command:

- `lango status dead-letter-summary`

The command is intentionally separate from `lango status dead-letters`.

Roles:

- `dead-letters`
  - row-oriented backlog inspection
- `dead-letter-summary`
  - overview-oriented backlog summary

This keeps list and summary concerns separate and makes operator intent clearer.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- existing `list_dead_lettered_post_adjudication_executions`

The summary command:

1. reads the current dead-letter backlog through the existing list surface
2. performs a thin CLI-side aggregation
3. renders the overview

This keeps the CLI, cockpit, and read model aligned to the same source of truth.

## Summary Scope

The first slice stays intentionally narrow.

Included summary fields:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`

This means operators can immediately see:

- how large the backlog is
- how much of it is retryable
- whether backlog pressure is concentrated in `release` vs `refund`
- whether backlog pressure is concentrated in `retry`, `manual-retry`, or `dead-letter`

This slice intentionally excludes:

- top reasons
- actor breakdown
- dispatch-reference breakdown
- time-window trend summaries
- transaction-global family rollups

## Output Model

The output model follows the current `status` CLI conventions:

- default `table`
- optional `json`

The expected summary payload shape is:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`

The first slice should keep this compact and stable.

## Execution / Parallelization Model

This workstream is handled as a larger batch.

Execution model:

- one spec
- one implementation plan
- two workers in parallel

### Worker A

Owns:

- `internal/cli/status/*`
- summary command
- aggregation logic
- CLI tests

### Worker B

Owns:

- `docs/cli/*`
- `docs/architecture/*`
- `README.md`
- `openspec/*`

This first slice does not touch cockpit code, so two workers are sufficient.

## Implementation Shape

Recommended implementation shape:

- extend `internal/cli/status`
  - add `dead-letter-summary` subcommand
  - load the existing dead-letter backlog
  - aggregate:
    - total
    - retryable count
    - adjudication buckets
    - latest-family buckets
  - render `table` / `json`

- update docs / OpenSpec
  - `docs/cli/status.md`
  - `docs/cli/index.md`
  - `README.md` when command inventory is shown
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not redesign the current dead-letter read model or operator recovery path.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. richer dead-letter summaries
- top reasons
- actor / dispatch breakdown

2. cockpit summary surface
- summary pane
- grouped overview in TUI

3. broader operator summary work
- background-task summaries
- grouped recovery summaries
