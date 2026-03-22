## Context

The current provenance implementation is split across three maturity levels:

1. **Checkpoints** are mostly implemented and persisted.
2. **Session tree** exists as in-memory domain logic plus an unused Ent schema.
3. **Attribution and bundles** exist only as placeholder types and CLI stubs.

Meanwhile, the repo already has the raw data sources needed to finish the feature:

- RunLedger journal positions for checkpoint anchoring
- `session.InMemoryChildStore` lifecycle hooks for fork/merge/discard signals
- `token_usage` persistence for per-session and per-agent cost accounting
- workspace gitbundle/log/diff/merge paths for commit/file evidence
- wallet public keys and DID derivation for peer-verifiable signatures

The implementation should complete Provenance by layering on top of those sources, keeping the journal and workspace repos authoritative for their own domains.

## Goals / Non-Goals

**Goals**
- Persist session tree nodes in Ent and surface them through the CLI
- Persist attribution rows with workspace/file/commit/session evidence
- Join token usage into provenance reports without duplicating token counters
- Export/import signed provenance bundles with `none|content|full` redaction
- Support P2P-verifiable provenance bundle transport using existing DID/public key identity

**Non-Goals**
- Rewriting RunLedger to become the source of workspace attribution
- Reconstructing or mutating existing session/run state during bundle import
- Adding unsigned fallback provenance transport
- Extending attribution capture into RunLedger worktree validation in this change

## Decisions

### D1. Session tree uses Ent as the runtime store

`session_provenance` already exists as schema and matches the required session tree fields. The change will add an Ent-backed `SessionTreeStore` and use it in the provenance module whenever `boot.DBClient` is available.

### D2. Attribution is first-class persisted data

Attribution will not be computed only from ad hoc joins at read time. A dedicated `provenance_attribution` table will persist coarse contribution evidence:

- session/run/workspace identity
- author type and author id
- file path / commit hash / step id when known
- source of capture (`workspace_merge`, `workspace_bundle_push`, `workspace_bundle_apply`, `session_fork`, etc.)
- line deltas

Token usage remains in `token_usage` and is joined into provenance views and reports.

### D3. Git-aware attribution is workspace-operation centric

The authoritative git evidence in the repo today is the workspace gitbundle layer. Attribution will be captured from:

- task branch merge operations
- bundle creation/push
- bundle application/fetch
- commit log and diff data

General sessions without workspace evidence still produce token-only author/session aggregates.

### D4. Runtime session lifecycle is wired from child-session hooks

The existing `SessionIsolation` and child-session abstractions are currently not used in runtime. This change wires a real child-session source into the app and connects its lifecycle hook to Provenance so fork/merge/discard is durably recorded.

### D5. Bundle signing uses wallet/DID identity, not security.CryptoProvider

`security.CryptoProvider` only exposes `Sign`, `Encrypt`, and `Decrypt`, but bundle verification must work across peers. The existing wallet + DID path already has:

- secp256k1 message signing
- compressed public key access
- DID derivation from public keys

Bundle signatures will therefore use wallet message signing and DID-based verification helpers.

### D6. Import is verify-and-store only

Bundle import validates the signer DID, signature, and redaction envelope, then stores provenance records in provenance-owned tables. It does not mutate existing sessions, runs, or workspace repos.

## Risks / Trade-offs

- **Runtime child-session integration risk**: the current multi-agent path does not yet consume child sessions. The change must add wiring without regressing orchestration behavior.
- **Git evidence ambiguity**: workspace repos may not always expose a stable DID→git author mapping. The implementation uses deterministic fallback rules: DID beats agent name, agent name beats raw git author string.
- **Bundle verification scope**: P2P-verifiable bundles require explicit signature verification helpers, but the change will constrain that scope to provenance bundle payload verification rather than a generic wallet trust framework.
