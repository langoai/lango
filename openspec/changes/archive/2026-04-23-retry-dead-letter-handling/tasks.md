# Tasks

## 1. Background Retry Substrate

- [x] 1.1 Add retry metadata to background tasks.
- [x] 1.2 Add a post-adjudication-specific retry hook with exponential backoff.
- [x] 1.3 Stop automatic retry after the third retry and keep terminal failure.

## 2. Receipt Evidence

- [x] 2.1 Add append-only retry scheduled evidence.
- [x] 2.2 Add append-only dead-letter evidence.
- [x] 2.3 Keep canonical adjudication and settlement progression unchanged.

## 3. Docs and OpenSpec

- [x] 3.1 Add `docs/architecture/retry-dead-letter-handling.md`.
- [x] 3.2 Wire the page into `docs/architecture/index.md`, `docs/architecture/p2p-knowledge-exchange-track.md`, and `zensical.toml`.
- [x] 3.3 Sync main specs under `openspec/specs/`.
- [x] 3.4 Archive the completed change under `openspec/changes/archive/2026-04-23-retry-dead-letter-handling`.
