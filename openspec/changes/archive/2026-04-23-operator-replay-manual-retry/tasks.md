# Tasks

## 1. Replay Service

- [x] 1.1 Add `internal/postadjudicationreplay` request, result, and service types.
- [x] 1.2 Gate replay on dead-letter evidence and canonical adjudication.
- [x] 1.3 Reuse the existing background post-adjudication dispatch path.

## 2. Receipt Evidence

- [x] 2.1 Add `manual-retry-requested` evidence helper.
- [x] 2.2 Keep canonical adjudication unchanged and preserve prior dead-letter evidence.

## 3. Meta Tool

- [x] 3.1 Add `retry_post_adjudication_execution`.
- [x] 3.2 Return canonical adjudication snapshot plus dispatch receipt.

## 4. Docs and OpenSpec

- [x] 4.1 Add `docs/architecture/operator-replay-manual-retry.md`.
- [x] 4.2 Wire the page into `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`, and `zensical.toml`.
- [x] 4.3 Sync main specs under `openspec/specs/`.
- [x] 4.4 Archive the completed change under `openspec/changes/archive/2026-04-23-operator-replay-manual-retry`.
