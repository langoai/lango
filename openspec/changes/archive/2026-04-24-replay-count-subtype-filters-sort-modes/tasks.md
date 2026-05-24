# Tasks

## 1. Read Model

- [x] 1.1 Extend `internal/postadjudicationstatus` with subtype/count/sort query inputs.
- [x] 1.2 Derive `manual_retry_count`, `latest_manual_replay_at`, and `latest_status_subtype` from submission receipt trail evidence.
- [x] 1.3 Add bounded sort modes while keeping the list surface read-only.

## 2. Meta Tools

- [x] 2.1 Extend `list_dead_lettered_post_adjudication_executions` with subtype/count/sort query params.
- [x] 2.2 Keep the existing page response shape while returning richer row fields.

## 3. Docs and OpenSpec

- [x] 3.1 Update `docs/architecture/dead-letter-browsing-status-observation.md`.
- [x] 3.2 Update `docs/architecture/p2p-knowledge-exchange-track.md`.
- [x] 3.3 Sync main specs under `openspec/specs/`.
- [x] 3.4 Archive the completed change under `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes`.
