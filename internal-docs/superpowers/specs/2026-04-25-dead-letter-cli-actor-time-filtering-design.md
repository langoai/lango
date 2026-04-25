# Dead-Letter CLI Actor / Time Filtering Design

## Purpose / Scope

This design upgrades the landed dead-letter CLI surface so operators can narrow the backlog by the latest manual replay actor and latest dead-letter time window directly from the CLI.

This slice extends:

- `lango status dead-letters`

with:

- `--manual-replay-actor`
- `--dead-lettered-after`
- `--dead-lettered-before`

This slice directly includes:

- list-command filter flags
- actor text filter
- RFC3339 validation for time flags
- existing `table` / `json` output behavior

This slice does not directly include:

- detail-command changes
- retry-command changes
- reason/dispatch filters
- `any_match_family`
- after/before ordering validation

## Command Surface Extension

The slice extends only:

- `lango status dead-letters`

Added flags:

- `--manual-replay-actor`
- `--dead-lettered-after`
- `--dead-lettered-before`

The detail and retry commands remain unchanged in this slice.

## Flag Model

The flags follow the existing CLI pattern:

- `--manual-replay-actor`
  - free text string
- `--dead-lettered-after`
  - RFC3339 string
- `--dead-lettered-before`
  - RFC3339 string

This slice intentionally does not add:

- date-only shorthand
- unix timestamps
- grouped filter syntax

The goal is a narrow extension that mirrors the existing dead-letter read model directly.

## Validation Model

Validation in this slice is:

- `--manual-replay-actor`
  - no extra validation beyond normal string handling
- `--dead-lettered-after`
  - RFC3339 parse validation
- `--dead-lettered-before`
  - RFC3339 parse validation

This slice intentionally does not add:

- ordering validation between after/before
- partial date parsing
- timezone coercion helpers

The first step is early rejection of malformed time input.

## Data Source Reuse

This slice reuses the existing list surface:

- `list_dead_lettered_post_adjudication_executions`

The CLI forwards the three new values through the current dead-letter list bridge.

No new backend path or direct store read is introduced.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/status`
  - add:
    - `--manual-replay-actor`
    - `--dead-lettered-after`
    - `--dead-lettered-before`
  - add RFC3339 validation for both time flags
  - forward all three values into the existing dead-letter list bridge

- extend status CLI tests
  - valid actor/time forwarding
  - invalid time rejection

- update CLI docs/help text
- update public docs and OpenSpec

This slice does not change:

- backend contracts
- detail command behavior
- retry command behavior

## Follow-On Inputs

Natural follow-on work after this slice:

1. remaining CLI filters
- reason/dispatch
- `any_match_family`

2. richer CLI recovery UX
- polling
- richer failure detail

3. broader operator CLI
- grouped summaries
- background-task views
