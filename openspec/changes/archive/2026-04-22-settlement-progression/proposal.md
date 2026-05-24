# Proposal

## Why

`knowledge exchange v1` now has a runtime control-plane slice, but it still needs a dedicated transaction-level settlement progression layer so release outcomes can map into canonical settlement state before a full executor or dispute engine exists.

## What Changes

- add the first public settlement progression architecture page
- expose a receipts-backed `apply_settlement_progression` meta tool
- wire the settlement progression page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for project docs, docs references, and meta tools
- archive the completed change after sync

## Capabilities

### Modified Capabilities

- `project-docs`: publish the settlement progression architecture page from the public docs surface
- `docs-only`: keep the architecture landing page and knowledge-exchange track aligned with the landed slice
- `meta-tools`: record the receipts-backed settlement progression meta tool contract

## Impact

- Affected docs: `docs/architecture/settlement-progression.md`, `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`
- Affected site config: `zensical.toml`
- Affected code surface: `internal/app/tools_meta.go`, `internal/app/tools_parity_test.go`, `internal/app/tools_meta_settlementprogression_test.go`
- Affected OpenSpec specs: `openspec/specs/project-docs/spec.md`, `openspec/specs/docs-only/spec.md`, `openspec/specs/meta-tools/spec.md`
