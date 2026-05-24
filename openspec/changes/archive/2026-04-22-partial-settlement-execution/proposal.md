# Proposal

## Why

`actual settlement execution` now covers the first full direct settlement path, but the runtime still lacks a first partial execution slice for transactions that are approved for settlement yet should only execute a canonical partial amount.

## What Changes

- add the first public architecture page for direct partial settlement execution
- expose a receipts-backed `execute_partial_settlement` meta tool
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync

## Capabilities

### Modified Capabilities

- `project-docs`: publish the partial settlement execution architecture page
- `docs-only`: keep the architecture landing page and knowledge-exchange track aligned with the landed partial slice
- `meta-tools`: record the `execute_partial_settlement` tool contract

## Impact

- Affected docs: `docs/architecture/partial-settlement-execution.md`, `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`
- Affected site config: `zensical.toml`
- Affected code surface: `internal/partialsettlementexecution/*`, `internal/app/tools_meta.go`, `internal/app/tools_parity_test.go`, `internal/app/tools_meta_partialsettlementexecution_test.go`
- Affected OpenSpec specs: `openspec/specs/project-docs/spec.md`, `openspec/specs/docs-only/spec.md`, `openspec/specs/meta-tools/spec.md`
