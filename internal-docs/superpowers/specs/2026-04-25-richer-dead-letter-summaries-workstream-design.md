# Richer Dead-Letter Summaries Workstream Design

## Purpose / Scope

This workstream extends the landed dead-letter summary surface so operators can understand backlog shape more quickly.

The first slice extends the existing:

- `lango status dead-letter-summary`

with:

- top 5 latest dead-letter reasons

This workstream directly includes:

- extending the existing summary command
- top-5 latest dead-letter reason aggregation
- `table` / `json` output extension

This workstream does not directly include:

- actor breakdown
- dispatch breakdown
- cockpit summary pane
- grouped reason families
- configurable top-N flags

The goal is a richer dead-letter summary surface, not a full analytics system.

## CLI Surface

This first slice does not add a new command.

Target surface:

- `lango status dead-letter-summary`

The current summary remains:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`

This slice adds:

- `top_latest_dead_letter_reasons`

This keeps the same operator entrypoint while making the summary more informative.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- existing `list_dead_lettered_post_adjudication_executions`

The summary command:

1. reads the current dead-letter backlog through the existing list surface
2. uses `latest_dead_letter_reason` from each row
3. aggregates the top 5 latest dead-letter reasons in the CLI layer
4. renders the extended overview

This keeps the summary aligned with the same read model as the existing backlog list.

## Reason Summary Scope

The reason summary stays intentionally narrow.

Aggregation basis:

- `latest dead-letter reason`

Output shape:

- top `5` reasons
- each with:
  - `reason`
  - `count`

This means the summary reflects the distribution of each transaction's current/latest dead-letter reason rather than a full historical reason histogram.

This slice intentionally excludes:

- full reason histogram
- grouped reason families
- reason normalization pipeline
- history-wide reason accumulation

## Output Model

The current summary command contract is extended additively.

Existing fields:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`

New field:

- `top_latest_dead_letter_reasons`

The expected item shape is:

- `reason`
- `count`

The `table` view should gain one additional section, and the `json` view should gain one additional array field.

## Execution / Parallelization Model

This workstream is handled as a larger batch.

Execution model:

- one spec
- one implementation plan
- two workers in parallel

### Worker A

Owns:

- `internal/cli/status/*`
- aggregation logic
- output rendering
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
  - extend summary result types
  - aggregate top 5 latest dead-letter reasons
  - extend `table` rendering
  - extend `json` rendering
  - update tests

- update docs / OpenSpec
  - `docs/cli/status.md`
  - `docs/cli/index.md` when summary description changes
  - `README.md` only if summary surface is described there
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`
  - `openspec/specs/docs-only/spec.md`
  - one archive change for the workstream

This workstream is additive. It does not redesign the current dead-letter read model or operator recovery path.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. actor / dispatch summary work
- manual replay actor breakdown
- dispatch reference breakdown

2. cockpit summary surface
- summary pane
- grouped overview in TUI

3. broader summary evolution
- grouped reason families
- configurable top-N
- trend/time-window summaries
