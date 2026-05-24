# Proposal

## Why

The dead-letter backlog already supports actor/time filtering, but operators still need to narrow the list by the human-readable dead-letter reason and by the dispatch reference associated with the latest retry path.

## What Changes

- extend the post-adjudication status list filters with:
  - `dead_letter_reason_query`
  - `latest_dispatch_reference`
- keep the list response shape unchanged
- surface the new query params through the existing read-only backlog tool
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
