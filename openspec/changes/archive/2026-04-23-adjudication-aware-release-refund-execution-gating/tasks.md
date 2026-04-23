# Tasks

## 1. Receipt and Adjudication Logic

- [x] 1.1 Update escrow adjudication to atomically record branch decision and settlement progression transition.
- [x] 1.2 Preserve append-only adjudication evidence and failure evidence.

## 2. Execution Gates

- [x] 2.1 Strengthen escrow release to require `escrow_adjudication = release`.
- [x] 2.2 Strengthen escrow refund to require `escrow_adjudication = refund`.
- [x] 2.3 Deny execution when opposite-branch evidence already exists.

## 3. Docs and OpenSpec

- [x] 3.1 Add `docs/architecture/adjudication-aware-release-refund-execution-gating.md`.
- [x] 3.2 Wire the page into `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`, and `zensical.toml`.
- [x] 3.3 Sync main specs under `openspec/specs/`.
- [x] 3.4 Archive the completed change under `openspec/changes/archive/2026-04-23-adjudication-aware-release-refund-execution-gating`.
