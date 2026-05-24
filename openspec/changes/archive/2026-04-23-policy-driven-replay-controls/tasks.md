# Tasks

## 1. Replay-Service Policy Gate

- [x] 1.1 Add actor resolution to replay.
- [x] 1.2 Add fail-closed deny reasons for unresolved actor and replay-not-allowed.
- [x] 1.3 Keep existing dead-letter and canonical adjudication gates unchanged.

## 2. Config-Backed Policy

- [x] 2.1 Add replay allowlist config fields.
- [x] 2.2 Pass config-backed replay policy into the replay service.

## 3. Docs and OpenSpec

- [x] 3.1 Add `docs/architecture/policy-driven-replay-controls.md`.
- [x] 3.2 Wire the page into `docs/architecture/index.md`, `docs/architecture/operator-replay-manual-retry.md`, `docs/architecture/p2p-knowledge-exchange-track.md`, and `zensical.toml`.
- [x] 3.3 Sync main specs under `openspec/specs/`.
- [x] 3.4 Archive the completed change under `openspec/changes/archive/2026-04-23-policy-driven-replay-controls`.
