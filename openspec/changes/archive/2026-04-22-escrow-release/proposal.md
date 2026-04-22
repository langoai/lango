# Proposal

## Why

The knowledge-exchange escrow path already lands `create + fund`, but it still lacks the first release execution slice that connects funded escrow to canonical settlement completion.

## What Changes

- add the first public architecture page for escrow release
- expose a receipts-backed `release_escrow_settlement` meta tool
- wire the page into the architecture landing page, track page, and docs navigation
- sync the OpenSpec requirements for docs and meta tools
- archive the completed change after sync

## Capabilities

### Modified Capabilities

- `project-docs`: publish the escrow release architecture page
- `docs-only`: keep the architecture landing page and knowledge-exchange track aligned with the landed escrow release slice
- `meta-tools`: record the `release_escrow_settlement` tool contract

## Impact

- Affected docs: `docs/architecture/escrow-release.md`, `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`
- Affected site config: `zensical.toml`
- Affected code surface: `internal/escrowrelease/*`, `internal/app/tools_meta.go`, `internal/app/tools_parity_test.go`, `internal/app/tools_meta_escrowrelease_test.go`
- Affected OpenSpec specs: `openspec/specs/project-docs/spec.md`, `openspec/specs/docs-only/spec.md`, `openspec/specs/meta-tools/spec.md`
