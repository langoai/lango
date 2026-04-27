# Dead-Letter CLI Surface Design

## Purpose / Scope

This design extends the landed dead-letter operator read surface beyond the cockpit/TUI into a first dedicated CLI surface for non-interactive operator workflows.

This slice adds:

- `lango status dead-letters`
- `lango status dead-letter <transaction-receipt-id>`

The target is the CLI status surface.

This slice directly includes:

- backlog list command
- per-transaction detail command
- default `table` output
- optional `json` output
- minimal list filters:
  - `query`
  - `adjudication`

This slice does not directly include:

- richer CLI filters
- replay or write actions
- background-task browsing commands
- `plain` output polish
- bulk operator workflows

## Command Surface

The first slice introduces two commands.

### `lango status dead-letters`

This command shows the current dead-letter backlog.

First-slice filters:

- `--query`
- `--adjudication`

This gives operators a fast way to scan backlog rows from the CLI without opening the cockpit.

### `lango status dead-letter <transaction-receipt-id>`

This command shows the current status detail for one selected dead-letter transaction.

The detail includes:

- canonical receipts-backed status
- latest retry / dead-letter summary
- optional latest background-task bridge

This keeps the CLI split clean:

- one command for backlog triage
- one command for per-transaction inspection

## Data Source Reuse

This slice does not add a CLI-specific backend service.

The CLI reuses the existing read surfaces:

- list command:
  - `list_dead_lettered_post_adjudication_executions`
- detail command:
  - `get_post_adjudication_execution_status`

That means the CLI:

- does not read receipts stores directly
- does not read the background manager directly
- does not fork the cockpit data path

The CLI stays aligned with the same canonical read model the cockpit already uses.

## Output Model

The first-slice output model is:

- default `table`
- optional `json`

### List command

`lango status dead-letters`

- `table` output is the human-scannable default
- `json` output exposes the structured backlog payload for machine use

### Detail command

`lango status dead-letter <transaction-receipt-id>`

- `table` output is a human-readable structured status summary
- `json` output exposes the structured status payload directly

This slice intentionally does not add:

- `plain` tuning
- custom column selection
- CSV export

## Filter Model

The first CLI slice intentionally exposes only the smallest useful subset:

- `--query`
- `--adjudication`

### `--query`

Free-text receipt-ID query over transaction and submission receipt identifiers.

### `--adjudication`

Allowed values:

- `release`
- `refund`

The cockpit already supports richer filters, but the first CLI slice intentionally stays smaller so the surface can land with minimal complexity.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/status`
  - add `dead-letters` list subcommand
  - add `dead-letter` detail subcommand
- reuse the existing app/tool bridge path
- add `--output`
  - default `table`
  - optional `json`
- add minimal list flags
  - `--query`
  - `--adjudication`
- update CLI docs/help text
- update public docs and OpenSpec

This slice does not add:

- new backend endpoints
- new direct store reads
- new canonical state

It is purely a new CLI wrapper over the existing dead-letter read model.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer dead-letter CLI filters
- subtype/family
- actor/time
- reason/dispatch

2. CLI recovery actions
- replay from CLI

3. broader operator CLI
- background-task status views
- grouped operator summaries
