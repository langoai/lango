## Why

The dispute runtime moved beyond the earlier first-slice docs. Canonical keep-hold and re-escalation behavior now exists in receipts and recovery paths, settlement progression has deeper disagreement semantics, and dispute-linked tool receipts now expose dispute lifecycle state directly.

The public architecture docs and main `docs-only` OpenSpec requirements need to reflect that landed runtime instead of describing now-obsolete limits.

## What Changes

- update dispute / settlement / escrow architecture pages to match the landed canonical runtime
- document `hold-active` and `re-escalated` lifecycle behavior
- document richer settlement progression semantics, including dispute-ready re-entry and partial-hint preservation
- document dispute-linked tool receipts that now return `dispute_lifecycle_status`
- sync `openspec/specs/docs-only/spec.md`
- archive the completed docs-only workstream

## Impact

- public docs describe the current dispute runtime truthfully
- docs-only requirements align with the runtime that Worker A and Worker B landed
- follow-on work is framed around the remaining real gaps instead of already-completed slices
