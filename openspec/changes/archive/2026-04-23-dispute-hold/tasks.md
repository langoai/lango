# Tasks

## 1. Service

- [x] 1.1 Add `internal/disputehold` request, result, and service types.
- [x] 1.2 Gate execution on funded escrow, dispute-ready settlement progression, current submission, and escrow reference.
- [x] 1.3 Record hold success and failure shapes while leaving canonical transaction state unchanged.

## 2. Receipt Evidence

- [x] 2.1 Extend `internal/receipts` with dispute hold evidence request types.
- [x] 2.2 Add store helpers to append dispute hold success and failure evidence.
- [x] 2.3 Verify receipt tests keep escrow and settlement progression state unchanged.

## 3. Meta Tool

- [x] 3.1 Add `hold_escrow_for_dispute` to runtime-aware meta tool assembly.
- [x] 3.2 Add focused meta-tool tests for availability, canonical success, and failure behavior.

## 4. Docs and OpenSpec

- [x] 4.1 Add `docs/architecture/dispute-hold.md`.
- [x] 4.2 Wire the page into `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`, and `zensical.toml`.
- [x] 4.3 Sync main specs under `openspec/specs/`.
- [x] 4.4 Archive the completed change under `openspec/changes/archive/2026-04-23-dispute-hold`.
