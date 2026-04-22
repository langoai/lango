# Proposal

## Why

`knowledge exchange v1` already has landed first slices for exportability, artifact release approval, upfront payment approval, direct prepay gating, dispute-ready receipts, and escrow recommendation execution. What is still missing is one canonical architecture page that explains how those slices compose into the first transaction-oriented runtime control plane without over-claiming a broader runtime implementation.

## What Changes

- Add `docs/architecture/knowledge-exchange-runtime.md` as the first runtime control-plane design slice for knowledge exchange.
- Wire the new page into the architecture landing page, the P2P knowledge exchange track page, and the docs site navigation.
- Record the documentation slice in OpenSpec with matching delta specs for `project-docs`, `docs-only`, `meta-tools`, and `mkdocs-documentation-site`.
- Sync the main specs so the repository docs, track ledger, docs-site navigation, and meta-tools contract stay aligned with the landed slice.
- Archive the completed OpenSpec change after the docs and specs are updated.

## Capabilities

### Modified Capabilities
- `project-docs`: Add an architecture landing-page entry for the knowledge exchange runtime design slice.
- `docs-only`: Add the runtime architecture page and the track-page linkage that closes out the first slice.
- `mkdocs-documentation-site`: Make the new runtime page navigable from the public architecture docs.
- `meta-tools`: Record the first runtime control-plane composition around the existing receipt-backed tools.

## Impact

- Affected docs: `docs/architecture/knowledge-exchange-runtime.md`, `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`
- Affected site config: `zensical.toml`
- Affected OpenSpec artifacts: `openspec/changes/knowledge-exchange-runtime-e2e/**`
- Affected specs: `openspec/specs/project-docs/spec.md`, `openspec/specs/docs-only/spec.md`, `openspec/specs/meta-tools/spec.md`
