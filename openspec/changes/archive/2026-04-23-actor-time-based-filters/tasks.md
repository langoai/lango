# Tasks

## 1. Read Model

- [x] 1.1 Extend `internal/postadjudicationstatus` types with actor/time filters and response fields.
- [x] 1.2 Extract latest manual replay actor and latest dead-letter timestamp from submission receipt trail evidence.
- [x] 1.3 Apply actor/time filters together with existing filters using `AND`.

## 2. Meta Tools

- [x] 2.1 Extend `list_dead_lettered_post_adjudication_executions` with actor/time query parameters.
- [x] 2.2 Preserve the existing page response shape while returning the richer entry fields.

## 3. Docs and OpenSpec

- [x] 3.1 Update `docs/architecture/dead-letter-browsing-status-observation.md`.
- [x] 3.2 Update `docs/architecture/index.md` and `docs/architecture/p2p-knowledge-exchange-track.md`.
- [x] 3.3 Sync main specs under `openspec/specs/`.
- [x] 3.4 Archive the completed change under `openspec/changes/archive/2026-04-23-actor-time-based-filters`.
