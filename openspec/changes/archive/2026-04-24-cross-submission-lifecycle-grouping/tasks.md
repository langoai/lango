# Tasks

## 1. Read Model

- [x] 1.1 Extend `internal/postadjudicationstatus` with transaction-global row fields.
- [x] 1.2 Add transaction-global count/family filters.
- [x] 1.3 Aggregate over all submissions belonging to the current transaction.

## 2. Meta Tools

- [x] 2.1 Extend `list_dead_lettered_post_adjudication_executions` with transaction-global filter inputs.
- [x] 2.2 Keep the page response shape unchanged while exposing transaction-global row fields.

## 3. Docs and OpenSpec

- [x] 3.1 Update `docs/architecture/dead-letter-browsing-status-observation.md`.
- [x] 3.2 Update `docs/architecture/p2p-knowledge-exchange-track.md`.
- [x] 3.3 Sync main specs under `openspec/specs/`.
- [x] 3.4 Archive the completed change under `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping`.
