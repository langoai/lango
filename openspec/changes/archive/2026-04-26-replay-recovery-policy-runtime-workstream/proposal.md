## Why

The replay and recovery runtime contract has now been normalized in code, but the public architecture pages and docs-only OpenSpec requirements still described several pieces of that runtime as pending or disconnected.

This archive captures the completed docs-only truth-alignment for:

- policy-driven post-adjudication execution defaults
- normalized retry / dead-letter policy shape
- replay / recovery substrate alignment

## What Changes

- update architecture pages for automatic, background, retry/dead-letter, operator replay, and replay-policy behavior
- update the P2P knowledge exchange track to stop listing already-landed runtime work as pending
- sync the main docs-only OpenSpec requirement set with the landed runtime contract
- archive the completed workstream artifacts

## Impact

- public docs now match the landed runtime behavior in `internal/background`, `internal/postadjudicationreplay`, `internal/receipts`, and `internal/app`
- the track backlog is narrowed to follow-on work that is still actually missing
- docs-only validation can reason about the normalized replay/recovery runtime as one coherent substrate
