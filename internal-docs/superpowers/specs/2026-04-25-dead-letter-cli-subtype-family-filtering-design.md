# Dead-Letter CLI Subtype / Family Filtering Design

## Purpose / Scope

This design upgrades the landed dead-letter CLI surface so operators can narrow the backlog by the latest retry lifecycle phase directly from the CLI.

This slice extends:

- `lango status dead-letters`

with:

- `--latest-status-subtype`
- `--latest-status-subtype-family`

This slice directly includes:

- list-command filter flags
- subtype validation
- latest-family validation
- existing `table` and `json` output behavior

This slice does not directly include:

- `any_match_family`
- actor/time filters
- reason/dispatch filters
- detail-command changes
- CLI recovery actions

## Command Surface Extension

The first slice only extends the list command:

- `lango status dead-letters`

Additional flags:

- `--latest-status-subtype`
- `--latest-status-subtype-family`

The detail command:

- `lango status dead-letter <transaction-receipt-id>`

remains unchanged in this slice.

## Flag Model

The flags use the existing CLI pattern of simple raw string flags:

- `--latest-status-subtype`
- `--latest-status-subtype-family`

This slice intentionally does not add:

- grouped filter objects
- multi-value flags
- comma-separated parsing

The first step is a narrow extension of the existing list-command surface.

## Validation Model

Allowed values match the currently landed read model exactly.

### `--latest-status-subtype`

Allowed values:

- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

### `--latest-status-subtype-family`

Allowed values:

- `retry`
- `manual-retry`
- `dead-letter`

The CLI validates these values explicitly instead of passing arbitrary strings through.

That keeps:

- operator error handling clearer
- help text more concrete
- CLI and read-model semantics aligned

## Data Source Reuse

This slice reuses the existing list read surface:

- `list_dead_lettered_post_adjudication_executions`

The CLI simply forwards the two new filter values through the current dead-letter list bridge.

No new backend path or direct store read is introduced.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/status`
  - add:
    - `--latest-status-subtype`
    - `--latest-status-subtype-family`
  - validate both flags
  - forward both values to the existing dead-letter list bridge

- extend status CLI tests
  - valid subtype/family values
  - invalid value rejection
  - filter forwarding

- update CLI docs/help text
- update public docs and OpenSpec

This slice does not change:

- backend contracts
- detail command behavior
- output modes

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer dead-letter CLI filters
- `any_match_family`
- actor/time
- reason/dispatch

2. CLI recovery actions
- replay from CLI

3. broader operator CLI
- grouped summaries
- background-task CLI views
