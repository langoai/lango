## Why

The dead-letter operator surfaces had already landed as separate slices, but the remaining operator-facing gaps were still split across summary, filtering, and retry follow-up behavior. This consolidation workstream closes those remaining surface gaps in one pass so the docs and OpenSpec no longer describe them as pending work.

## What Changes

- document CLI `--any-match-family` parity on `lango status dead-letters`
- document grouped dispatch-family summaries in both CLI and cockpit
- document the landed top-N plus trend/time-window summary behavior
- document the landed retry follow-up UX in CLI and cockpit
- sync public docs and the main docs-only OpenSpec requirement set
- archive the completed workstream with docs-only change artifacts

## Impact

- operator-facing documentation now matches the shipped CLI and cockpit surfaces
- the knowledge-exchange track no longer lists already-landed operator-surface gaps as remaining work
- the remaining backlog is narrowed to broader taxonomy, history, and retry-policy follow-on work
