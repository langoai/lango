# Proposal

## Why

`settlement progression` now creates canonical `approved-for-settlement` state, but there is still no first execution slice that turns that state into direct settlement execution.

## What Changes

- add the first public architecture page for direct actual settlement execution
- expose a receipts-backed `execute_settlement` meta tool
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync

## Capabilities

### Modified Capabilities

- `project-docs`: publish the actual settlement execution architecture page
- `docs-only`: keep the architecture landing page and knowledge-exchange track aligned with the landed execution slice
- `meta-tools`: record the `execute_settlement` tool contract

## Impact

- Affected docs: `docs/architecture/actual-settlement-execution.md`, `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`
- Affected site config: `zensical.toml`
- Affected code surface: `internal/settlementexecution/*`, `internal/app/tools_meta.go`, `internal/app/tools_parity_test.go`, `internal/app/tools_meta_settlementexecution_test.go`
- Affected OpenSpec specs: `openspec/specs/project-docs/spec.md`, `openspec/specs/docs-only/spec.md`, `openspec/specs/meta-tools/spec.md`
