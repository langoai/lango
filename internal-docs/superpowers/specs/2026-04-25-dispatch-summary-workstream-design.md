# Dispatch Summary Workstream Design

## Purpose / Scope

This workstream extends the landed dead-letter summary surface so operators can quickly see which dispatch references are most associated with the current dead-letter backlog.

The first slice extends the existing:

- `lango status dead-letter-summary`

with:

- top 5 latest dispatch references

This workstream directly includes:

- extending the existing summary command
- latest dispatch reference top-5 aggregation
- `table` / `json` output extension

This workstream does not directly include:

- grouped dispatch families
- time-window trend views
- cockpit summary pane
- configurable top-N flags

The goal is a dispatch-focused summary slice, not a full analytics or trend system.

## CLI Surface

This first slice does not add a new command.

Target surface:

- `lango status dead-letter-summary`

The current summary remains:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`
- `top_latest_dead_letter_reasons`
- `top_latest_manual_replay_actors`

This slice adds:

- `top_latest_dispatch_references`

This keeps the same operator entrypoint while making the summary more informative.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- existing `list_dead_lettered_post_adjudication_executions`

The summary command:

1. reads the current dead-letter backlog through the existing list surface
2. uses `latest_dispatch_reference` from each row
3. aggregates the top 5 latest dispatch references in the CLI layer
4. renders the extended overview

This keeps the summary aligned with the same read model as the existing backlog list.

## Dispatch Summary Scope

The dispatch summary stays intentionally narrow.

Aggregation basis:

- `latest dispatch reference`

Output shape:

- top `5` dispatch references
- each with:
  - `dispatch_reference`
  - `count`

This means the summary reflects the distribution of each transaction's current/latest dispatch reference rather than a full historical dispatch histogram.

This slice intentionally excludes:

- grouped dispatch families
- trend/time-window summaries
- dispatch normalization pipeline
- history-wide dispatch accumulation

## Output Model

The current summary command contract is extended additively.

Existing fields:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`
- `top_latest_dead_letter_reasons`
- `top_latest_manual_replay_actors`

New field:

- `top_latest_dispatch_references`

The expected item shape is:

- `dispatch_reference`
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
  - aggregate top 5 latest dispatch references
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

1. cockpit summary surface
- summary pane
- grouped overview in TUI

2. broader summary evolution
- grouped reason / actor / dispatch families
- configurable top-N
- trend/time-window summaries

3. broader operator summary work
- background-task summaries
- grouped recovery summaries
