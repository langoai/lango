# Dead-Letter CLI Reason / Dispatch Filtering Design

## Purpose / Scope

This design upgrades the landed dead-letter CLI surface so operators can narrow the backlog by latest dead-letter reason and latest dispatch reference directly from the CLI.

This slice extends:

- `lango status dead-letters`

with:

- `--dead-letter-reason-query`
- `--latest-dispatch-reference`

This slice directly includes:

- list-command filter flags
- reason string filter
- dispatch string filter
- existing `table` / `json` output behavior

This slice does not directly include:

- detail-command changes
- retry-command changes
- `any_match_family`
- dispatch format validation
- reason minimum-length enforcement

## Command Surface Extension

The slice extends only:

- `lango status dead-letters`

Added flags:

- `--dead-letter-reason-query`
- `--latest-dispatch-reference`

The detail and retry commands remain unchanged in this slice.

## Flag Model

The flags follow the existing CLI pattern:

- `--dead-letter-reason-query`
  - free text string
- `--latest-dispatch-reference`
  - free text string

This slice intentionally does not add:

- grouped filter syntax
- pickers
- dispatch-format helpers

The first step is a narrow extension that mirrors the existing dead-letter read model directly.

## Validation Model

Validation is intentionally light in this slice.

- `--dead-letter-reason-query`
  - no extra validation
- `--latest-dispatch-reference`
  - no extra validation

Both values are passed through as strings.

This slice intentionally does not add:

- dispatch-reference pattern validation
- reason minimum-length rules
- autocomplete or suggestions

## Data Source Reuse

This slice reuses the existing list surface:

- `list_dead_lettered_post_adjudication_executions`

The CLI forwards both values through the current dead-letter list bridge.

No new backend path or direct store read is introduced.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/status`
  - add:
    - `--dead-letter-reason-query`
    - `--latest-dispatch-reference`
  - forward both values into the existing dead-letter list bridge

- extend status CLI tests
  - valid reason/dispatch forwarding
  - no extra validation behavior

- update CLI docs/help text
- update public docs and OpenSpec

This slice does not change:

- backend contracts
- detail command behavior
- retry command behavior

## Follow-On Inputs

Natural follow-on work after this slice:

1. remaining CLI filters
- `any_match_family`

2. richer CLI recovery UX
- polling
- richer failure detail

3. broader operator CLI
- grouped summaries
- background-task views
