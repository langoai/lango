# Tasks

## 1. Service

- [x] 1.1 Add `internal/escrowadjudication` request, result, and service types.
- [x] 1.2 Gate adjudication on funded escrow, dispute-ready settlement progression, current submission, and recorded hold evidence.
- [x] 1.3 Record release-vs-refund outcomes without mutating settlement progression or escrow execution state.

## 2. Receipt Evidence

- [x] 2.1 Extend `internal/receipts` with escrow adjudication state and request types.
- [x] 2.2 Add store helpers to apply adjudication decisions and append adjudication failure evidence.
- [x] 2.3 Verify receipt tests keep settlement progression and escrow execution state unchanged.

## 3. Meta Tool

- [x] 3.1 Add `adjudicate_escrow_dispute` to the receipt-backed meta tool surface.
- [x] 3.2 Add focused meta-tool tests for availability, canonical success, and failure behavior.

## 4. Docs and OpenSpec

- [x] 4.1 Add `docs/architecture/release-vs-refund-adjudication.md`.
- [x] 4.2 Wire the page into `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`, and `zensical.toml`.
- [x] 4.3 Sync main specs under `openspec/specs/`.
- [x] 4.4 Archive the completed change under `openspec/changes/archive/2026-04-23-release-vs-refund-adjudication`.
