# Actor / Dispatch Summary Workstream Design

## Purpose / Scope

This workstream extends the landed dead-letter summary surface so operators can quickly see who is most associated with recent manual replay activity.

The first slice extends the existing:

- `lango status dead-letter-summary`

with:

- top 5 latest manual replay actors

This workstream directly includes:

- extending the existing summary command
- latest manual replay actor top-5 aggregation
- `table` / `json` output extension

This workstream does not directly include:

- dispatch breakdown
- cockpit summary pane
- grouped actor families
- configurable top-N flags

The goal is an actor-focused richer summary slice, not a full operator analytics system.

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

This slice adds:

- `top_latest_manual_replay_actors`

This keeps the same operator entrypoint while making the summary more informative.

## Data Source Reuse

This slice does not introduce a new backend summary path.

Source:

- existing `list_dead_lettered_post_adjudication_executions`

The summary command:

1. reads the current dead-letter backlog through the existing list surface
2. uses `latest_manual_replay_actor` from each row
3. aggregates the top 5 latest manual replay actors in the CLI layer
4. renders the extended overview

This keeps the summary aligned with the same read model as the existing backlog list.

## Actor Summary Scope

The actor summary stays intentionally narrow.

Aggregation basis:

- `latest manual replay actor`

Output shape:

- top `5` actors
- each with:
  - `actor`
  - `count`

This means the summary reflects the distribution of each transaction's current/latest manual replay actor rather than a full historical actor histogram.

This slice intentionally excludes:

- full actor histogram
- actor grouping by outcome or family
- actor normalization pipeline
- history-wide actor accumulation

## Output Model

The current summary command contract is extended additively.

Existing fields:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`
- `top_latest_dead_letter_reasons`

New field:

- `top_latest_manual_replay_actors`

The expected item shape is:

- `actor`
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
  - aggregate top 5 latest manual replay actors
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

1. dispatch summary work
- top latest dispatch references

2. cockpit summary surface
- summary pane
- grouped overview in TUI

3. broader summary evolution
- grouped reason / actor families
- configurable top-N
- trend/time-window summaries
