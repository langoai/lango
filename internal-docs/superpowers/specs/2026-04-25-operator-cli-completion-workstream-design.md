# Operator CLI Completion Workstream Design

## Purpose / Scope

This workstream upgrades the landed dead-letter operator surface from a cockpit-first experience into a more complete non-interactive CLI operator surface.

This workstream bundles:

- dead-letter CLI retry action
- dead-letter CLI subtype / latest-family filters
- dead-letter CLI actor / time filters
- dead-letter CLI reason / dispatch filters

The target surface remains under the existing `status` command group:

- `lango status dead-letters ...`
- `lango status dead-letter <transaction-receipt-id>`
- `lango status dead-letter retry <transaction-receipt-id>`

This workstream directly includes:

- CLI parity with the current cockpit core filter set
- first CLI retry action

This workstream does not directly include:

- `any_match_family`
- polling / result follow-up UX
- bulk recovery
- dedicated background-task browsing commands
- broader operator summary dashboards

## CLI Surface

The workstream closes this CLI surface:

### `lango status dead-letters`

- dead-letter backlog list
- default `table`
- optional `json`
- richer operator filters

### `lango status dead-letter <transaction-receipt-id>`

- per-transaction detail
- canonical receipts-backed status
- optional latest background-task bridge

### `lango status dead-letter retry <transaction-receipt-id>`

- first CLI recovery action
- precheck + confirm + retry

The workstream intentionally keeps all of this under `status`, rather than creating a new top-level command group.

## Filter Scope

The list command filter scope for this workstream is:

- `--query`
- `--adjudication`
- `--latest-status-subtype`
- `--latest-status-subtype-family`
- `--manual-replay-actor`
- `--dead-lettered-after`
- `--dead-lettered-before`
- `--dead-letter-reason-query`
- `--latest-dispatch-reference`

This workstream intentionally does not include:

- `--any-match-family`
- multi-value filters
- advanced grouped filter syntax

The goal is CLI parity with the currently landed cockpit core filter set, not a maximal operator query model.

## Retry Action Scope

The retry action remains a thin CLI wrapper around the existing control plane.

Command:

- `lango status dead-letter retry <transaction-receipt-id>`

Flow:

1. existing detail read
2. `can_retry` precheck
3. default confirm prompt
4. `--yes` bypass
5. existing `retry_post_adjudication_execution` invocation

This workstream keeps:

- dead-letter evidence gate reuse
- adjudication gate reuse
- replay policy gate reuse

It intentionally does not add:

- polling
- action history
- bulk retry
- richer post-action workflow

## Execution / Parallelization Model

This workstream is intentionally larger than the earlier micro-slices.

Execution model:

- one spec
- one implementation plan
- two workers in parallel

### Worker A

Owns:

- `internal/cli/status/*`
- command wiring
- flag validation
- output rendering
- tests

### Worker B

Owns:

- `docs/cli/*`
- `docs/architecture/*`
- `README.md`
- `openspec/*`

This keeps the superpowers workflow intact while reducing overhead from repeated micro-artifact cycles.

## Implementation Shape

Recommended implementation shape:

- extend `internal/cli/status`
  - richer list filters
  - retained detail command
  - retry subcommand
  - shared dead-letter bridge consolidation
  - table/json output support

- update documentation set
  - `docs/cli/status.md`
  - `docs/cli/index.md`
  - `README.md`
  - `docs/architecture/dead-letter-browsing-status-observation.md`
  - `docs/architecture/p2p-knowledge-exchange-track.md`

- sync OpenSpec
  - main specs
  - one archive change for the workstream

The workstream is still additive. It does not redesign the existing dead-letter read model.

## Follow-On Inputs

Natural follow-on work after this workstream:

1. richer CLI filters
- `any_match_family`
- transaction-global filters
- count/sort expansion

2. richer CLI recovery UX
- polling
- structured result output
- richer failure detail

3. broader operator CLI
- grouped summaries
- background-task views
- bulk recovery
