# Tasks

## 1. Read Model

- [x] 1.1 Extend `internal/postadjudicationstatus.DeadLetterListOptions` with reason and dispatch filters.
- [x] 1.2 Extend the existing list matcher with reason substring and dispatch exact matching.
- [x] 1.3 Keep the backlog response shape unchanged.

## 2. Meta Tools

- [x] 2.1 Extend `list_dead_lettered_post_adjudication_executions` with the new query params.
- [x] 2.2 Keep the existing page response contract unchanged.

## 3. Docs and OpenSpec

- [x] 3.1 Update `docs/architecture/dead-letter-browsing-status-observation.md`.
- [x] 3.2 Update `docs/architecture/p2p-knowledge-exchange-track.md`.
- [x] 3.3 Sync main specs under `openspec/specs/`.
- [x] 3.4 Archive the completed change under `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters`.
